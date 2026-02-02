package cmd

import (
	"testing"
	"time"
)

func TestNormalizeEmailSearchQuery(t *testing.T) {
	now := time.Date(2026, 1, 28, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "relative yesterday",
			in:   "subject:meeting after:yesterday",
			want: "subject:meeting after:2026-01-27",
		},
		{
			name: "relative with quotes",
			in:   "after:\"2h ago\" before:today",
			want: "after:2026-01-28 before:2026-01-28",
		},
		{
			name: "absolute date",
			in:   "after:2026-01-01",
			want: "after:2026-01-01",
		},
		{
			name: "rfc3339 passthrough",
			in:   "after:2026-01-01T10:00:00Z",
			want: "after:2026-01-01T10:00:00Z",
		},
		{
			name: "case preserved",
			in:   "After:yesterday",
			want: "After:2026-01-27",
		},
		{
			name: "no date tokens",
			in:   "from:alice@example.com",
			want: "from:alice@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEmailSearchQuery(tt.in, now)
			if err != nil {
				t.Fatalf("normalizeEmailSearchQuery error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeEmailSearchQuery(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeEmailSearchQuery_Invalid(t *testing.T) {
	now := time.Date(2026, 1, 28, 15, 4, 5, 0, time.UTC)
	if _, err := normalizeEmailSearchQuery("after:not-a-date", now); err == nil {
		t.Fatalf("expected error for invalid date token")
	}
}
