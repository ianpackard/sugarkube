package kappsot

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
)

// Uses Helm to determine which kapps are already installed in a target cluster
type HelmKappSot struct {
	charts HelmOutput
}

// Wrapper around Helm output
type HelmOutput struct {
	Next     string
	Releases []HelmRelease
}

// struct returned by `helm list --output yaml`
type HelmRelease struct {
	AppVersion string
	Chart      string
	Name       string
	Namespace  string
	Revision   int
	Status     string
	Updated    string
}

// Refreshes the list of Helm charts
func (s HelmKappSot) refresh() error {
	var stdout bytes.Buffer
	// todo - add the --kube-context
	cmd := exec.Command("helm", "list", "--all", "--output", "yaml")
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "Error running 'helm list'")
	}

	// parse stdout
	output := HelmOutput{}
	err = yaml.Unmarshal(stdout.Bytes(), &output)
	if err != nil {
		return errors.Wrapf(err, "Error parsing 'Helm list' output: %s",
			stdout.String())
	}

	s.charts = output

	return nil
}

// Returns whether a helm chart is already successfully installed on the cluster
func (s HelmKappSot) isInstalled(name string, version string) (bool, error) {

	// todo - make sure we refresh this for each manifest to catch the same
	// chart being installed by different manifests accidentally.
	if s.charts.Releases == nil {
		err := s.refresh()
		if err != nil {
			return false, errors.WithStack(err)
		}
	}

	chart := fmt.Sprintf("%s=%s", name, version)

	for _, release := range s.charts.Releases {
		if release.Chart == chart {
			if release.Status == "DEPLOYED" {
				log.Infof("Chart '%s' is already installed", chart)
				return true, nil
			}

			if release.Status == "FAILED" {
				log.Infof("The previous release of chart '%s' failed", chart)
				return false, nil
			}

			if release.Status == "DELETED" {
				log.Infof("Chart '%s' was installed but was deleted", chart)
				return false, nil
			}
		}
	}

	log.Debugf("Chart '%s' isn't installed", chart)

	return false, nil
}