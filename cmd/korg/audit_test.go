/*
Copyright The Kubernetes Authors.

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
	"os"
	"path/filepath"
	"testing"
)

func writeOrgConfig(t *testing.T, root, orgName, contents string) {
	t.Helper()
	dir := filepath.Join(root, "config", orgName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "org.yaml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A member who IS present in devstats but whose contribution count is below the
// configured threshold must be reported as below-threshold. Because the guard in
// usernameBelowActivityThreshold was inverted, the function bailed out for anyone
// present in the contributions map and never compared the count.
func TestUsernameBelowActivityThreshold_PresentMemberBelowThreshold(t *testing.T) {
	contribs := map[string]Contribution{
		"alice": {Username: "alice", ContribCount: 5},
	}

	if !usernameBelowActivityThreshold(contribs, "alice", 100) {
		t.Fatalf("alice has 5 contributions and the threshold is 100: expected her to be flagged as below threshold, but she was not")
	}
}

// A member present in devstats with a count above the threshold is active and
// must not be flagged.
func TestUsernameBelowActivityThreshold_PresentMemberAboveThreshold(t *testing.T) {
	contribs := map[string]Contribution{
		"alice": {Username: "alice", ContribCount: 500},
	}

	if usernameBelowActivityThreshold(contribs, "alice", 100) {
		t.Fatalf("alice has 500 contributions against a threshold of 100: she must not be flagged as below threshold")
	}
}

// GetAllUsersInOrgs must load the orgs it is asked to audit (validOrgs) rather
// than only the orgs named by the --org flag. Otherwise a plain `korg audit`
// (no --org) loads an empty config and reports zero members.
func TestGetAllUsersInOrgs_LoadsRequestedOrgsWithoutOrgFlag(t *testing.T) {
	dir := t.TempDir()
	writeOrgConfig(t, dir, "kubernetes", "members:\n- alice\n")

	// o.Orgs is intentionally empty, mimicking `korg audit` with no --org.
	o := Options{RepoRoot: dir}
	users, err := GetAllUsersInOrgs(o, []string{"kubernetes"})
	if err != nil {
		t.Fatalf("GetAllUsersInOrgs: %v", err)
	}

	if _, ok := users["alice"]; !ok {
		t.Fatalf("expected member alice to be loaded from the requested org even without --org, got %d users", len(users))
	}
}

// GitHub logins are case-insensitive, and org.yaml, exceptions.csv and devstats
// may disagree on casing. Lookups must be case-insensitive so a differently
// cased member is not falsely reported as a non-contributor.
func TestUsernameNotInContributors_IsCaseInsensitive(t *testing.T) {
	contribs := map[string]Contribution{
		"alice": {Username: "alice", ContribCount: 5},
	}

	if usernameNotInContributors(contribs, "Alice") {
		t.Fatalf("Alice should match contributor alice case-insensitively")
	}
}

// A differently cased exception entry must still match.
func TestUsernameInExceptions_IsCaseInsensitive(t *testing.T) {
	if !usernameInExceptions([]string{"alice"}, "Alice") {
		t.Fatalf("Alice should match the exception 'alice' case-insensitively")
	}
}
