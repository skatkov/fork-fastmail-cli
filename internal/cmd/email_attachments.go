package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/format"
	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

func newEmailAttachmentsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachments <emailId>",
		Short: "List attachments for an email",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			attachments, err := client.GetEmailAttachments(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get attachments: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"emailId":     args[0],
					"attachments": attachments,
				})
			}

			if len(attachments) == 0 {
				printNoResults("No attachments")
				return nil
			}

			tw := newTabWriter()
			fmt.Fprintln(tw, "BLOB ID\tNAME\tTYPE\tSIZE")
			for _, att := range attachments {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					att.BlobID,
					sanitizeTab(att.Name),
					att.Type,
					format.FormatSize(att.Size),
				)
			}
			tw.Flush()

			return nil
		},
	}

	return cmd
}

func newEmailDownloadCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <emailId> <blobId> [output-file]",
		Short: "Download an email attachment",
		Long: `Download an email attachment by blob ID.

If output-file is not specified, the attachment will be saved with its original name
in the current directory. You can get the blob ID from the 'attachments' command.`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			emailID := args[0]
			blobID := args[1]
			var outputFile string

			// If output file not specified, get the attachment name from the email
			if len(args) < 3 {
				var attachments []jmap.Attachment
				attachments, err = client.GetEmailAttachments(cmd.Context(), emailID)
				if err != nil {
					return fmt.Errorf("failed to get attachments: %w", err)
				}

				// Find the attachment with matching blob ID
				var found bool
				for _, att := range attachments {
					if att.BlobID == blobID {
						outputFile = att.Name
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("blob ID '%s' not found in email '%s'", blobID, emailID)
				}

				if outputFile == "" {
					outputFile = "attachment"
				}

				// SECURITY: Sanitize auto-detected filename to prevent path traversal attacks
				// Only applied when filename comes from email attachment metadata (untrusted source)
				outputFile = sanitizeFilename(outputFile)
			} else {
				// User explicitly provided output path - respect it as-is
				// (both absolute paths like /tmp/file.pdf and relative paths like ./file.pdf)
				outputFile = args[2]
			}

			// Check if file already exists
			if _, statErr := os.Stat(outputFile); statErr == nil {
				return fmt.Errorf("file '%s' already exists. Specify a different output file", outputFile)
			}

			// Download the blob
			reader, err := client.DownloadBlob(cmd.Context(), blobID)
			if err != nil {
				downloadErr := cerrors.WithContext(err, "downloading attachment")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(downloadErr, cerrors.SuggestionReauth)
				}
				return downloadErr
			}
			defer reader.Close()

			// Create output file
			outFile, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			// Copy content
			written, err := io.Copy(outFile, reader)
			if err != nil {
				return fmt.Errorf("failed to write attachment: %w", err)
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"emailId":    emailID,
					"blobId":     blobID,
					"outputFile": outputFile,
					"size":       written,
				})
			}

			fmt.Printf("Downloaded attachment to %s (%s)\n", outputFile, format.FormatSize(written))
			return nil
		},
	}

	return cmd
}
