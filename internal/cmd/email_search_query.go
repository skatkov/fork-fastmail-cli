package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/salmonumbrella/fastmail-cli/internal/dateparse"
)

var emailDateTokenRE = regexp.MustCompile(`(?i)\b(after|before|on):(?:"([^"]+)"|'([^']+)'|([^\s]+))`)

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
