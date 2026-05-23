// Copyright 2026 mvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"testing"
)

// TestBuildUnpaired covers the no-pair diagnostic payload: per-venue top
// N rows materialize into compareVenue shapes for the JSON envelope.
func TestBuildUnpaired(t *testing.T) {
	t.Parallel()
	pm := []rawMarket{
		{Venue: "polymarket", ID: "p1", Title: "A", YesProbability: 0.50, URL: "https://polymarket.com/market/p1"},
		{Venue: "polymarket", ID: "p2", Title: "B", YesProbability: 0.30, URL: "https://polymarket.com/market/p2"},
		{Venue: "polymarket", ID: "p3", Title: "C", YesProbability: 0.10, URL: "https://polymarket.com/market/p3"},
	}
	kalshi := []rawMarket{
		{Venue: "kalshi", ID: "K1", Title: "X", YesProbability: 0.70},
	}
	got := buildUnpaired(pm, kalshi, 2)
	if len(got.Polymarket) != 2 {
		t.Errorf("polymarket count = %d, want 2", len(got.Polymarket))
	}
	if len(got.Kalshi) != 1 {
		t.Errorf("kalshi count = %d, want 1", len(got.Kalshi))
	}
	if got.Polymarket[0].ID != "p1" {
		t.Errorf("polymarket[0].ID = %q, want p1", got.Polymarket[0].ID)
	}
	// yesPercent is populated even when missing from input (0 -> 0)
	if got.Polymarket[0].YesPercent != 50.0 {
		t.Errorf("polymarket[0].YesPercent = %v, want 50.0", got.Polymarket[0].YesPercent)
	}
	if got.Kalshi[0].YesPercent != 70.0 {
		t.Errorf("kalshi[0].YesPercent = %v, want 70.0", got.Kalshi[0].YesPercent)
	}
}

// TestBuildUnpaired_Empty handles the both-sides-empty case used when the
// topic doesn't match anything: returns a non-nil envelope with both
// lists empty so JSON shape is consistent.
func TestBuildUnpaired_Empty(t *testing.T) {
	t.Parallel()
	got := buildUnpaired(nil, nil, 5)
	if got == nil {
		t.Fatalf("buildUnpaired returned nil envelope")
	}
	if len(got.Polymarket) != 0 || len(got.Kalshi) != 0 {
		t.Errorf("expected empty lists, got pm=%d kalshi=%d", len(got.Polymarket), len(got.Kalshi))
	}
}

// TestCompareVenueFromRaw_UntradedPropagates verifies that compareVenue
// carries the Untraded flag from the underlying rawMarket so JSON
// consumers see untraded markets clearly when they appear in unpaired
// candidate lists.
func TestCompareVenueFromRaw_UntradedPropagates(t *testing.T) {
	t.Parallel()
	r := rawMarket{Venue: "kalshi", ID: "X", Title: "Untraded", YesProbability: 0, Untraded: true}
	v := compareVenueFromRaw(r)
	if !v.Untraded {
		t.Errorf("expected Untraded=true in compareVenue")
	}
	if v.YesPercent != 0 {
		t.Errorf("expected YesPercent=0 for zero-prob untraded market, got %v", v.YesPercent)
	}
}

// TestCompareVenueFromRaw_YesPercentPopulated locks the canonical pairing:
// canonical yesProbability (0-1 float) plus a yesPercent (0-100 rounded)
// for apples-to-apples display.
func TestCompareVenueFromRaw_YesPercentPopulated(t *testing.T) {
	t.Parallel()
	r := rawMarket{Venue: "polymarket", ID: "p", Title: "X", YesProbability: 0.792}
	v := compareVenueFromRaw(r)
	if v.YesPercent != 79.2 {
		t.Errorf("YesPercent = %v, want 79.2", v.YesPercent)
	}
	if v.YesProbability != 0.792 {
		t.Errorf("YesProbability = %v, want 0.792", v.YesProbability)
	}
}
