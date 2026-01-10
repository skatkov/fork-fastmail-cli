package cmd

import (
	"fmt"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/spf13/cobra"
)

func newEmailDeleteCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <emailId>",
		Short: "Delete email (move to trash)",
		Args:  cobra.ExactArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			err = client.DeleteEmail(cmd.Context(), args[0])
			if err != nil {
				return cerrors.WithContext(err, "deleting email")
			}

			if app.IsJSON(cmd.Context()) {
				return app.PrintJSON(cmd, map[string]any{
					"deleted": args[0],
				})
			}

			fmt.Printf("Email %s moved to trash\n", args[0])
			return nil
		}),
	}

	return cmd
}

func newEmailBulkDeleteCmd(app *App) *cobra.Command {
	var dryRun bool
	var yesFlag bool

	cmd := &cobra.Command{
		Use:   "bulk-delete <emailId>...",
		Short: "Delete multiple emails (move to trash)",
		Args:  cobra.MinimumNArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			// Handle dry-run mode
			if dryRun {
				return printDryRunList(app, cmd, fmt.Sprintf("Would delete %d emails:", len(args)), "wouldDelete", args, nil)
			}

			// Prompt for confirmation unless --yes flag is set or JSON output mode
			confirmed, err := app.Confirm(cmd, yesFlag, fmt.Sprintf("Delete %d emails? [y/N] ", len(args)), "y", "yes")
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("cancelled")
			}

			// Delete emails using bulk API
			results, err := client.DeleteEmails(cmd.Context(), args)
			if err != nil {
				return cerrors.WithContext(err, "deleting emails")
			}

			// Handle JSON output
			if app.IsJSON(cmd.Context()) {
				output := map[string]any{
					"succeeded": results.Succeeded,
				}
				if len(results.Failed) > 0 {
					output["failed"] = results.Failed
				}
				return app.PrintJSON(cmd, output)
			}

			// Handle text output
			succeededCount := len(results.Succeeded)
			failedCount := len(results.Failed)
			printBulkResults("Deleted", "emails", succeededCount, failedCount, results.Failed)

			return nil
		}),
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without making changes")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newEmailMoveCmd(app *App) *cobra.Command {
	var targetMailbox string

	cmd := &cobra.Command{
		Use:   "move <emailId>",
		Short: "Move email to mailbox",
		Args:  cobra.ExactArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			if targetMailbox == "" {
				return fmt.Errorf("--to is required")
			}

			// Resolve target mailbox ID or name
			resolvedID, err := client.ResolveMailboxID(cmd.Context(), targetMailbox)
			if err != nil {
				return fmt.Errorf("invalid target mailbox: %w", err)
			}
			targetMailbox = resolvedID

			err = client.MoveEmail(cmd.Context(), args[0], targetMailbox)
			if err != nil {
				return cerrors.WithContext(err, "moving email")
			}

			if app.IsJSON(cmd.Context()) {
				return app.PrintJSON(cmd, map[string]any{
					"moved":   args[0],
					"mailbox": targetMailbox,
				})
			}

			fmt.Printf("Email %s moved to mailbox %s\n", args[0], targetMailbox)
			return nil
		}),
	}

	cmd.Flags().StringVar(&targetMailbox, "to", "", "Target mailbox ID or name")

	return cmd
}

