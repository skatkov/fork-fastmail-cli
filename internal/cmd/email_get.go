package cmd

import (
	"fmt"
	"strings"

	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
	"github.com/salmonumbrella/fastmail-cli/internal/format"
	"github.com/spf13/cobra"
)

func newEmailGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <emailId>",
		Short: "Get email by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(flags)
			if err != nil {
				return err
			}

			email, err := client.GetEmailByID(cmd.Context(), args[0])
			if err != nil {
				getErr := cerrors.WithContext(err, "fetching email")
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
					return cerrors.WithSuggestion(getErr, cerrors.SuggestionReauth)
				}
				return getErr
			}

			if isJSON(cmd.Context()) {
				return printJSON(cmd, emailToOutput(*email))
			}

			// Text output
			fmt.Printf("ID:        %s\n", email.ID)
			fmt.Printf("Subject:   %s\n", email.Subject)
			fmt.Printf("From:      %s\n", format.FormatEmailAddressList(email.From))
			fmt.Printf("To:        %s\n", format.FormatEmailAddressList(email.To))
			if len(email.CC) > 0 {
				fmt.Printf("CC:        %s\n", format.FormatEmailAddressList(email.CC))
			}
			fmt.Printf("Date:      %s\n", email.ReceivedAt)
			fmt.Printf("Thread ID: %s\n", email.ThreadID)
			fmt.Printf("Attachments: %d\n", len(email.Attachments))
			fmt.Println()

			// Print body
			if len(email.TextBody) > 0 && len(email.BodyValues) > 0 {
				for _, part := range email.TextBody {
					if body, ok := email.BodyValues[part.PartID]; ok {
						fmt.Println(body.Value)
					}
				}
			} else if email.Preview != "" {
				fmt.Println(email.Preview)
			}

			return nil
		},
	}

	return cmd
}
