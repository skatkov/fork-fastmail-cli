package cmd

import "testing"

func TestSanitizeTab(t *testing.T) {
	if got := sanitizeTab("a\tb\tc"); got != "a b c" {
		t.Fatalf("sanitizeTab() = %q, want %q", got, "a b c")
	}
	if got := sanitizeTab("no tabs"); got != "no tabs" {
		t.Fatalf("sanitizeTab() = %q, want %q", got, "no tabs")
	}
}
