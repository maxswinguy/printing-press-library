// Copyright 2026 Maxime Delavergne and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/snipd/internal/snipd"
)

// TestNovelPullHelpWires smoke-tests that the pull command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelPullHelpWires(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{"pull", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pull --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "pull"} {
		if !strings.Contains(help, want) {
			t.Fatalf("pull --help missing %q in output:\n%s", want, help)
		}
	}
}

// TestFallbackSnipIDDistinguishesDistinctSnips guards the UUID-less fallback id:
// two snips sharing episode+start+title but with different content must NOT
// collide (the second Upsert would silently overwrite the first), while the same
// snip re-pulled must stay stable so a re-pull updates rather than duplicates.
func TestFallbackSnipIDDistinguishesDistinctSnips(t *testing.T) {
	base := snipd.Snip{EpisodeID: "ep1", Start: "15:11", Title: "Same Title"}
	a := base
	a.Note = "first note"
	b := base
	b.Note = "second note"

	if fallbackSnipID(a) == fallbackSnipID(b) {
		t.Fatal("distinct UUID-less snips sharing episode/start/title collided on the fallback id")
	}
	if fallbackSnipID(a) != fallbackSnipID(a) {
		t.Fatal("fallbackSnipID is not content-stable across calls")
	}
	if !strings.HasPrefix(fallbackSnipID(a), "ep1#") {
		t.Fatalf("fallback id %q is not namespaced under the episode id", fallbackSnipID(a))
	}
}

// TestTsLater guards the incremental cursor comparison: it must order by real
// instant, not lexically, so timestamps with different offsets or fractional
// precision can't park the cursor before episodes it should still cover.
func TestTsLater(t *testing.T) {
	cases := []struct {
		name, a, b string
		want       bool
	}{
		{"any ts beats the empty initial cursor", "2026-07-13T10:00:00Z", "", true},
		{"an earlier instant does not beat a later one", "2026-07-13T09:00:00Z", "2026-07-13T10:00:00Z", false},
		// Same instant, different offset + precision — lexical order would say true.
		{"same instant across offsets is not later", "2026-07-13T12:00:00.000+02:00", "2026-07-13T10:00:00Z", false},
		// Later instant whose string sorts lower than the Z form.
		{"a genuinely later instant across offsets", "2026-07-13T13:00:00+02:00", "2026-07-13T10:00:00Z", true},
	}
	for _, tc := range cases {
		if got := tsLater(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: tsLater(%q, %q) = %v, want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}
