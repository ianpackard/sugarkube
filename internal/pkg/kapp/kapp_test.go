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
	"github.com/stretchr/testify/assert"
	"github.com/sugarkube/sugarkube/internal/pkg/acquirer"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"gopkg.in/yaml.v2"
	"testing"
)

func init() {
	log.ConfigureLogger("debug", false)
}

func TestParseManifestYaml(t *testing.T) {
	manifest := Manifest{
		Uri:          "fake/uri",
		ConfiguredId: "test-manifest",
	}

	tests := []struct {
		name                 string
		desc                 string
		input                string
		inputShouldBePresent bool
		expectValues         []Kapp
		expectedError        bool
	}{
		{
			name: "good_parse",
			desc: "check parsing acceptable input works",
			input: `
kapps:
  example1:
    state: present
    templates:        
      - source: example/template1.tpl
        dest: example/dest.txt
    sources:
      pathA:
        uri: git@github.com:exampleA/repoA.git//example/pathA#branchA
    sampleNameB:
      uri: git@github.com:exampleB/repoB.git//example/pathB#branchB

  example2:
    state: present
    sources:
      pathA:
        uri: git@github.com:exampleA/repoA.git//example/pathA#branchA
    vars:
      someVarA: valueA
      someList:
      - val1
      - val2

  example3:
    state: absent
    sources:
      pathA:
        uri: git@github.com:exampleA/repoA.git//example/pathA#branchA
`,
			expectValues: []Kapp{
				{
					Id:       "example1",
					State:    "present",
					manifest: &manifest,
					Templates: []Template{
						{
							"example/template1.tpl",
							"example/dest.txt",
						},
					},
					//Sources: []acquirer.Acquirer{
					//	discardErr(acquirer.NewGitAcquirer(
					//		"pathA",
					//		"git@github.com:exampleA/repoA.git",
					//		"branchA",
					//		"example/pathA",
					//		"")),
					//	discardErr(acquirer.NewGitAcquirer(
					//		"sampleNameB",
					//		"git@github.com:exampleB/repoB.git",
					//		"branchB",
					//		"example/pathB",
					//		"")),
					//},
					Sources: []acquirer.Source{
						{Id: "pathA",
							Uri: "git@github.com:exampleA/repoA.git//example/pathA#branchA"},
					},
				},
				{
					Id:       "example2",
					State:    "present",
					manifest: &manifest,
					//Sources: []acquirer.Acquirer{
					//	discardErr(acquirer.NewGitAcquirer(
					//		"pathA",
					//		"git@github.com:exampleA/repoA.git",
					//		"branchA",
					//		"example/pathA",
					//		"")),
					//},
					Sources: []acquirer.Source{
						{Id: "pathA",
							Uri: "git@github.com:exampleA/repoA.git//example/pathA#branchA"},
					},
					vars: map[string]interface{}{
						"someVarA": "valueA",
						"someList": []interface{}{
							"val1",
							"val2",
						},
					},
				},
				{
					Id:       "example3",
					State:    "absent",
					manifest: &manifest,
					//Sources: []acquirer.Acquirer{
					//	discardErr(acquirer.NewGitAcquirer(
					//		"pathA",
					//		"git@github.com:exampleA/repoA.git",
					//		"branchA",
					//		"example/pathA",
					//		"")),
					//},
					Sources: []acquirer.Source{
						{
							Id:  "pathA",
							Uri: "git@github.com:exampleA/repoA.git//example/pathA#branchA",
						},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		result := map[string]interface{}{}
		err := yaml.Unmarshal([]byte(test.input), &manifest)
		assert.Nil(t, err)

		if test.expectedError {
			assert.NotNil(t, err)
			assert.Nil(t, result)
		} else {
			assert.Equal(t, test.expectValues, result, "unexpected conversion result for %s", test.name)
			assert.Nil(t, err)
		}
	}

	assert.NotEqual(t, manifest, Manifest{})
}
