/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bot

import (
	"testing"

	"github.com/gravitational/teleport/.github/workflows/robot/internal/github"

	"github.com/stretchr/testify/require"
)

// TestSkipUpdate checks if PRs are appropriately skipped over.
func TestSkipUpdate(t *testing.T) {
	tests := []struct {
		desc      string
		pr        github.PullRequest
		cache     map[string]string
		canUpdate bool
	}{
		{
			desc: "fork-skip",
			pr: github.PullRequest{
				Author:        "foo",
				Repository:    "bar",
				UnsafeHeadRef: "baz/qux",
				HeadSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				UnsafeBaseRef: "master",
				BaseSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				Fork:          true,
				AutoMerge:     true,
			},
			cache: map[string]string{
				"master": "0000000000000000000000000000000000000000000000000000000000000001",
			},
			canUpdate: false,
		},
		{
			desc: "non-master-skip",
			pr: github.PullRequest{
				Author:        "foo",
				Repository:    "bar",
				UnsafeHeadRef: "baz/qux",
				HeadSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				UnsafeBaseRef: "branch/v0",
				BaseSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				Fork:          false,
				AutoMerge:     true,
			},
			cache: map[string]string{
				"master": "0000000000000000000000000000000000000000000000000000000000000001",
			},
			canUpdate: false,
		},
		{
			desc: "no-auto-merge-skip",
			pr: github.PullRequest{
				Author:        "foo",
				Repository:    "bar",
				UnsafeHeadRef: "baz/qux",
				HeadSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				UnsafeBaseRef: "master",
				BaseSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				Fork:          false,
				AutoMerge:     false,
			},
			cache: map[string]string{
				"master": "0000000000000000000000000000000000000000000000000000000000000001",
			},
			canUpdate: false,
		},
		{
			desc: "up-to-date-skip",
			pr: github.PullRequest{
				Author:        "foo",
				Repository:    "bar",
				UnsafeHeadRef: "baz/qux",
				HeadSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				UnsafeBaseRef: "master",
				BaseSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				Fork:          false,
				AutoMerge:     true,
			},
			cache: map[string]string{
				"master": "0000000000000000000000000000000000000000000000000000000000000001",
			},
			canUpdate: false,
		},
		{
			desc: "not-up-to-date-update",
			pr: github.PullRequest{
				Author:        "foo",
				Repository:    "bar",
				UnsafeHeadRef: "baz/qux",
				HeadSHA:       "0000000000000000000000000000000000000000000000000000000000000001",
				UnsafeBaseRef: "master",
				BaseSHA:       "0000000000000000000000000000000000000000000000000000000000000002",
				Fork:          false,
				AutoMerge:     true,
			},
			cache: map[string]string{
				"master": "0000000000000000000000000000000000000000000000000000000000000001",
			},
			canUpdate: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ok := canUpdate(test.pr, test.cache)
			require.Equal(t, ok, test.canUpdate)
		})
	}
}
