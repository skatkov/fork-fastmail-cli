package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExecute_JSONErrorsAreStructuredAndStdoutIsClean(t *testing.T) {
	t.Setenv("FASTMAIL_OUTPUT", "text") // ensure default doesn't affect test

	stdout := captureStdout(t, func() {
		stderr := captureStderr(t, func() {
			err := Execute([]string{"--output=json", "email", "search"})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		// Stderr should be a single JSON document.
		var payload map[string]any
		if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
			t.Fatalf("stderr is not valid JSON: %v; stderr=%q", err, stderr)
		}

		errObj, ok := payload["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected payload.error object, got: %T (%v)", payload["error"], payload["error"])
		}
		msg, _ := errObj["message"].(string)
		if msg == "" || !strings.Contains(msg, "accepts 1 arg") {
			t.Fatalf("unexpected error.message: %q", msg)
		}
	})

	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected stdout to be empty for JSON error, got: %q", stdout)
	}
}

func TestExecute_TextErrorsAreNotJSON(t *testing.T) {
	t.Setenv("FASTMAIL_OUTPUT", "text")

	out := captureStderr(t, func() {
		err := Execute([]string{"email", "search"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("expected non-JSON stderr in text mode, got: %q", out)
	}
	if !strings.Contains(out, "Error:") {
		t.Fatalf("expected stderr to contain 'Error:', got: %q", out)
	}
}

func TestRootShortcutsExist(t *testing.T) {
	app := NewApp()
	root := NewRootCmd(app)

	want := map[string]bool{
		"search":    false,
		"list":      false,
		"get":       false,
		"send":      false,
		"thread":    false,
		"mailboxes": false,
	}

	for _, c := range root.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Fatalf("expected root shortcut command %q to exist", name)
		}
	}
}
