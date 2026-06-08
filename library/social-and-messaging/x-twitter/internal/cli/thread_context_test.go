// Copyright 2026 Cathryn Lavery and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import "testing"

func TestThreadParticipantsHandlesOptionalSections(t *testing.T) {
	result := &threadContextResult{
		FocusTweet: &resolvedPostRecord{
			TweetID: "1",
			Author:  &postAuthorSummary{ID: "10", Username: "alice"},
		},
		Ancestors: []resolvedPostRecord{{
			TweetID: "0",
			Author:  &postAuthorSummary{ID: "11", Username: "bob"},
		}},
		Replies: []threadContextReply{{
			resolvedPostRecord: resolvedPostRecord{
				TweetID: "2",
				Author:  &postAuthorSummary{ID: "10", Username: "alice"},
			},
			InReplyTo: "1",
			Depth:     1,
		}},
	}

	participants := threadParticipants(result)
	if len(participants) != 2 {
		t.Fatalf("participants len = %d, want 2: %+v", len(participants), participants)
	}
	if participants[0].ID != "10" || participants[1].ID != "11" {
		t.Fatalf("participants = %+v", participants)
	}
}
