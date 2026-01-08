package cmd

import (
	"strings"
	"testing"
)

func TestPrintAlready(t *testing.T) {
	out := captureStdout(t, func() {
		printAlready("Already done")
	})

	if strings.TrimSpace(out) != "Already done" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFormatAlready(t *testing.T) {
	msg := formatAlready("Already %s %d", "done", 2)
	if msg != "Already done 2" {
		t.Fatalf("unexpected format: %q", msg)
	}
}
