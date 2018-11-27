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

package kapps

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/sugarkube/sugarkube/internal/pkg/cmd/cli/utils"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/plan"
	"io"
)

type applyCmd struct {
	out         io.Writer
	diffPath    string
	cacheDir    string
	dryRun      bool
	approved    bool
	oneShot     bool
	force       bool
	stackName   string
	stackFile   string
	provider    string
	provisioner string
	//kappVarsDirs cmd.Files
	profile string
	account string
	cluster string
	region  string
	//manifests    cmd.Files
	// todo - add options to :
	// * filter the kapps to be processed (use strings like e.g. manifest:kapp-id to refer to kapps)
	// * exclude manifests / kapps from being processed
}

func newApplyCmd(out io.Writer) *cobra.Command {
	c := &applyCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "apply [cache-dir]",
		Short: fmt.Sprintf("Install/destroy kapps into a cluster"),
		Long:  `Apply cached kapps to a target cluster according to manifests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("the path to the kapp cache dir is required")
			}
			c.cacheDir = args[0]
			return c.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&c.dryRun, "dry-run", false, "show what would happen but don't create a cluster")
	f.BoolVar(&c.approved, "approved", false, "actually apply a cluster diff to install/destroy kapps. If false, kapps "+
		"will be expected to plan their changes but not make any destrucive changes (e.g. should run 'terraform plan', etc. but not "+
		"apply it).")
	f.BoolVar(&c.oneShot, "one-shot", false, "apply a cluster diff in a single pass by invoking each kapp with "+
		"'APPROVED=false' then 'APPROVED=true' to install/destroy kapps in a single invocation of sugarkube")
	f.BoolVar(&c.force, "force", false, "don't require a cluster diff, just blindly install/destroy all the kapps "+
		"defined in a manifest(s)/stack config, even if they're already present/absent in the target cluster")
	f.StringVarP(&c.diffPath, "diff-path", "d", "", "Path to the cluster diff to apply. If not given, a "+
		"diff will be generated")
	f.StringVarP(&c.stackName, "stack-name", "n", "", "name of a stack to launch (required when passing --stack-config)")
	f.StringVarP(&c.stackFile, "stack-config", "s", "", "path to file defining stacks by name")
	f.StringVar(&c.provider, "provider", "", "name of provider, e.g. aws, local, etc.")
	f.StringVar(&c.provisioner, "provisioner", "", "name of provisioner, e.g. kops, minikube, etc.")
	f.StringVar(&c.profile, "profile", "", "launch profile, e.g. dev, test, prod, etc.")
	f.StringVarP(&c.cluster, "cluster", "c", "", "name of cluster to launch, e.g. dev1, dev2, etc.")
	f.StringVarP(&c.account, "account", "a", "", "string identifier for the account to launch in (for providers that support it)")
	f.StringVarP(&c.region, "region", "r", "", "name of region (for providers that support it)")
	// these are commented for now to keep things simple, but ultimately we should probably support taking these as CLI args
	//f.VarP(&c.kappVarsDirs, "dir", "f", "Paths to YAML directory to load kapp values from (can specify multiple)")
	//f.VarP(&c.manifests, "manifest", "m", "YAML manifest file to load (can specify multiple but will replace any configured in a stack)")
	return cmd
}

func (c *applyCmd) run() error {

	// CLI overrides - will be merged with any loaded from a stack config file
	cliStackConfig := &kapp.StackConfig{
		Provider:    c.provider,
		Provisioner: c.provisioner,
		Profile:     c.profile,
		Cluster:     c.cluster,
		Region:      c.region,
		Account:     c.account,
		//KappVarsDirs: c.kappVarsDirs,
		//Manifests:    cliManifests,
	}

	stackConfig, providerImpl, _, err := utils.ProcessCliArgs(c.stackName,
		c.stackFile, cliStackConfig, c.out)
	if err != nil {
		return errors.WithStack(err)
	}

	var actionPlan *plan.Plan

	if !c.force {
		panic("Cluster diffing not implemented. Pass --force")

		if c.diffPath != "" {
			// todo load a cluster diff from a file

			// todo - validate that the embedded stack config matches the target cluster.

			// in future we may want to be able to work entirely from a cluster
			// diff, in which case it'd really be a plan for us
			if len(stackConfig.Manifests) > 0 {
				// todo - validate that the cluster diff matches the manifests, e.g. that
				// the versions of kapps in the manifests match the versions in the cluster
				// diff
			}
		} else {
			// todo - create a cluster diff based on stackConfig.Manifests
		}

		// todo - diff the cache against the kapps in the cluster diff and abort if
		// it's out-of-sync (unless flags are set to ignore cache changes), e.g.:
		//cacheDiff, err := cacher.DiffKappCache(clusterDiff, c.cacheDir)
		//if err != nil {
		//	return errors.WithStack(err)
		//}
		//if len(diff) != 0 {
		//	return errors.New("Cache out-of-sync with manifests: %s", diff)
		//}

		// todo - create an action plan from the validated cluster diff
		//actionPlan, err := plan.FromDiff(clusterDiff)

	} else {
		fmt.Fprintln(c.out, "Planning operations on kapps")

		// force mode, so no need to perform validation. Just create a plan
		actionPlan, err = plan.Create(stackConfig, c.cacheDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if !c.oneShot {
		fmt.Fprintf(c.out, "Applying the plan with APPROVED=%#v...\n", c.approved)

		// run the plan either preparing or applying changes
		err := actionPlan.Run(c.approved, providerImpl, c.dryRun)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		fmt.Fprintln(c.out, "Applying the plan in a single pass")

		fmt.Fprintln(c.out, "First applying the plan with APPROVED=false for "+
			"kapps to plan their changes...")
		// one-shot mode, so prepare and apply the plan straight away
		err = actionPlan.Run(false, providerImpl, c.dryRun)
		if err != nil {
			return errors.WithStack(err)
		}

		fmt.Fprintln(c.out, "Now running with APPROVED=true to actually apply changes...")
		err = actionPlan.Run(true, providerImpl, c.dryRun)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	fmt.Fprintln(c.out, "Kapp change plan successfully applied")

	return nil
}