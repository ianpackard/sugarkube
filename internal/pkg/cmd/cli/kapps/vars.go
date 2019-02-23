package kapps

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/sugarkube/sugarkube/internal/pkg/cmd/cli/utils"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/provider"
	"io"
)

type varsConfig struct {
	out         io.Writer
	cacheDir    string
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
	includeKapps []string
	excludeKapps []string
}

func newVarsCmd(out io.Writer) *cobra.Command {
	c := &varsConfig{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "vars",
		Short: fmt.Sprintf("Display all variables available for a kapp"),
		Long: `Merges variables from all sources and displays them. If a kapp is given, variables available for that 
specific kapp will be displayed. If not, all generally avaialble variables for the stack will be shown.`,
		RunE: c.run,
	}

	f := cmd.Flags()
	f.StringVarP(&c.stackName, "stack-name", "n", "", "name of a stack to launch (required when passing --stack-config)")
	f.StringVarP(&c.stackFile, "stack-config", "s", "", "path to file defining stacks by name")
	f.StringVarP(&c.cacheDir, "cache-dir", "d", "", "kapp cache directory")
	f.StringVar(&c.provider, "provider", "", "name of provider, e.g. aws, local, etc.")
	f.StringVar(&c.provisioner, "provisioner", "", "name of provisioner, e.g. kops, minikube, etc.")
	f.StringVar(&c.profile, "profile", "", "launch profile, e.g. dev, test, prod, etc.")
	f.StringVarP(&c.cluster, "cluster", "c", "", "name of cluster to launch, e.g. dev1, dev2, etc.")
	f.StringVarP(&c.account, "account", "a", "", "string identifier for the account to launch in (for providers that support it)")
	f.StringVarP(&c.region, "region", "r", "", "name of region (for providers that support it)")
	f.StringArrayVarP(&c.includeKapps, "include", "i", []string{}, "only process specified kapps (can specify multiple, formatted manifest-id:kapp-id)")
	f.StringArrayVarP(&c.excludeKapps, "exclude", "x", []string{}, "exclude individual kapps (can specify multiple, formatted manifest-id:kapp-id)")
	// these are commented for now to keep things simple, but ultimately we should probably support taking these as CLI args
	//f.VarP(&c.kappVarsDirs, "dir", "f", "Paths to YAML directory to load kapp values from (can specify multiple)")
	//f.VarP(&c.manifests, "manifest", "m", "YAML manifest file to load (can specify multiple but will replace any configured in a stack)")
	return cmd
}

func (c *varsConfig) run(cmd *cobra.Command, args []string) error {

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

	candidateKapps := map[string]kapp.Kapp{}

	if len(c.includeKapps) > 0 {
		log.Logger.Debugf("Adding %d kapps to the candidate template set", len(c.includeKapps))
		candidateKapps, err = getKappsByFullyQualifiedId(c.includeKapps, stackConfig)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		log.Logger.Debugf("Adding all kapps to the candidate template set")

		log.Logger.Debugf("Stack config has %d manifests", len(stackConfig.AllManifests()))

		// select all kapps
		for _, manifest := range stackConfig.AllManifests() {
			log.Logger.Debugf("Manifest '%s' contains %d kapps", manifest.Id, len(manifest.Kapps))

			for _, manifestKapp := range manifest.Kapps {
				candidateKapps[manifestKapp.FullyQualifiedId()] = manifestKapp
			}
		}
	}

	log.Logger.Debugf("There are %d candidate kapps for templating (before applying exclusions)",
		len(candidateKapps))

	if len(c.excludeKapps) > 0 {
		// delete kapps
		excludedKapps, err := getKappsByFullyQualifiedId(c.excludeKapps, stackConfig)
		if err != nil {
			return errors.WithStack(err)
		}

		log.Logger.Debugf("Excluding %d kapps from the templating set", len(excludedKapps))

		for k := range excludedKapps {
			if _, ok := candidateKapps[k]; ok {
				delete(candidateKapps, k)
			}
		}
	}

	_, err = fmt.Fprintf(c.out, "Displaying variables for %d kapps\n", len(candidateKapps))
	if err != nil {
		return errors.WithStack(err)
	}

	providerVars := provider.GetVars(providerImpl)

	for _, kappObj := range candidateKapps {
		mergedKappVars, err := kapp.MergeVarsForKapp(&kappObj, stackConfig, providerVars,
			map[string]interface{}{})

		_, err = fmt.Fprintf(c.out, "Variables for kapp '%s': \n%s", kappObj.Id, mergedKappVars)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
