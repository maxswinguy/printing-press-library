package store_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/printing-press-library/library/payments/prediction-goat/internal/store"
)

// TestFTSSpeedrunQueries pins the FTS index against the five Problem
// Frame regressions that motivated this plan — Portugal World Cup,
// Bitcoin 100k, Oscars best picture, Fed rate cut, NBA championship.
// Each query must surface the correct top ticker against a seeded
// fixture that includes both the right answer and the false positives
// observed in the pre-fix CLI (notably KXFUSION for "oscars").
//
// This is the store-layer regression. U9 adds a CLI-layer end-to-end
// speedrun test that runs the full topic / compare flow.
func TestFTSSpeedrunQueries(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "speedrun.db")
	s, err := store.OpenWithContext(context.Background(), dbPath)
	if err != nil { t.Fatal(err) }
	defer s.Close()

	// Seed a realistic mix: the right answer for each query plus some
	// of the noise markets that polluted the original results.
	fixtures := []struct{ rt, id, data string }{
		// Portugal World Cup
		{"kalshi_markets", "KXMENWORLDCUP-26-PT", `{"ticker":"KXMENWORLDCUP-26-PT","event_ticker":"KXMENWORLDCUP-26","title":"Will the Portugal win the 2026 Men's World Cup?","yes_sub_title":"Portugal","category":"Sports"}`},
		{"kalshi_markets", "KXMENWORLDCUP-26-FR", `{"ticker":"KXMENWORLDCUP-26-FR","event_ticker":"KXMENWORLDCUP-26","title":"Will the France win the 2026 Men's World Cup?","yes_sub_title":"France","category":"Sports"}`},
		{"kalshi_series", "KXMENWORLDCUP", `{"ticker":"KXMENWORLDCUP","title":"Men's World Cup winner","category":"Sports"}`},

		// Bitcoin 100k
		{"kalshi_series", "KXBTCMAX100", `{"ticker":"KXBTCMAX100","title":"When will bitcoin hit 100k?","category":"Crypto"}`},
		{"kalshi_series", "KXBTC2026250", `{"ticker":"KXBTC2026250","title":"Will Bitcoin hit 250k in 2026?","category":"Crypto"}`},
		{"kalshi_events", "BTCETHATH-29DEC31", `{"event_ticker":"BTCETHATH-29DEC31","title":"Ethereum hits new all-time high before Bitcoin?","category":"Crypto"}`},

		// Oscars best picture - the original false positive KXFUSION must be excluded
		{"kalshi_series", "KXFUSION", `{"ticker":"KXFUSION","title":"Nuclear fusion","category":"Science","source_agencies":[{"name":"Academy of Motion Picture Arts and Sciences","url":"https://www.oscars.org/oscars"}]}`},
		{"kalshi_series", "KXOSCARCOUNTCONCLAVE", `{"ticker":"KXOSCARCOUNTCONCLAVE","title":"Conclave Oscar wins","category":"Entertainment"}`},
		{"kalshi_series", "KXOSCARSCORE", `{"ticker":"KXOSCARSCORE","title":"Oscars Best Score","category":"Entertainment"}`},

		// Fed rate cut
		{"kalshi_series", "KXRATECUT", `{"ticker":"KXRATECUT","title":"Fed rate cut","category":"Economics"}`},
		{"kalshi_series", "KXRATECUTS", `{"ticker":"KXRATECUTS","title":"Number of rate cuts","category":"Economics"}`},

		// NBA championship
		{"kalshi_series", "KXNBACHAMP", `{"ticker":"KXNBACHAMP","title":"NBA Championship winner","category":"Sports"}`},
	}
	for _, f := range fixtures {
		if err := s.Upsert(f.rt, f.id, json.RawMessage(f.data)); err != nil {
			t.Fatalf("upsert %s: %v", f.id, err)
		}
	}

	type expect struct {
		query        string
		wantTop      []string
		mustExclude  []string
	}
	cases := []expect{
		{"portugal world cup", []string{"KXMENWORLDCUP-26-PT"}, nil},
		{"bitcoin 100k", []string{"KXBTCMAX100"}, nil},
		{"oscars best picture", []string{"KXOSCARSCORE", "KXOSCARCOUNTCONCLAVE"}, []string{"KXFUSION"}},
		{"fed rate cut", []string{"KXRATECUT", "KXRATECUTS"}, nil},
		{"nba championship", []string{"KXNBACHAMP"}, nil},
	}

	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			tokens := strings.Fields(c.query)
			quoted := make([]string, 0, len(tokens))
			for _, tok := range tokens { quoted = append(quoted, `"`+tok+`"`) }
			ftsExpr := strings.Join(quoted, " OR ")

			rows, err := s.DB().Query(`SELECT id FROM resources_fts WHERE resources_fts MATCH ? ORDER BY rank LIMIT 10`, ftsExpr)
			if err != nil { t.Fatalf("query: %v", err) }
			defer rows.Close()
			var hits []string
			for rows.Next() { var id string; rows.Scan(&id); hits = append(hits, id) }
			t.Logf("query %q -> hits: %v", c.query, hits)

			// Top hit must be one of the expected
			if len(hits) == 0 {
				t.Errorf("no hits for query %q", c.query)
				return
			}
			topOK := false
			for _, w := range c.wantTop { if hits[0] == w { topOK = true; break } }
			if !topOK {
				t.Errorf("top hit for %q = %q, want one of %v", c.query, hits[0], c.wantTop)
			}
			for _, must := range c.mustExclude {
				for _, h := range hits {
					if h == must {
						t.Errorf("query %q surfaced forbidden hit %q in result %v", c.query, must, hits)
					}
				}
			}
		})
	}
}
