// Copyright 2026 Maxime Delavergne and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"

	"github.com/mvanhorn/printing-press-library/library/media-and-entertainment/substack-reader/internal/client"
)

// resolveSubstackPublicationIDTemplate is a hand-authored no-op that satisfies
// calls the generator emits into internal/cli/sync.go and channel_workflow.go.
//
// Because the API name is "substack", the generator applied Substack-specific
// scaffolding that assumes a `drafts` resource requiring publication-ID
// resolution. This keyless reader spec defines no `drafts` resource, so the
// call site (`if resource == "drafts"`) is never reached at runtime — but the
// generated code references this symbol, and the generator did not emit its
// definition (a known "emits calls to an undefined helper" generator quirk).
//
// This file lives outside the generated set so it survives regeneration. If a
// draft/publication-scoped resource is ever added, implement the real
// publication-ID resolution (subdomain/custom-domain -> numeric publication_id
// via /api/v1/homepage_data) here.
func resolveSubstackPublicationIDTemplate(ctx context.Context, c *client.Client, flags *rootFlags) error {
	_ = ctx
	_ = c
	_ = flags
	return nil
}
