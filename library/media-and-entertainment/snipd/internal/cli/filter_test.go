// Copyright 2026 Maxime Delavergne and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestNovelFilterHelpWires smoke-tests that the filter command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelFilterHelpWires(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{"filter", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("filter --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "filter"} {
		if !strings.Contains(help, want) {
			t.Fatalf("filter --help missing %q in output:\n%s", want, help)
		}
	}
}

// TestValidateDateFlag guards --since/--until input validation: a valid date or
// RFC3339 timestamp (and the empty/unset case) passes, while a typo that would
// otherwise become a silent lexical SQL predicate is rejected with an error.
func TestValidateDateFlag(t *testing.T) {
	valid := []string{"", "2026-05-01", "2026-05-01T12:30:00Z", "2026-05-01T12:30:00.500+02:00", "2026-05-01T12:30:00"}
	for _, v := range valid {
		if err := validateDateFlag("since", v); err != nil {
			t.Errorf("validateDateFlag(since, %q) = %v, want nil", v, err)
		}
	}
	invalid := []string{"zzzz", "2026-13-40", "last week", "05/01/2026"}
	for _, v := range invalid {
		if err := validateDateFlag("until", v); err == nil {
			t.Errorf("validateDateFlag(until, %q) = nil, want an input error", v)
		}
	}
}
