package installer

import (
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/provider"
)

// This is a generic way of inspecting kapps to see what they contain and what
// env vars/CLI parameters should be passed to their installers

// todo - move into a default config file, but allow them to be overridden/
// addtional interfaces defined.
//var parameteriserConfig = `
//kapp_interfaces:	# different things a kapp might contain. A kapp may 'implement'
//					# multiple interfaces (e.g. contain both a helm chart and terraform configs
//  helm_chart:
//    heuristics: 			# inspections we can carry out on a kapp to see what it contains
//    - file:
//        pattern: Chart.yaml		# regex to search for under the kapp root dir
//    params:
//      env:
//      - name: KUBE_CONTEXT
//        value:
//          type: vars_lookup
//          path: provider
//          key: kube_context
//      - name: NAMESPACE		# default value. Allow overriding it in the installer config.
//        value: 				# think of how to configure that here...
//          type: obj_field
//          path: kapp
//          key: Id
//      - name: RELEASE
//        value:
//          type: obj_field
//          path: kapp
//          key: Id
//      cliArgs:
//      - name: helm-opts
//        components:
//        - key: -f
//          value
//            pattern: values-(\w+).yaml
//
//  k8s_resource:             # a naked k8s resource. No heuristics. Expect to find
//    params:					# it listed in 'sugarkube.yaml'
//      env:
//      - name: KUBE_CONTEXT
//        value:
//          type: vars_lookup
//          path: provider
//          key: kube_context
//
//  terraform:
//    heuristics:
//    - file:
//        pattern: terraform.*
//        type: dir
//    params:
//      cliArgs:
//      - name: tf-opts
//        components:			# by default collapse multiple values into a
//        - key: -var-file		# single CLI arg
//          value:
//            pattern: vars/(\w+).tfvars
//`

const IMPLEMENTS_HELM = "helm"
const IMPLEMENTS_TERRAFORM = "terraform"
const IMPLEMENTS_K8S = "k8s"

type Parameteriser struct {
	Name    string
	kappObj *kapp.Kapp
}

const KUBE_CONTEXT_KEY = "kube_context"

// Return a map of env vars that should be passed to the kapp by the installer
func (i *Parameteriser) GetEnvVars(vars provider.Values) map[string]string {
	envVars := make(map[string]string)

	if i.Name == IMPLEMENTS_HELM {
		envVars["NAMESPACE"] = i.kappObj.Id
		envVars["RELEASE"] = i.kappObj.Id
		envVars["KUBE_CONTEXT"] = vars[KUBE_CONTEXT_KEY].(string)
	}

	return envVars
}

// Returns a list of args that the installer should pass to the kapp
func (i *Parameteriser) GetCliArgs(validPatternMatches []string) ([]string, error) {
	pattern := ""
	argName := ""

	if i.Name == IMPLEMENTS_HELM {
		pattern = "values-(?P<Var>\\w*).yaml"
		argName = "helm-opts"
	}

	if i.Name == IMPLEMENTS_TERRAFORM {
		pattern = "terraform.*"
		argName = "tf-opts"
	}

	if pattern == "" {
		return []string{}, nil
	}

	matches, err := findFilesByPattern(i.kappObj.RootDir, pattern, true)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cliArgs := make([]string, 0)

	// make sure the matching group in each match is in the valid pattern matches list
	for _, match := range matches {
		matchingGroups := getRexExpCapturingGroups(pattern, match)

		// don't punish yourself by saying the words "functional programming"...
		for _, v := range matchingGroups {
			for _, valid := range validPatternMatches {
				if v == valid {
					cliArgs = append(cliArgs, match)
				}
			}
		}
	}

	return cliArgs, nil
}

// Examines a kapp to find out what it contains, and therefore what env vars/
// CLI args need passing to it by an Installer.
func identifyKappInterfaces(kappObj *kapp.Kapp) ([]Parameteriser, error) {
	// todo - parse the above config and test the kapp using it.
	// todo - also look in the kapp's sugarkube.yaml file if it exists

	parameterisers := make([]Parameteriser, 0)

	// todo - remove IMPLEMENTS_K8S from this. It's a temporary kludge until we
	// can get it from the kapp's sugarkube.yaml file
	parameterisers = append(parameterisers, Parameteriser{
		Name: IMPLEMENTS_K8S, kappObj: kappObj})

	// todo - remove this kludge to find out whether the kapp contains a helm chart.
	chartPaths, err := findFilesByPattern(kappObj.RootDir, "Chart.yaml", true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(chartPaths) > 0 {
		parameterisers = append(parameterisers, Parameteriser{
			Name: IMPLEMENTS_HELM, kappObj: kappObj})
	}

	// todo - remove this kludge to find out whether the kapp contains terraform configs
	terraformPaths, err := findFilesByPattern(kappObj.RootDir, "terraform", true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(terraformPaths) > 0 {
		parameterisers = append(parameterisers, Parameteriser{
			Name: IMPLEMENTS_TERRAFORM, kappObj: kappObj})
	}

	return parameterisers, nil
}
