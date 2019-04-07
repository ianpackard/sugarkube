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

package interfaces

import (
	"github.com/sugarkube/sugarkube/internal/pkg/acquirer"
	"github.com/sugarkube/sugarkube/internal/pkg/structs"
)

// this encapsulates different package formats that sugarkube can install in
// a target stack
type IInstallable interface {
	Id() string
	FullyQualifiedId() string
	ManifestId() string
	State() string
	PostActions() []string
	GetDescriptor() structs.KappDescriptorWithMaps
	LoadConfigFile(cacheDir string) error
	SetTopLevelCacheDir(cacheDir string) error
	GetCacheDir() string
	Acquirers() (map[string]acquirer.Acquirer, error)
	TemplateDescriptor(templateVars map[string]interface{}) error
	GetCliArgs(installerName string, command string) []string
	GetEnvVars() map[string]interface{}
	Vars(stack IStack) (map[string]interface{}, error)
	AddDescriptor(config structs.KappDescriptorWithMaps, prepend bool) error
	RenderTemplates(templateVars map[string]interface{}, stackConfig IStackConfig,
		requireTemplateDestDirs bool, dryRun bool) ([]string, error)
	GetOutputs(dryRun bool) (map[string]interface{}, error)
}
