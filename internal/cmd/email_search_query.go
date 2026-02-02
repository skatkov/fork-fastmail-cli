package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/salmonumbrella/fastmail-cli/internal/dateparse"
)

var emailDateTokenRE = regexp.MustCompile(`(?i)\b(after|before|on):(?:"([^"]+)"|'([^']+)'|([^\s]+))`)

// emailSearchFilter holds parsed email search query components.
// This is an internal type - use jmap.EmailSearchFilter for JMAP operations.
type emailSearchFilter struct {
	Text   string // Remaining text for full-text search
	After  string // RFC3339 timestamp for "after" filter
	Before string // RFC3339 timestamp for "before" filter
}

// parseEmailSearchFilter parses a query string into JMAP filter components.
// It extracts after: and before: date tokens as proper JMAP filter properties,
// and leaves remaining text as the "text" filter for full-text search.
func parseEmailSearchFilter(query string, now time.Time) (*emailSearchFilter, error) {
	filter := &emailSearchFilter{}

	matches := emailDateTokenRE.FindAllStringSubmatchIndex(query, -1)
	if len(matches) == 0 {
		filter.Text = strings.TrimSpace(query)
		return filter, nil
	}

	// Build the remaining text (everything except after:/before: tokens)
	// and extract date filters
	var textParts strings.Builder
	last := 0

	for _, match := range matches {
		// Add text before this match
		textParts.WriteString(query[last:match[0]])

		key := strings.ToLower(query[match[2]:match[3]])
		value := extractEmailDateValue(query, match)

		// Skip "on:" for now - leave it in text search
		if key == "on" {
			textParts.WriteString(query[match[0]:match[1]])
			last = match[1]
			continue
		}

		// Parse and convert date to RFC3339 for JMAP
		timestamp, err := parseDateToRFC3339(value, now)
		if err != nil {
			return nil, fmt.Errorf("invalid %s date %q (use RFC3339, YYYY-MM-DD, or relative like yesterday, 2h ago, monday)", key, value)
		}

		switch key {
		case "after":
			filter.After = timestamp
		case "before":
			filter.Before = timestamp
		}

		last = match[1]
	}

	// Add remaining text after last match
	textParts.WriteString(query[last:])

	// Clean up the remaining text (collapse multiple spaces, trim)
	remaining := strings.TrimSpace(textParts.String())
	remaining = collapseSpaces(remaining)
	filter.Text = remaining

	return filter, nil
}

// parseDateToRFC3339 converts a date value to RFC3339 format for JMAP.
// For date-only values (YYYY-MM-DD or relative like "yesterday"), it returns
// the start of that day in UTC (e.g., "2026-01-15T00:00:00Z").
func parseDateToRFC3339(value string, now time.Time) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("empty date")
	}

	// If already RFC3339, use as-is
	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}

	// If YYYY-MM-DD, convert to RFC3339 at start of day UTC
	if t, err := time.Parse("2006-01-02", trimmed); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}

	// Parse relative date and convert to RFC3339
	t, err := dateparse.ParseDateTime(trimmed, now)
	if err != nil {
		return "", err
	}

	return t.UTC().Format(time.RFC3339), nil
}

// collapseSpaces replaces multiple consecutive spaces with a single space.
func collapseSpaces(s string) string {
	var result strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' {
			if !prevSpace {
				result.WriteRune(r)
			}
			prevSpace = true
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}
	return result.String()
}

func normalizeEmailSearchQuery(query string, now time.Time) (string, error) {
	matches := emailDateTokenRE.FindAllStringSubmatchIndex(query, -1)
	if len(matches) == 0 {
		return query, nil
	}

	var b strings.Builder
	last := 0
	for _, match := range matches {
		b.WriteString(query[last:match[0]])

		key := query[match[2]:match[3]]
		value := extractEmailDateValue(query, match)

		normalized, err := normalizeEmailDateValue(value, now)
		if err != nil {
			return "", fmt.Errorf("invalid %s date %q (use RFC3339, YYYY-MM-DD, or relative like yesterday, 2h ago, monday)", strings.ToLower(key), value)
		}

		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(normalized)
		last = match[1]
	}
	b.WriteString(query[last:])

	return b.String(), nil
}

func extractEmailDateValue(query string, match []int) string {
	if len(match) < 10 {
		return ""
	}
	switch {
	case match[4] != -1:
		return query[match[4]:match[5]]
	case match[6] != -1:
		return query[match[6]:match[7]]
	case match[8] != -1:
		return query[match[8]:match[9]]
	default:
		return ""
	}
}

func normalizeEmailDateValue(value string, now time.Time) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("empty date")
	}

	if _, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return trimmed, nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err == nil {
		return trimmed, nil
	}

	t, err := dateparse.ParseDateTime(trimmed, now)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02"), nil
}
