// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func mockFetchCommitSource(url, sha string) (*object.Commit, error) {
	return &object.Commit{
		Hash: plumbing.NewHash("6c1f093c0c197add71579d392da8a79a984fcd62"),
		Author: object.Signature{
			Name:  "ec RedHat",
			Email: "ec@gmail.com",
			When:  time.Time{},
		},
		Message: "Signed-off-by: EC <ec@redhat.com>",
	}, nil
}

func paramsInput(input string) attestation {
	params := materials{}
	if input == "good-commit" {
		params.Digest = map[string]string{
			"sha1": "6c1f093c0c197add71579d392da8a79a984fcd62",
		}
		params.Uri = "https://github.com/joejstuart/ec-cli.git"
	} else if input == "bad-commit" {
		params.Uri = ""
	} else if input == "bad-git" {
		params.Uri = ""
	} else if input == "good-git" {
		params.Uri = "https://github.com/joejstuart/ec-cli.git"
		params.Digest = map[string]string{
			"sha1": "6c1f093c0c197add71579d392da8a79a984fcd62",
		}
	}

	materials := []materials{
		params,
	}

	pred := predicate{
		Materials: materials,
	}
	att := attestation{
		Predicate: pred,
	}

	return att
}

func Test_NewGitSource(t *testing.T) {
	tests := []struct {
		input attestation
		want  *GitSource
		err   error
	}{
		{
			paramsInput("good-commit"),
			&GitSource{
				repoUrl:     "https://github.com/joejstuart/ec-cli.git",
				commitSha:   "6c1f093c0c197add71579d392da8a79a984fcd62",
				fetchSource: mockFetchCommitSource,
			},
			nil,
		},
		{
			paramsInput("bad-commit"),
			nil,
			fmt.Errorf(
				"there is no authorization source in attestation. sha: %v, url: %v",
				"",
				"",
			),
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("NewGitSource=%d", i), func(t *testing.T) {
			gitSource, err := tc.input.NewGitSource()
			assert.ObjectsAreEqualValues(tc.want, gitSource)
			assert.Equal(t, tc.err, err)
		})
	}
}

func Test_GetBuildCommitSha(t *testing.T) {
	tests := []struct {
		input attestation
		want  string
	}{
		{paramsInput("good-commit"), "6c1f093c0c197add71579d392da8a79a984fcd62"},
		{paramsInput("bad-commit"), ""},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("GetBuildCommitSha=%d", i), func(t *testing.T) {
			got := tc.input.getBuildCommitSha()
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_GetBuildSCM(t *testing.T) {
	tests := []struct {
		input attestation
		want  string
	}{
		{paramsInput("good-git"), "https://github.com/joejstuart/ec-cli.git"},
		{paramsInput("bad-git"), ""},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("GetBuildSCM=%d", i), func(t *testing.T) {
			got := tc.input.getBuildSCM()
			assert.Equal(t, tc.want, got)
		})
	}
}
