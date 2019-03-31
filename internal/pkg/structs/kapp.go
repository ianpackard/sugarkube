/*
 * Copyright 2019 The Sugarkube Authors
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

package structs

// Structs to load a kapp's sugarkube.yaml file

// A fragment of configuration for a program or kapp. It can be loaded either
// from a kapp's sugarkube.yaml file or the global sugarkube config file. It
// allows default env vars and arguments to be configured in one place and reused.
type ProgramConfig struct {
	EnvVars map[string]interface{} `yaml:"envVars"`
	Version string
	Args    map[string]map[string][]map[string]string
}

type Template struct {
	Source    string
	Dest      string
	Sensitive bool // sensitive templates will be templated just-in-time then deleted immediately after
	// executing the kapp. This provides a way of passing secrets to kapps while keeping them off
	// disk as much as possible.
}

// Outputs generated by a kapp that should be parsed and added to the registry
type Output struct {
	Id        string
	Path      string
	Type      string
	Sensitive bool // sensitive outputs will be deleted after adding the data to the registry to try to prevent
	// secrets lingering on disk
}

type Source struct {
	Id      string
	Uri     string
	Options map[string]interface{} // we don't have explicit path/branch fields because this struct must be
	// generic enough for all acquirers, not be specific to git
	IncludeValues bool // todo - decide if this is needed and remove if not
}

// A struct for an actual sugarkube.yaml file
type KappConfig struct {
	State         string
	ProgramConfig `yaml:",inline"`
	Requires      []string
	PostActions   []string `yaml:"post_actions"`
	Templates     []Template
	Vars          map[string]interface{}
}

// KappDescriptors describe where to find a kapp plus some other data, but isn't the kapp itself.
// There are two types - one that has certain values declared as lists and one as maps where keys
// are that element's ID. The list version is more concise and is used in manifest files. The
// version with maps is used when overriding values (e.g. in stack files)

type KappDescriptorWithLists struct {
	Id         string
	KappConfig `yaml:",inline"`
	Sources    []Source
	Outputs    []Output
}

type KappDescriptorWithMaps struct {
	Id         string
	KappConfig `yaml:",inline"`
	Sources    map[string]Source // keys are object IDs so values for individual objects can be overridden
	Outputs    map[string]Output // keys are object IDs so values for individual objects can be overridden
}
