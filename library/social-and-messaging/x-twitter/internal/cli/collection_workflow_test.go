// Copyright 2026 Cathryn Lavery and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCollectionSaveInputs(t *testing.T) {
	name, inputs, err := collectionSaveInputs("research", []string{"123"}, "")
	if err != nil {
		t.Fatalf("collectionSaveInputs documented form error: %v", err)
	}
	if name != "research" || len(inputs) != 1 || inputs[0] != "123" {
		t.Fatalf("documented form = %q %v", name, inputs)
	}

	name, inputs, err = collectionSaveInputs("", []string{"research", "123"}, "")
	if err != nil {
		t.Fatalf("collectionSaveInputs positional form error: %v", err)
	}
	if name != "research" || len(inputs) != 1 || inputs[0] != "123" {
		t.Fatalf("positional form = %q %v", name, inputs)
	}
}

func TestCollectionSaveInputsRequiresCollection(t *testing.T) {
	if _, _, err := collectionSaveInputs("", []string{"123"}, ""); err == nil {
		t.Fatal("collectionSaveInputs missing collection returned nil error")
	}
	if _, _, err := collectionSaveInputs("research", []string{"123"}, "agentic coding"); err == nil {
		t.Fatal("collectionSaveInputs accepted --from-search with explicit inputs")
	}
}

func TestWriteCollectionExportMarkdownAndJSONL(t *testing.T) {
	items := []collectionItemSnapshot{{
		TweetID: "123",
		URL:     "https://x.com/alice/status/123",
		Author:  &postAuthorSummary{Username: "alice"},
		Text:    "A useful post",
		Note:    "source material",
		Tags:    []string{"research"},
		SavedAt: "2026-01-01T00:00:00Z",
	}}

	var md bytes.Buffer
	if err := writeCollectionExport(&md, "research", items, "markdown"); err != nil {
		t.Fatalf("markdown export error: %v", err)
	}
	for _, want := range []string{"# research", "https://x.com/alice/status/123", "- Note: source material", "A useful post"} {
		if !strings.Contains(md.String(), want) {
			t.Fatalf("markdown export missing %q:\n%s", want, md.String())
		}
	}

	var jsonl bytes.Buffer
	if err := writeCollectionExport(&jsonl, "research", items, "jsonl"); err != nil {
		t.Fatalf("jsonl export error: %v", err)
	}
	if lines := strings.Split(strings.TrimSpace(jsonl.String()), "\n"); len(lines) != 1 || !strings.Contains(lines[0], `"tweet_id":"123"`) {
		t.Fatalf("jsonl export = %q", jsonl.String())
	}
}
