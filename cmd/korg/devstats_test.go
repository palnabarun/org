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

func TestParseDevStatsResponse_ValidBody(t *testing.T) {
	body := `{"results":{"A":{"frames":[{"data":{"values":[[1,2],["Alice","bob"],[10,5]]}}]}}}`

	contribs, err := parseDevStatsResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Contributors are keyed by their normalized (lower-cased) login.
	alice, ok := contribs["alice"]
	if !ok {
		t.Fatalf("expected contributor alice, got keys %v", keysOf(contribs))
	}
	if alice.ContribCount != 10 {
		t.Fatalf("alice ContribCount = %d, want 10", alice.ContribCount)
	}
	if _, ok := contribs["bob"]; !ok {
		t.Fatalf("expected contributor bob, got keys %v", keysOf(contribs))
	}
}

// A devstats endpoint that returns a body which does not match the expected
// shape (an error payload, empty frames, missing rows, or a wrongly typed cell)
// must yield an error, not a panic.
func TestParseDevStatsResponse_Malformed(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty object", `{}`},
		{"empty frames", `{"results":{"A":{"frames":[]}}}`},
		{"frame with no rows", `{"results":{"A":{"frames":[{"data":{"values":[]}}]}}}`},
		{"non-numeric contribution count", `{"results":{"A":{"frames":[{"data":{"values":[[1],["alice"],["not-a-number"]]}}]}}}`},
		{"mismatched column lengths", `{"results":{"A":{"frames":[{"data":{"values":[[1,2],["alice"],[5,6]]}}]}}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("parseDevStatsResponse panicked on a malformed response (want a graceful error): %v", r)
				}
			}()

			if _, err := parseDevStatsResponse([]byte(tc.body)); err == nil {
				t.Fatalf("expected an error for a malformed devstats response, got nil")
			}
		})
	}
}

func keysOf(m map[string]Contribution) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
