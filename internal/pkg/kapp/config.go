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

package kapp

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/vars"
	"os"
	"path/filepath"
)

const KAPP_CONFIG_FILE = "sugarkube.yaml"

// Loads the kapp's sugarkube.yaml file and sets it as an attribute on the kapp
func (k *Kapp) Load() error {
	configFilePath := filepath.Join(k.CacheDir(), KAPP_CONFIG_FILE)

	// return an error if the kapp doesn't have a sugarkube.yaml file.
	if _, err := os.Stat(configFilePath); err != nil {
		if os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("No '%s' file found for kapp "+
				"'%s' at %s", KAPP_CONFIG_FILE, k.FullyQualifiedId(), k.CacheDir()))
		}
	}

	config := Config{}
	err := vars.LoadYamlFile(configFilePath, &config)
	if err != nil {
		return errors.WithStack(err)
	}

	k.Config = config
	return nil
}
