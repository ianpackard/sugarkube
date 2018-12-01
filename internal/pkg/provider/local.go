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

package provider

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/constants"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"os"
	"path/filepath"
)

type LocalProvider struct {
	stackConfigVars map[string]interface{}
}

const LOCAL_PROVIDER_NAME = "local"

// Associate provider variables with the provider
func (p *LocalProvider) setVars(values map[string]interface{}) {
	p.stackConfigVars = values
}

// Returns the variables loaded by the Provider
func (p *LocalProvider) getVars() map[string]interface{} {
	return p.stackConfigVars
}

// Return vars loaded from configs that should be passed on to all kapps by
// installers so kapps can be installed into this provider
func (p *LocalProvider) getInstallerVars() map[string]interface{} {
	return map[string]interface{}{}
}

// Returns directories to look for values files in specific to this provider
func (p *LocalProvider) varsDirs(sc *kapp.StackConfig) ([]string, error) {

	paths := make([]string, 0)

	prefix := sc.Dir()

	for _, path := range sc.ProviderVarsDirs {
		// prepend the directory of the stack config file if the path is relative
		if !filepath.IsAbs(path) {
			path = filepath.Join(prefix, path)
			log.Logger.Debugf("Prepended dir of stack config to relative path. New path %s", path)
		}

		profileDir := filepath.Join(path, LOCAL_PROVIDER_NAME, constants.PROFILE_DIR, sc.Profile)
		clusterDir := filepath.Join(path, LOCAL_PROVIDER_NAME, constants.PROFILE_DIR, sc.Profile, constants.CLUSTER_DIR, sc.Cluster)

		if err := abortIfNotDir(profileDir,
			fmt.Sprintf("No profile directory found at %s", profileDir)); err != nil {
			return nil, err
		}

		if err := abortIfNotDir(clusterDir,
			fmt.Sprintf("No cluster directory found at %s", clusterDir)); err != nil {
			return nil, err
		}

		paths = append(paths, filepath.Join(path))
		paths = append(paths, filepath.Join(path, LOCAL_PROVIDER_NAME))
		paths = append(paths, filepath.Join(path, LOCAL_PROVIDER_NAME, constants.PROFILE_DIR))
		paths = append(paths, profileDir)
		paths = append(paths, filepath.Join(path, LOCAL_PROVIDER_NAME, constants.PROFILE_DIR, sc.Profile, constants.CLUSTER_DIR))
		paths = append(paths, clusterDir)
	}

	return paths, nil
}

// Returns an error if the given path doesn't exist or isn't a directory
func abortIfNotDir(path string, errorMessage string) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, errorMessage)
	}

	if !info.IsDir() {
		return errors.New(fmt.Sprintf("Path '%s' is not a directory", path))
	}

	return nil
}

// Returns the name of this provider
func (p *LocalProvider) getName() string {
	return LOCAL_PROVIDER_NAME
}
