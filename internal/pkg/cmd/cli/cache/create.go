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

package cache

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/sugarkube/sugarkube/internal/pkg/cacher"
	"github.com/sugarkube/sugarkube/internal/pkg/cmd/cli/kapps"
	"github.com/sugarkube/sugarkube/internal/pkg/constants"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/stack"
	"github.com/sugarkube/sugarkube/internal/pkg/structs"
	"io"
	"path/filepath"
)

type createCmd struct {
	out             io.Writer
	dryRun          bool
	stackName       string
	stackFile       string
	provider        string
	provisioner     string
	profile         string
	account         string
	cluster         string
	region          string
	cacheDir        string
	renderTemplates bool
}

func newCreateCmd(out io.Writer) *cobra.Command {
	c := &createCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "create [flags] [stack-file] [stack-name] [cache-dir]",
		Short: fmt.Sprintf("Create kapp caches"),
		Long: `Create/update a local kapps cache for a given manifest(s), and renders any 
templates defined by kapps.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 3 {
				return errors.New("some required arguments are missing")
			} else if len(args) > 3 {
				return errors.New("too many arguments supplied")
			}
			c.stackFile = args[0]
			c.stackName = args[1]
			c.cacheDir = args[2]
			return c.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&c.dryRun, "dry-run", "n", false, "show what would happen but don't create a cluster")
	f.BoolVarP(&c.renderTemplates, "template", "t", false, "render templates for kapps ignoring any errors")
	f.StringVar(&c.provider, "provider", "", "name of provider, e.g. aws, local, etc.")
	f.StringVar(&c.provisioner, "provisioner", "", "name of provisioner, e.g. kops, minikube, etc.")
	f.StringVar(&c.profile, "profile", "", "launch profile, e.g. dev, test, prod, etc.")
	f.StringVarP(&c.cluster, "cluster", "c", "", "name of cluster to launch, e.g. dev1, dev2, etc.")
	f.StringVarP(&c.account, "account", "a", "", "string identifier for the account to launch in (for providers that support it)")
	f.StringVarP(&c.region, "region", "r", "", "name of region (for providers that support it)")

	return cmd
}

func (c *createCmd) run() error {

	log.Logger.Debugf("Got CLI args: %#v", c)

	// CLI args override configured args, so merge them in
	cliStackConfig := &structs.StackFile{
		Provider:    c.provider,
		Provisioner: c.provisioner,
		Profile:     c.profile,
		Cluster:     c.cluster,
		Region:      c.region,
		Account:     c.account,
	}

	stackObj, err := stack.BuildStack(c.stackName, c.stackFile, cliStackConfig, c.out)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Logger.Debugf("Loaded %d manifest(s)", len(stackObj.GetConfig().Manifests()))

	// todo - why is this here? why don't we always validate manifests?
	for _, manifest := range stackObj.GetConfig().Manifests() {
		err = stack.ValidateManifest(manifest)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	log.Logger.Debugf("Manifests validated.")

	absRootCacheDir, err := filepath.Abs(c.cacheDir)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Logger.Debugf("Caching manifests into %s...", absRootCacheDir)

	// don't use the abs cache path here to keep the output simpler
	_, err = fmt.Fprintf(c.out, "Caching kapps into '%s'...\n", c.cacheDir)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, manifest := range stackObj.GetConfig().Manifests() {
		err := cacher.CacheManifest(manifest, absRootCacheDir, c.dryRun)
		if err != nil {
			return errors.WithStack(err)
		}

		// reload each installable now its been cached so we can render templates
		for _, installableObj := range manifest.Installables() {
			err := installableObj.LoadConfigFile(absRootCacheDir)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	_, err = fmt.Fprintln(c.out, "Kapps successfully cached")
	if err != nil {
		return errors.WithStack(err)
	}

	log.Logger.Infof("Manifests cached to: %s", absRootCacheDir)

	if c.renderTemplates {
		_, err = fmt.Fprintln(c.out, "Rendering templates for kapps...")
		if err != nil {
			return errors.WithStack(err)
		}

		// create a DAG to template all the kapps
		dagObj, err := kapps.BuildDagForSelected(stackObj, c.cacheDir, []string{}, []string{},
			false, "", c.out)
		if err != nil {
			return errors.WithStack(err)
		}

		err = dagObj.Execute(constants.DagActionTemplate, stackObj, false, true, true,
			true, true, c.dryRun)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = fmt.Fprintln(c.out, "Templates successfully rendered")
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		_, err = fmt.Fprintln(c.out, "Skipping rendering templates for kapps")
		if err != nil {
			return errors.WithStack(err)
		}
	}

	_, err = fmt.Fprintf(c.out, "Kapps successfully cached into '%s'\n", absRootCacheDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
