package cmd

import (
	"fmt"
	"os"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

func newEmailMailboxesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailboxes",
		Short: "List mailboxes (folders)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			mailboxes, err := client.GetMailboxes(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get mailboxes: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, mailboxes)
			}

			tw := newTabWriter()
			fmt.Fprintln(tw, "ID\tNAME\tROLE\tUNREAD\tTOTAL")
			for _, mb := range mailboxes {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\n",
					mb.ID,
					sanitizeTab(mb.Name),
					mb.Role,
					mb.UnreadEmails,
					mb.TotalEmails,
				)
			}
			tw.Flush()

			return nil
		},
	}

	return cmd
}

func newMailboxCreateCmd(flags *rootFlags) *cobra.Command {
	var parentID string

	cmd := &cobra.Command{
		Use:   "mailbox-create <name>",
		Short: "Create a new mailbox (folder)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			opts := jmap.CreateMailboxOpts{
				Name:     args[0],
				ParentID: parentID,
			}

			mailbox, err := client.CreateMailbox(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to create mailbox: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, mailbox)
			}

			fmt.Printf("Created mailbox '%s' (ID: %s)\n", mailbox.Name, mailbox.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&parentID, "parent", "", "Parent mailbox ID (for nested folders)")

	return cmd
}

func newMailboxDeleteCmd(flags *rootFlags) *cobra.Command {
	var yesFlag bool

	cmd := &cobra.Command{
		Use:   "mailbox-delete <mailbox-id-or-name>",
		Short: "Delete a mailbox (folder)",
		Long:  "Delete a mailbox. Emails in the mailbox will be moved to trash.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			// First resolve the mailbox ID
			mailboxID, err := client.ResolveMailboxID(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("invalid mailbox: %w", err)
			}

			// Then get mailboxes to find the name for the prompt
			mailboxes, err := client.GetMailboxes(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get mailboxes: %w", err)
			}

			// Find the mailbox name for the confirmation prompt
			var mailboxName string
			for _, mb := range mailboxes {
				if mb.ID == mailboxID {
					mailboxName = mb.Name
					break
				}
			}

			// Prompt for confirmation unless --yes flag is set or JSON output mode
			if !isJSON(cmd.Context()) && !yesFlag {
				confirmed, confirmErr := confirmPrompt(os.Stderr, fmt.Sprintf("Delete mailbox '%s' (ID: %s)? [y/N] ", mailboxName, mailboxID), "y", "yes")
				if confirmErr != nil {
					return confirmErr
				}
				if !confirmed {
					return fmt.Errorf("cancelled")
				}
			}

			err = client.DeleteMailbox(cmd.Context(), mailboxID)
			if err != nil {
				return fmt.Errorf("failed to delete mailbox: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"deleted": mailboxID,
				})
			}

			fmt.Printf("Deleted mailbox %s\n", mailboxID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newMailboxRenameCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailbox-rename <mailbox-id-or-name> <new-name>",
		Short: "Rename a mailbox (folder)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			// Resolve mailbox ID or name
			mailboxID, err := client.ResolveMailboxID(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("invalid mailbox: %w", err)
			}

			err = client.RenameMailbox(cmd.Context(), mailboxID, args[1])
			if err != nil {
				return fmt.Errorf("failed to rename mailbox: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"mailboxId": mailboxID,
					"newName":   args[1],
				})
			}

			fmt.Printf("Renamed mailbox %s to '%s'\n", mailboxID, args[1])
			return nil
		},
	}

	return cmd
}

func newEmailIdentitiesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identities",
		Short: "List sending identities (aliases)",
		Long:  "List all email identities/aliases you can send from.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			identities, err := client.GetIdentities(cmd.Context())
			if err != nil {
				idErr := cerrors.WithContext(err, "fetching identities")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(idErr, cerrors.SuggestionReauth)
				}
				return idErr
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, identities)
			}

			if len(identities) == 0 {
				printNoResults("No identities found")
				return nil
			}

			tw := newTabWriter()
			fmt.Fprintln(tw, "ID\tEMAIL\tNAME\tDEFAULT")
			for _, id := range identities {
				// MayDelete=false indicates the primary account identity
				isDefault := ""
				if !id.MayDelete {
					isDefault = "*"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					id.ID,
					id.Email,
					sanitizeTab(id.Name),
					isDefault,
				)
			}
			tw.Flush()

			return nil
		},
	}

	return cmd
}
