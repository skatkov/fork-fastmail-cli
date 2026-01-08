package cmd

import (
	"fmt"
	"os"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

func newEmailImportCmd(flags *rootFlags) *cobra.Command {
	var mailbox string
	var markRead bool

	cmd := &cobra.Command{
		Use:   "import <file.eml>",
		Short: "Import an email from a .eml file",
		Long: `Import a raw RFC 5322 email message (.eml file) into your mailbox.

The email will be imported with its original headers and content.
By default, emails are imported to the Inbox and marked as unread.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			emlPath := args[0]

			// Verify file exists
			fileInfo, err := os.Stat(emlPath)
			if err != nil {
				return fmt.Errorf("cannot access file '%s': %w", emlPath, err)
			}
			if fileInfo.IsDir() {
				return fmt.Errorf("cannot import directory: %s", emlPath)
			}

			// Determine target mailbox
			targetMailboxID := mailbox
			if targetMailboxID == "" {
				// Default to inbox
				var inbox *jmap.Mailbox
				inbox, err = client.GetMailboxByName(cmd.Context(), "inbox")
				if err != nil {
					return fmt.Errorf("failed to find inbox: %w", err)
				}
				targetMailboxID = inbox.ID
			} else {
				// Resolve mailbox name/ID
				var resolvedID string
				resolvedID, err = client.ResolveMailboxID(cmd.Context(), targetMailboxID)
				if err != nil {
					return fmt.Errorf("invalid mailbox: %w", err)
				}
				targetMailboxID = resolvedID
			}

			// Open and upload the .eml file
			file, err := os.Open(emlPath)
			if err != nil {
				return fmt.Errorf("failed to open file '%s': %w", emlPath, err)
			}
			defer file.Close()

			uploadResult, err := client.UploadBlob(cmd.Context(), file, "message/rfc822")
			if err != nil {
				return fmt.Errorf("failed to upload email: %w", err)
			}

			// Build import options
			opts := jmap.ImportEmailOpts{
				BlobID:     uploadResult.BlobID,
				MailboxIDs: map[string]bool{targetMailboxID: true},
			}

			if markRead {
				opts.Keywords = map[string]bool{"$seen": true}
			}

			emailID, err := client.ImportEmail(cmd.Context(), opts)
			if err != nil {
				importErr := cerrors.WithContext(err, "importing email")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(importErr, cerrors.SuggestionReauth)
				}
				return importErr
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"emailId":   emailID,
					"blobId":    uploadResult.BlobID,
					"mailboxId": targetMailboxID,
					"file":      emlPath,
				})
			}

			fmt.Printf("Imported email (ID: %s) from %s\n", emailID, emlPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&mailbox, "mailbox", "", "Target mailbox ID or name (default: Inbox)")
	cmd.Flags().BoolVar(&markRead, "read", false, "Mark imported email as read")

	return cmd
}
