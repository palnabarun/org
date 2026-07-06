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
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
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

// Transient devstats failures should be retried.
func TestFetchContributionsFromDevStats_RetriesOnServerError(t *testing.T) {
	restore := devstatsRetryBackoff
	devstatsRetryBackoff = time.Millisecond
	defer func() { devstatsRetryBackoff = restore }()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, `{"results":{"A":{"frames":[{"data":{"values":[[1],["alice"],[5]]}}]}}}`)
	}))
	defer srv.Close()

	source := devstatsSource{URL: srv.URL, Name: "test", DatasourceID: 1}
	contribs, err := fetchContributionsFromDevStats("y", source)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if _, ok := contribs["alice"]; !ok {
		t.Fatalf("expected alice after a successful retry")
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts (2 failures then success), got %d", got)
	}
}

// One failing source must not sink the whole audit.
func TestCombineContributions_ContinuesWhenOneSourceFails(t *testing.T) {
	restore := devstatsRetryBackoff
	devstatsRetryBackoff = time.Millisecond
	defer func() { devstatsRetryBackoff = restore }()

	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"results":{"A":{"frames":[{"data":{"values":[[1],["alice"],[5]]}}]}}}`)
	}))
	defer good.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()

	sources := []devstatsSource{
		{URL: good.URL, Name: "good", DatasourceID: 1},
		{URL: bad.URL, Name: "bad", DatasourceID: 1},
	}

	combined, err := combineContributions("y", sources)
	if err != nil {
		t.Fatalf("expected to continue despite one failing source, got %v", err)
	}
	if _, ok := combined["alice"]; !ok {
		t.Fatalf("expected contributor data from the healthy source")
	}
}

// A well-formed but empty result (e.g. an invalid --period) must be treated as
// an error rather than silently flagging every member as inactive.
func TestCombineContributions_ErrorsWhenNoContributors(t *testing.T) {
	restore := devstatsRetryBackoff
	devstatsRetryBackoff = time.Millisecond
	defer func() { devstatsRetryBackoff = restore }()

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"results":{"A":{"frames":[{"data":{"values":[[],[],[]]}}]}}}`)
	}))
	defer empty.Close()

	sources := []devstatsSource{{URL: empty.URL, Name: "empty", DatasourceID: 1}}
	if _, err := combineContributions("y", sources); err == nil {
		t.Fatalf("expected an error when devstats returns zero contributors")
	}
}

func keysOf(m map[string]Contribution) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
