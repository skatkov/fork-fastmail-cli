package cmd

import (
	"fmt"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/format"
	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

func newEmailListCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var mailboxID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List emails",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			// Resolve mailbox ID or name
			if mailboxID != "" {
				var resolvedID string
				resolvedID, err = client.ResolveMailboxID(cmd.Context(), mailboxID)
				if err != nil {
					return fmt.Errorf("invalid mailbox: %w", err)
				}
				mailboxID = resolvedID
			}

			emails, err := client.GetEmails(cmd.Context(), mailboxID, limit)
			if err != nil {
				listErr := cerrors.WithContext(err, "listing emails")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(listErr, cerrors.SuggestionReauth)
				}
				return listErr
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, emailsToOutput(emails))
			}

			if len(emails) == 0 {
				printNoResults("No emails found")
				return nil
			}

			tw := newTabWriter()
			fmt.Fprintln(tw, "ID\tSUBJECT\tFROM\tDATE\tUNREAD")
			for _, email := range emails {
				from := format.FormatEmailAddressList(email.From)
				date := format.FormatEmailDate(email.ReceivedAt)
				unread := ""
				if email.Keywords != nil && !email.Keywords["$seen"] {
					unread = "*"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					email.ID,
					sanitizeTab(format.Truncate(email.Subject, 50)),
					sanitizeTab(format.Truncate(from, 30)),
					date,
					unread,
				)
			}
			tw.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of emails to list")
	cmd.Flags().StringVar(&mailboxID, "mailbox", "", "Mailbox ID or name to filter emails")

	return cmd
}

func newEmailSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var snippets bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search emails",
		Long: `Search emails using JMAP query syntax.

Examples:
  fastmail email search "from:alice@example.com"
  fastmail email search --snippets "invoice"
  fastmail email search "subject:meeting after:2025-01-01"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			var emails []jmap.Email
			var searchSnippets []jmap.SearchSnippet

			if snippets {
				emails, searchSnippets, err = client.SearchEmailsWithSnippets(cmd.Context(), args[0], limit)
			} else {
				emails, err = client.SearchEmails(cmd.Context(), args[0], limit)
			}

			if err != nil {
				searchErr := cerrors.WithContext(err, "searching emails")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(searchErr, cerrors.SuggestionReauth)
				}
				return searchErr
			}

			if isJSON(cmd.Context()) {
				result := map[string]any{"emails": emailsToOutput(emails)}
				if snippets && len(searchSnippets) > 0 {
					result["snippets"] = searchSnippets
				}
				return printJSON(cmd, result)
			}

			if len(emails) == 0 {
				printNoResults("No emails found matching '%s'", args[0])
				return nil
			}

			// Build snippet map for quick lookup
			snippetMap := make(map[string]jmap.SearchSnippet)
			for _, s := range searchSnippets {
				snippetMap[s.EmailID] = s
			}

			tw := newTabWriter()
			fmt.Fprintln(tw, "ID\tSUBJECT\tFROM\tDATE\tUNREAD")
			for _, email := range emails {
				from := format.FormatEmailAddressList(email.From)
				date := format.FormatEmailDate(email.ReceivedAt)
				unread := ""
				if email.Keywords != nil && !email.Keywords["$seen"] {
					unread = "*"
				}

				subject := email.Subject
				if snippets {
					if s, ok := snippetMap[email.ID]; ok && s.Subject != "" {
						subject = s.Subject // Use highlighted subject
					}
				}

				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					email.ID,
					sanitizeTab(format.Truncate(subject, 50)),
					sanitizeTab(format.Truncate(from, 30)),
					date,
					unread,
				)

				// Show snippet preview if available
				if snippets {
					if s, ok := snippetMap[email.ID]; ok && s.Preview != "" {
						fmt.Fprintf(tw, "\t%s\t\t\t\n", sanitizeTab(format.Truncate(s.Preview, 80)))
					}
				}
			}
			tw.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of results")
	cmd.Flags().BoolVar(&snippets, "snippets", false, "Show highlighted search snippets")

	return cmd
}
