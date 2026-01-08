package cmd

import (
	"fmt"
	"os"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/format"
	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
)

func newEmailSendCmd(flags *rootFlags) *cobra.Command {
	var to, cc, bcc []string
	var subject, body, htmlBody string
	var draft bool
	var replyTo string
	var attachments []string
	var fromIdentity string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email",
		Long: `Send an email with optional attachments.

Examples:
  fastmail email send --to user@example.com --subject "Hello" --body "Hi there"
  fastmail email send --to user@example.com --subject "Report" --body "See attached" --attach report.pdf
  fastmail email send --to user@example.com --subject "Q4 Results" --attach /docs/q4.pdf:Q4-Report.pdf`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			// For drafts with --reply-to, --to and --subject are optional (auto-filled)
			if !draft && replyTo == "" && len(to) == 0 {
				return fmt.Errorf("--to is required (or use --draft to save without sending)")
			}
			if replyTo == "" && subject == "" {
				return fmt.Errorf("--subject is required")
			}
			if body == "" && htmlBody == "" {
				return fmt.Errorf("--body or --html is required")
			}

			// Validate email addresses (only those provided)
			allAddrs := append(append(to, cc...), bcc...)
			for _, addr := range allAddrs {
				if !isValidEmail(addr) {
					return fmt.Errorf("invalid email address: %s", addr)
				}
			}

			// Process attachments
			var attachmentOpts []jmap.AttachmentOpts
			for _, att := range attachments {
				var attPath, attName string
				attPath, attName, err = parseAttachmentFlag(att)
				if err != nil {
					return fmt.Errorf("invalid attachment: %w", err)
				}

				// Verify file exists and get size
				var fileInfo os.FileInfo
				fileInfo, err = os.Stat(attPath)
				if err != nil {
					return fmt.Errorf("cannot access attachment '%s': %w", attPath, err)
				}
				if fileInfo.IsDir() {
					return fmt.Errorf("cannot attach directory: %s", attPath)
				}

				// Check file size before upload
				if fileInfo.Size() > jmap.MaxUploadSize {
					return fmt.Errorf("attachment '%s' too large (%s, max 50 MB)", attPath, format.FormatBytes(fileInfo.Size()))
				}

				// Open and upload the file
				var file *os.File
				file, err = os.Open(attPath)
				if err != nil {
					return fmt.Errorf("failed to open attachment '%s': %w", attPath, err)
				}

				mimeType := getMimeType(attPath)
				var uploadResult *jmap.UploadBlobResult
				uploadResult, err = client.UploadBlob(cmd.Context(), file, mimeType)
				_ = file.Close()
				if err != nil {
					return fmt.Errorf("failed to upload attachment '%s': %w", attPath, err)
				}

				attachmentOpts = append(attachmentOpts, jmap.AttachmentOpts{
					BlobID: uploadResult.BlobID,
					Name:   attName,
					Type:   mimeType,
				})
			}

			opts := jmap.SendEmailOpts{
				To:          to,
				CC:          cc,
				BCC:         bcc,
				Subject:     subject,
				TextBody:    body,
				HTMLBody:    htmlBody,
				From:        fromIdentity,
				Attachments: attachmentOpts,
			}

			if draft {
				var draftID string

				if replyTo != "" {
					// Create a threaded reply draft
					draftID, err = client.CreateReplyDraft(cmd.Context(), replyTo, opts)
				} else {
					// Create a standalone draft
					draftID, err = client.SaveDraft(cmd.Context(), opts)
				}

				if err != nil {
					return fmt.Errorf("failed to save draft: %w", err)
				}

				if isJSON(cmd.Context()) {
					return printJSON(cmd, map[string]any{
						"draftId": draftID,
						"status":  "draft",
						"replyTo": replyTo,
					})
				}

				fmt.Printf("Draft saved (ID: %s)\n", draftID)
				return nil
			}

			// Send the email
			submissionID, err := client.SendEmail(cmd.Context(), opts)
			if err != nil {
				sendErr := cerrors.WithContext(err, "sending email")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(sendErr, cerrors.SuggestionReauth)
				}
				// Check for invalid from address error and provide helpful suggestion
				if jmap.IsInvalidFromAddressError(err) {
					return cerrors.WithSuggestion(sendErr, cerrors.SuggestionListIdentity)
				}
				return sendErr
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, map[string]any{
					"submissionId": submissionID,
					"status":       "sent",
				})
			}

			fmt.Printf("Email sent successfully (submission ID: %s)\n", submissionID)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&to, "to", nil, "Recipient email addresses")
	cmd.Flags().StringSliceVar(&cc, "cc", nil, "CC email addresses")
	cmd.Flags().StringSliceVar(&bcc, "bcc", nil, "BCC email addresses")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&body, "body", "", "Email body (plain text)")
	cmd.Flags().StringVar(&htmlBody, "html", "", "Email body (HTML)")
	cmd.Flags().StringVar(&fromIdentity, "from", "", "Send from this identity email (see: fastmail email identities)")
	cmd.Flags().BoolVar(&draft, "draft", false, "Save as draft instead of sending")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Email ID to reply to (threads the draft)")
	cmd.Flags().StringSliceVar(&attachments, "attach", nil, "Attach files (path or path:name)")

	return cmd
}