func newEmailBulkMoveCmd(app *App) *cobra.Command {
	var targetMailbox string
	var dryRun bool
	var yesFlag bool

	cmd := &cobra.Command{
		Use:   "bulk-move <emailId>...",
		Short: "Move multiple emails to a mailbox",
		Args:  cobra.MinimumNArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			// Validate required flags before accessing keyring
			if targetMailbox == "" {
				return fmt.Errorf("--to is required")
			}

			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			// Resolve target mailbox ID or name
			resolvedID, err := client.ResolveMailboxID(cmd.Context(), targetMailbox)
			if err != nil {
				return fmt.Errorf("invalid target mailbox: %w", err)
			}

			// Get mailbox name for output
			mailboxes, err := client.GetMailboxes(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get mailboxes: %w", err)
			}

			var mailboxName string
			for _, mb := range mailboxes {
				if mb.ID == resolvedID {
					mailboxName = mb.Name
					break
				}
			}
			if mailboxName == "" {
				mailboxName = resolvedID
			}

			// Handle dry-run mode
			if dryRun {
				return printDryRunList(app, cmd, fmt.Sprintf("Would move %d emails to %s:", len(args), mailboxName), "wouldMove", args, map[string]any{
					"mailbox": mailboxName,
				})
			}

			// Prompt for confirmation unless --yes flag is set or JSON output mode
			confirmed, err := app.Confirm(cmd, yesFlag, fmt.Sprintf("Move %d emails to %s? [y/N] ", len(args), mailboxName), "y", "yes")
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("cancelled")
			}

			// Move emails using bulk API
			results, err := client.MoveEmails(cmd.Context(), args, resolvedID)
			if err != nil {
				return cerrors.WithContext(err, "moving emails")
			}

			// Handle JSON output
			if app.IsJSON(cmd.Context()) {
				output := map[string]any{
					"mailbox":   mailboxName,
					"succeeded": results.Succeeded,
				}
				if len(results.Failed) > 0 {
					output["failed"] = results.Failed
				}
				return app.PrintJSON(cmd, output)
			}

			// Handle text output
			succeededCount := len(results.Succeeded)
			failedCount := len(results.Failed)
			printBulkResults("Moved", fmt.Sprintf("emails to %s", mailboxName), succeededCount, failedCount, results.Failed)

			return nil
		}),
	}

	cmd.Flags().StringVar(&targetMailbox, "to", "", "Target mailbox ID or name")
	cmd.Flags().StringVar(&targetMailbox, "mailbox", "", "Target mailbox ID or name (alias for --to)")
	_ = cmd.Flags().MarkHidden("mailbox") // Hidden alias for agent compatibility
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be moved without making changes")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func newEmailMarkReadCmd(app *App) *cobra.Command {
	var unread bool

	cmd := &cobra.Command{
		Use:   "mark-read <emailId>",
		Short: "Mark email as read/unread",
		Args:  cobra.ExactArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			err = client.MarkEmailRead(cmd.Context(), args[0], !unread)
			if err != nil {
				return fmt.Errorf("failed to update email: %w", err)
			}

			status := "read"
			if unread {
				status = "unread"
			}

			if app.IsJSON(cmd.Context()) {
				return app.PrintJSON(cmd, map[string]any{
					"emailId": args[0],
					"status":  status,
				})
			}

			fmt.Printf("Email %s marked as %s\n", args[0], status)
			return nil
		}),
	}

	cmd.Flags().BoolVar(&unread, "unread", false, "Mark as unread instead of read")

	return cmd
}

func newEmailBulkMarkReadCmd(app *App) *cobra.Command {
	var unread bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "bulk-mark-read <emailId>...",
		Short: "Mark multiple emails as read/unread",
		Args:  cobra.MinimumNArgs(1),
		RunE: runE(app, func(cmd *cobra.Command, args []string, app *App) error {
			client, err := app.JMAPClient()
			if err != nil {
				return err
			}

			status := "read"
			if unread {
				status = "unread"
			}

			// Handle dry-run mode
			if dryRun {
				return printDryRunList(app, cmd, fmt.Sprintf("Would mark %d emails as %s:", len(args), status), "wouldMark", args, map[string]any{
					"status": status,
				})
			}

			// Mark emails using bulk API
			results, err := client.MarkEmailsRead(cmd.Context(), args, !unread)
			if err != nil {
				return cerrors.WithContext(err, "marking emails")
			}

			// Handle JSON output
			if app.IsJSON(cmd.Context()) {
				output := map[string]any{
					"status":    status,
					"succeeded": results.Succeeded,
				}
				if len(results.Failed) > 0 {
					output["failed"] = results.Failed
				}
				return app.PrintJSON(cmd, output)
			}

			// Handle text output
			succeededCount := len(results.Succeeded)
			failedCount := len(results.Failed)
			printBulkResults("Marked", fmt.Sprintf("emails as %s", status), succeededCount, failedCount, results.Failed)

			return nil
		}),
	}

	cmd.Flags().BoolVar(&unread, "unread", false, "Mark as unread instead of read")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")

	return cmd
}
