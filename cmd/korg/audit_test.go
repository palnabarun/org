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
	"testing"
)

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
