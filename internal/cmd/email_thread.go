package cmd

import (
	"fmt"

	"github.com/salmonumbrella/fastmail-cli/internal/format"
	"github.com/spf13/cobra"
)

func newEmailThreadCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread <threadId>",
		Short: "Get all emails in a thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			emails, err := client.GetThread(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get thread: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"threadId": args[0],
					"emails":   emailsToOutput(emails),
				})
			}

			if len(emails) == 0 {
				printNoResults("No emails found in thread")
				return nil
			}

			fmt.Printf("Thread: %s (%d messages)\n\n", args[0], len(emails))

			tw := newTabWriter()
			fmt.Fprintln(tw, "ID\tSUBJECT\tFROM\tDATE")
			for _, email := range emails {
				from := format.FormatEmailAddressList(email.From)
				date := format.FormatEmailDate(email.ReceivedAt)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					email.ID,
					sanitizeTab(format.Truncate(email.Subject, 40)),
					sanitizeTab(format.Truncate(from, 25)),
					date,
				)
			}
			tw.Flush()

			return nil
		},
	}

	return cmd
}
