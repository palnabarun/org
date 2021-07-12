/*
Copyright 2021 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"k8s.io/test-infra/prow/config/org"
	"sigs.k8s.io/yaml"
)

func stringInSlice(slice []string, key string) bool {
	for _, e := range slice {
		if key == e {
			return true
		}
	}

	return false
}

func validateOrgs(orgs []string) (valid []string, invalid []string, invalidPresent bool) {
	valid = []string{}
	invalid = []string{}
	invalidPresent = false

	for _, org := range orgs {
		if !stringInSlice(validOrgs, org) {
			invalid = append(invalid, org)
		} else {
			valid = append(valid, org)
		}
	}

	if len(invalid) > 0 {
		invalidPresent = true
	}

	return
}

func readConfig(path string) (*org.Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file at %s: %s", path, err)
	}

	config := org.Config{}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config from %s: %s", path, err)
	}

	return &config, nil
}

func saveConfig(path string, config *org.Config) error {
	b, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("unable to marshal config: %s", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("unable to fetch info for %s: %s", path, err)
	}

	if err := os.WriteFile(path, b, info.Mode()); err != nil {
		return fmt.Errorf("unable to write to %s: %s", path, err)
	}
	return nil
}

func commitChanges(repoRoot string, configsModified []string, message string) error {
	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("unable to open repository: %s", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("unable to fetch worktree: %s", err)
	}

	for _, configModified := range configsModified {
		_, err := w.Add(configModified)
		if err != nil {
			return fmt.Errorf("unable to stage changes: %s", err)
		}
	}

	commit, err := w.Commit(message, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("unable to commit changes: %s", err)
	}

	_, err = r.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("unable to write commit object to repo: %s", err)
	}

	return nil
}

// this is an custom implementation which sorts a string slice
// with a case agnostic heuristic. the default sorting algorithm
// in Go doesn't ignore case, whereas we do.
type caseAgnostic []string

func (s caseAgnostic) Len() int {
	return len(s)
}

func (s caseAgnostic) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s caseAgnostic) Less(i, j int) bool {
	li := strings.ToLower(s[i])
	lj := strings.ToLower(s[j])

	return sort.StringsAreSorted([]string{li, lj})
}

func caseAgnosticSort(arr []string) {
	sort.Sort(caseAgnostic(arr))
}

func parseTeam(team string) (string, string, string, error) {
	parts := strings.Split(team, "/")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", "", fmt.Errorf("invalid team: %s", team)
	}

	org := parts[0]
	if !stringInSlice(validOrgs, org) {
		return "", "", "", fmt.Errorf("invalid team: %s", team)
	}

	group := ""
	if len(parts) == 2 {
		group = ""
		team = parts[1]
	} else {
		group = parts[1]
		team = parts[2]
	}

	return org, group, team, nil

}

func userInOrg(username string, org string, options Options) bool {
	configPath := filepath.Join(options.RepoRoot, fmt.Sprintf(orgConfigPathFormat, org))
	config, err := readConfig(configPath)
	if err != nil {
		return false
	}

	if !stringInSlice(config.Members, username) {
		return false
	}

	return true
}
