// Copyright 2026 Matt Van Horn and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.

package cli

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/productivity/human-goat/internal/client"
	"github.com/mvanhorn/printing-press-library/library/productivity/human-goat/internal/cliutil"
	"github.com/mvanhorn/printing-press-library/library/productivity/human-goat/internal/source/taskrabbit"

	"github.com/spf13/cobra"
)

func newNovelCancelCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "cancel <booking-id>",
		Short:   "Cancels a booking and confirms it landed by re-reading status",
		Example: "human-goat-pp-cli cancel task_abc123 --agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !commandHasChangedFlags(cmd) {
				return cmd.Help()
			}
			if len(args) == 0 {
				return usageErr(fmt.Errorf("missing booking-id"))
			}
			if len(args) > 1 {
				return usageErr(fmt.Errorf("cancel accepts one booking-id"))
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return usageErr(fmt.Errorf("missing booking-id"))
			}
			if dryRunOK(flags) || cliutil.IsVerifyEnv() {
				fmt.Fprintf(cmd.OutOrStdout(), "would cancel booking %s and re-read status to verify\n", id)
				return nil
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			tr := taskrabbit.New(c)

			jobID, convErr := strconv.Atoi(id)
			if convErr != nil {
				return usageErr(fmt.Errorf("booking-id must be numeric (the jobId from `tasks list`), got %q", id))
			}

			// Look up the booking to get its rabbitId (taskers[0].id), which
			// cancelTask requires alongside jobId.
			active, err := tr.ListTasks(ctx, 1, 50, map[string]any{"status": "active"}, "en-US")
			if err != nil {
				return classifyAPIError(fmt.Errorf("cancel: look up booking %d: %w", jobID, err), flags)
			}
			var rabbitID int
			var found bool
			for _, b := range active {
				if b.JobID == jobID {
					rabbitID = b.RabbitID
					found = true
					break
				}
			}
			if !found {
				return classifyAPIError(fmt.Errorf("cancel: no active booking with id %d (run `tasks list` to see current bookings)", jobID), flags)
			}

			token, tokenErr := csrfToken(ctx, flags)
			tokenNote := ""
			if tokenErr != nil {
				tokenNote = "could not obtain CSRF token"
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", tokenNote, tokenErr)
			}

			if _, err := tr.CancelTask(ctx, jobID, rabbitID, "Plans changed, no longer need the help", token); err != nil {
				return classifyAPIError(fmt.Errorf("cancel TaskRabbit booking %d: %w", jobID, err), flags)
			}

			// Verify: re-read active bookings; the cancelled booking must be gone.
			after, err := tr.ListTasks(ctx, 1, 50, map[string]any{"status": "active"}, "en-US")
			if err != nil {
				return classifyAPIError(fmt.Errorf("cancel TaskRabbit booking %d succeeded, but verify re-read failed: %w", jobID, err), flags)
			}
			stillActive := false
			for _, b := range after {
				if b.JobID == jobID {
					stillActive = true
					break
				}
			}
			result := cancelResult{
				BookingID:        id,
				CancelResponseOK: true,
				Note:             tokenNote,
			}
			if stillActive {
				result.VerifiedStatus = "still-active"
				result.Note = joinNotes(result.Note, "WARNING: booking still appears active after cancel; verify in the app")
			} else {
				result.VerifiedStatus = "cancelled"
				result.Note = joinNotes(result.Note, "verified: booking no longer active; deposit refundable if cancelled >=24h before the appointment")
			}
			return printCancelResult(cmd, flags, result)
		},
	}
	return cmd
}

type cancelResult struct {
	BookingID        string `json:"booking_id"`
	CancelResponseOK bool   `json:"cancel_response_ok"`
	VerifiedStatus   string `json:"verified_status"`
	Note             string `json:"note,omitempty"`
}

func csrfToken(ctx context.Context, flags *rootFlags) (string, error) {
	c, err := flags.newClient()
	if err != nil {
		return "", err
	}
	body, err := c.GetWithHeadersNoCache(ctx, "/dashboard", nil, map[string]string{
		client.BinaryResponseHeader: "true",
		"Accept":                    "text/html,*/*",
	})
	if err != nil {
		return "", err
	}
	// TaskRabbit renders `<meta name="csrf-token" content="..." />` (self-closing,
	// space before />), so match up to the closing quote rather than a literal ">".
	matches := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)"`).FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("csrf-token meta tag not found")
	}
	return string(matches[1]), nil
}

func joinNotes(notes ...string) string {
	out := make([]string, 0, len(notes))
	for _, note := range notes {
		note = strings.TrimSpace(note)
		if note != "" {
			out = append(out, note)
		}
	}
	return strings.Join(out, "; ")
}

func printCancelResult(cmd *cobra.Command, flags *rootFlags, result cancelResult) error {
	if flags.asJSON || flags.agent {
		return printJSONFiltered(cmd.OutOrStdout(), result, flags)
	}
	rows := [][]string{
		{"Booking ID", result.BookingID},
		{"Cancel response OK", fmt.Sprintf("%t", result.CancelResponseOK)},
		{"Verified status", result.VerifiedStatus},
	}
	if result.Note != "" {
		rows = append(rows, []string{"Note", result.Note})
	}
	return flags.printTable(cmd, []string{"FIELD", "VALUE"}, rows)
}
