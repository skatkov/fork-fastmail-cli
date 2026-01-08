package cmd

import "github.com/spf13/cobra"

func newEmailCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "Email operations",
	}

	cmd.AddCommand(newEmailListCmd(flags))
	cmd.AddCommand(newEmailSearchCmd(flags))
	cmd.AddCommand(newEmailGetCmd(flags))
	cmd.AddCommand(newEmailSendCmd(flags))
	cmd.AddCommand(newEmailDeleteCmd(flags))
	cmd.AddCommand(newEmailBulkDeleteCmd(flags))
	cmd.AddCommand(newEmailMoveCmd(flags))
	cmd.AddCommand(newEmailBulkMoveCmd(flags))
	cmd.AddCommand(newEmailMarkReadCmd(flags))
	cmd.AddCommand(newEmailBulkMarkReadCmd(flags))
	cmd.AddCommand(newEmailThreadCmd(flags))
	cmd.AddCommand(newEmailAttachmentsCmd(flags))
	cmd.AddCommand(newEmailDownloadCmd(flags))
	cmd.AddCommand(newEmailMailboxesCmd(flags))
	cmd.AddCommand(newMailboxCreateCmd(flags))
	cmd.AddCommand(newMailboxDeleteCmd(flags))
	cmd.AddCommand(newMailboxRenameCmd(flags))
	cmd.AddCommand(newEmailImportCmd(flags))
	cmd.AddCommand(newEmailIdentitiesCmd(flags))
	cmd.AddCommand(newEmailTrackCmd(flags))

	return cmd
}
