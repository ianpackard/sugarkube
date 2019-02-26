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

package provisioner

import (
	"github.com/stretchr/testify/assert"
	"github.com/sugarkube/sugarkube/internal/pkg/cmd/cli/utils"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"os"
	"testing"
)

func init() {
	log.ConfigureLogger("debug", false)
}

func TestNewNonExistentProvisioner(t *testing.T) {
	actual, err := NewProvisioner("bananas", &kapp.StackConfig{})
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestNewMinikubeProvisioner(t *testing.T) {
	actual, err := NewProvisioner(MINIKUBE_PROVISIONER_NAME, &kapp.StackConfig{})
	assert.Nil(t, err)
	assert.Equal(t, MinikubeProvisioner{}, actual)
}

func TestNewKopsProvisioner(t *testing.T) {
	stackConfig, err := utils.BuildStackConfig("kops", "../../testdata/stacks.yaml",
		&kapp.StackConfig{}, os.Stdout)
	assert.Nil(t, err)

	actual, err := NewProvisioner(KOPS_PROVISIONER_NAME, stackConfig)
	assert.Nil(t, err)
	assert.Equal(t, KopsProvisioner{
		stackConfig: stackConfig,
		kopsConfig: KopsConfig{
			Binary: "kops",
		},
	}, actual)
}

func TestNewNoOpProvisioner(t *testing.T) {
	actual, err := NewProvisioner(NOOP_PROVISIONER_NAME, &kapp.StackConfig{})
	assert.Nil(t, err)
	assert.Equal(t, NoOpProvisioner{}, actual)
}
