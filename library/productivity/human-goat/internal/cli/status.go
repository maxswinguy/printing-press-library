// Copyright 2026 Matt Van Horn and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/productivity/human-goat/internal/source/taskrabbit"
)

func newNovelStatusCmd(flags *rootFlags) *cobra.Command {
	var flagOpen bool

	cmd := &cobra.Command{
		Use:         "status",
		Short:       "One list of every in-flight task across TaskRabbit bookings and Magic requests",
		Example:     "  human-goat-pp-cli status --open --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !commandHasChangedFlags(cmd) {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would list in-flight tasks across sources")
				return nil
			}
			if len(args) > 0 {
				return usageErr(fmt.Errorf("status does not accept positional arguments"))
			}

			out := statusOutput{
				Tasks:     make([]statusRow, 0),
				MagicNote: "Magic requests are tracked per-id; run `track <id>` for a specific request.",
			}

			c, err := flags.newClient()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "taskrabbit: %v\n", err)
			} else {
				tr := taskrabbit.New(c)
				bookings, err := tr.ListTasks(cmd.Context(), 1, 20, map[string]any{}, "en-US")
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "taskrabbit: %v\n", err)
				} else {
					for _, booking := range bookings {
						if flagOpen {
							// TaskRabbit's task list is already scoped to current bookings.
						}
						out.Tasks = append(out.Tasks, statusRow{
							Source: "taskrabbit",
							ID:     booking.ID,
							Status: booking.Status,
							Title:  "",
						})
					}
				}
			}

			if flags.asJSON || flags.agent {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			tableRows := make([][]string, 0, len(out.Tasks))
			for _, row := range out.Tasks {
				tableRows = append(tableRows, []string{row.Source, row.ID, row.Status, row.Title})
			}
			if err := flags.printTable(cmd, []string{"SOURCE", "ID", "STATUS", "TITLE"}, tableRows); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), out.MagicNote)
			return nil
		},
	}
	cmd.Flags().BoolVar(&flagOpen, "open", false, "Show only in-flight work")
	return cmd
}

type statusOutput struct {
	Tasks     []statusRow `json:"tasks"`
	MagicNote string      `json:"magic_note"`
}

type statusRow struct {
	Source string `json:"source"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}
