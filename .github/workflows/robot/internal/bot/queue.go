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
	"context"
	"log"

	"github.com/gravitational/teleport/.github/workflows/robot/internal/github"
	"github.com/gravitational/trace"
)

// Queue will update all eligible PRs with the base branch.
func (b *Bot) Queue(ctx context.Context) error {
	pulls, err := b.c.GitHub.ListPullRequests(ctx,
		b.c.Environment.Organization,
		b.c.Environment.Repository,
		"open")
	if err != nil {
		return trace.Wrap(err)
	}

	// Filter out any PRs that are not eligible to be updated (not out of date,
	// auto merge not enabled, etc).
	pulls, err = b.filter(ctx, pulls)
	if err != nil {
		return trace.Wrap(err)
	}

	for _, pull := range pulls {
		err = b.c.GitHub.UpdateBranch(ctx,
			b.c.Environment.Organization,
			b.c.Environment.Repository,
			pull.Number)
		if err != nil {
			log.Printf("Failed to update PR %v: %v.", pull.Number, err)
			continue
		}
	}

	return nil
}

func (b *Bot) filter(ctx context.Context, pulls []github.PullRequest) ([]github.PullRequest, error) {
	// Build out a cache of SHAs for base branches, this is to prevent hammering
	// the GitHub API when 100+ PRs are checked if they are up to date with
	// the base branch SHA.
	cache := map[string]string{}
	for _, pull := range pulls {
		if _, ok := cache[pull.UnsafeBaseRef]; !ok {
			branch, err := b.c.GitHub.GetBranch(ctx,
				b.c.Environment.Organization,
				b.c.Environment.Repository,
				pull.UnsafeBaseRef)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			cache[pull.UnsafeBaseRef] = branch.SHA
		}
	}

	var filtered []github.PullRequest
	for _, pull := range pulls {
		if !canUpdate(pull, cache) {
			continue
		}
		filtered = append(filtered, pull)
	}

	return filtered, nil
}

// canUpdate returns if this branch should be skipped over for updating. PRs
// are skipped over if they are from a fork, are a branch other than master,
// auto-merge is disabled, or are not mergeable.
func canUpdate(pull github.PullRequest, cache map[string]string) bool {
	// The "merge queue" does not support forks.
	if pull.Fork {
		return false
	}
	// For now, the "merge queue" only supports merging into master.
	if pull.UnsafeBaseRef != "master" {
		return false
	}
	// Skip branches that are up-to-date with the base branch.
	branchBaseSHA, ok := cache[pull.UnsafeBaseRef]
	if ok && pull.BaseSHA == branchBaseSHA {
		return false
	}
	// Skip over PRs that do not actually want to use the "merge queue".
	if !pull.AutoMerge {
		return false
	}

	return true
}
