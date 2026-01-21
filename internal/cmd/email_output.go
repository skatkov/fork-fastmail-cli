package cmd

import "github.com/salmonumbrella/fastmail-cli/internal/jmap"

// EmailOutput is a flattened representation of Email for agent-friendly JSON output.
// It includes computed fields like fromEmail and isUnread that are easier to parse.
type EmailOutput struct {
	ID            string              `json:"id"`
	Subject       string              `json:"subject"`
	From          []jmap.EmailAddress `json:"from,omitempty"`
	FromEmail     string              `json:"fromEmail,omitempty"`
	FromName      string              `json:"fromName,omitempty"`
	To            []jmap.EmailAddress `json:"to,omitempty"`
	ToEmail       string              `json:"toEmail,omitempty"`
	CC            []jmap.EmailAddress `json:"cc,omitempty"`
	ReceivedAt    string              `json:"receivedAt"`
	Preview       string              `json:"preview,omitempty"`
	HasAttachment bool                `json:"hasAttachment"`
	IsUnread      bool                `json:"isUnread"`
	ThreadID      string              `json:"threadId,omitempty"`
	Keywords      map[string]bool     `json:"keywords,omitempty"`
	MessageCount  int                 `json:"messageCount,omitempty"` // Count of messages in thread
}

// emailToOutput converts an Email to a flattened EmailOutput for JSON serialization.
func emailToOutput(e jmap.Email) EmailOutput {
	out := EmailOutput{
		ID:            e.ID,
		Subject:       e.Subject,
		From:          e.From,
		To:            e.To,
		CC:            e.CC,
		ReceivedAt:    e.ReceivedAt,
		Preview:       e.Preview,
		HasAttachment: e.HasAttachment,
		ThreadID:      e.ThreadID,
		Keywords:      e.Keywords,
	}
	// Flatten from address
	if len(e.From) > 0 {
		out.FromEmail = e.From[0].Email
		out.FromName = e.From[0].Name
	}
	// Flatten to address
	if len(e.To) > 0 {
		out.ToEmail = e.To[0].Email
	}
	// Compute isUnread from keywords (unread = $seen not present or false)
	out.IsUnread = e.Keywords == nil || !e.Keywords["$seen"]
	return out
}

// emailsToOutput converts a slice of emails to flattened output format.
func emailsToOutput(emails []jmap.Email) []EmailOutput {
	out := make([]EmailOutput, len(emails))
	for i, e := range emails {
		out[i] = emailToOutput(e)
	}
	return out
}

// emailsToOutputWithCounts converts emails to output format with thread message counts.
func emailsToOutputWithCounts(emails []jmap.Email, threadCounts map[string]int) []EmailOutput {
	out := make([]EmailOutput, len(emails))
	for i, email := range emails {
		out[i] = emailToOutput(email)
		if count, ok := threadCounts[email.ThreadID]; ok {
			out[i].MessageCount = count
		}
	}
	return out
}
