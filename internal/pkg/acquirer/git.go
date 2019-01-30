/*
 * Copyright 2018 The Sugarkube Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package acquirer

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type GitAcquirer struct {
	name          string
	uri           string
	branch        string
	path          string
	includeValues bool
}

// todo - make configurable, or use go-git
const GIT_PATH = "git"

const NAME = "name"
const URI = "uri"
const BRANCH = "branch"
const PATH = "path"
const INCLUDE_VALUES = "includeValues"

// Returns an instance. This allows us to build objects for testing instead of
// directly instantiating objects in the acquirer factory.
func NewGitAcquirer(name string, uri string, branch string, path string,
	includeValues string) GitAcquirer {
	if name == "" {
		name = filepath.Base(path)
	}

	// defaults to true
	includeValuesBool := true

	if strings.ToLower(includeValues) == "false" {
		includeValuesBool = false
	}

	return GitAcquirer{
		name:          name,
		uri:           uri,
		branch:        branch,
		path:          path,
		includeValues: includeValuesBool,
	}
}

// Generate an ID
func (a GitAcquirer) Id() (string, error) {
	// testing here simplifies testing but does mean invalid objects can be created...
	if strings.Count(a.uri, ":") != 1 {
		return "", errors.New(
			fmt.Sprintf("Unexpected git URI. Expected a single ':' "+
				"character in URI %s", a.uri))
	}

	// this doesn't contain the branch because we don't want to create complications
	// in case users create their own branches (e.g. if we've checked out into a
	// directory containing 'master' and they create a feature branch the dir name
	// will be misleading).
	orgRepo := strings.SplitAfter(a.uri, ":")
	hyphenatedOrg := strings.Replace(orgRepo[1], "/", "-", -1)
	hyphenatedOrg = strings.TrimSuffix(hyphenatedOrg, ".git")
	hyphenatedName := strings.Replace(a.name, "/", "-", -1)

	return strings.Join([]string{hyphenatedOrg, hyphenatedName}, "-"), nil
}

// return the name
func (a GitAcquirer) Name() string {
	return a.name
}

// return the path
func (a GitAcquirer) Path() string {
	return a.path
}

// return whether this source should be searched for values files
func (a GitAcquirer) IncludeValues() bool {
	return a.includeValues
}

// Acquires kapps via git and saves them to `dest`.
func (a GitAcquirer) acquire(dest string) error {

	var destExists bool

	if _, err := os.Stat(dest); err != nil {
		if os.IsNotExist(err) {
			log.Logger.Debugf("Destination directory '%s' doesn't exist... will create it", dest)
			destExists = false
		} else {
			return errors.WithStack(err)
		}
	} else {
		log.Logger.Debugf("Destination directory '%s' already exists... will update it", dest)
		destExists = true
	}

	if destExists {
		return a.update(dest)
	} else {
		return a.clone(dest)
	}
}

// Performs a sparse checkout for when the destination directory doesn't already exist
func (a GitAcquirer) clone(dest string) error {

	log.Logger.Infof("Cloning git source '%s' into '%s'", a.uri, dest)

	// create the dest dir if it doesn't exist
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.Wrapf(err, "Error creating directory '%s'", dest)
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	// git init
	err = utils.ExecCommand(GIT_PATH, []string{"init"}, map[string]string{},
		&stdoutBuf, &stderrBuf, dest, 5, false)
	if err != nil {
		return errors.WithStack(err)
	}

	// add origin
	err = utils.ExecCommand(GIT_PATH, []string{"remote", "add", "origin", a.uri},
		map[string]string{}, &stdoutBuf, &stderrBuf, dest, 5, false)
	if err != nil {
		return errors.WithStack(err)
	}

	// fetch
	err = utils.ExecCommand(GIT_PATH, []string{"fetch"}, map[string]string{},
		&stdoutBuf, &stderrBuf, dest, 60, false)
	if err != nil {
		return errors.WithStack(err)
	}

	// git configure sparse checkout
	err = utils.ExecCommand(GIT_PATH, []string{"config", "core.sparsecheckout", "true"},
		map[string]string{}, &stdoutBuf, &stderrBuf, dest, 90, false)
	if err != nil {
		return errors.WithStack(err)
	}

	err = utils.AppendToFile(filepath.Join(dest, ".git/info/sparse-checkout"),
		fmt.Sprintf("%s/*\n", strings.TrimSuffix(a.path, "/")))
	if err != nil {
		return errors.WithStack(err)
	}

	// git checkout
	err = utils.ExecCommand(GIT_PATH, []string{"checkout", a.branch},
		map[string]string{}, &stdoutBuf, &stderrBuf, dest, 90, false)
	if err != nil {
		return errors.WithStack(err)
	}

	// we could optionally verify tags with:
	// git tag -v a.branch 2>&1 >/dev/null | grep -E '{{ trusted_gpg_keys|join('|') }}'

	return nil
}

// Pulls a previously checked out source to update it
func (a GitAcquirer) update(dest string) error {

	var stdoutBuf, stderrBuf bytes.Buffer
	var err error

	// find out which branch is currently checked out
	err = utils.ExecCommand(GIT_PATH, []string{"branch", "--format", "%(refname:short)"},
		map[string]string{}, &stdoutBuf, &stderrBuf, dest, 2, false)

	log.Logger.Debugf("Stdout=%s", stdoutBuf.String())
	log.Logger.Debugf("Stderr=%s", stderrBuf.String())

	if err != nil {
		return errors.WithStack(err)
	}

	localBranch := strings.TrimSpace(stdoutBuf.String())

	if localBranch == a.branch {
		log.Logger.Debugf("Branch '%s' already checked out into local cache at '%s'. Will "+
			"update it...", localBranch, dest)
	} else {
		// todo - work out if there's anything we can do to help (a flag to force overwriting,
		// or a flag to just go ahead and switch? We could do a 'git status' and ignore the
		// different branches if there are no modified files, etc.)
		return errors.New(fmt.Sprintf("Error updating the cache. The path "+
			"at '%s' already contains the branch '%s', but we need to populate it with "+
			"the branch '%s'. Aborting to prevent losing work.",
			dest, localBranch, a.branch))
	}

	err = utils.ExecCommand(GIT_PATH, []string{"pull", "origin", a.branch},
		map[string]string{}, &stdoutBuf, &stderrBuf, dest, 90, false)

	log.Logger.Debugf("Stdout=%s", stdoutBuf.String())
	log.Logger.Debugf("Stderr=%s", stderrBuf.String())

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
