package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/fastmail-cli/internal/jmap"
	"github.com/salmonumbrella/fastmail-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

func TestPrintMaskedDryRunSingle_JSON(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), outputModeKey, outfmt.JSON)
	cmd.SetContext(ctx)

	out := captureStdout(t, func() {
		err := printMaskedDryRunSingle(cmd, "a@example.com", jmap.MaskedEmailDisabled, jmap.MaskedEmailEnabled)
		if err != nil {
			t.Fatalf("printMaskedDryRunSingle error: %v", err)
		}
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if payload["dry_run"] != true {
		t.Fatalf("expected dry_run true, got %v", payload["dry_run"])
	}
	if payload["email"] != "a@example.com" {
		t.Fatalf("expected email, got %v", payload["email"])
	}
	if payload["current_state"] != string(jmap.MaskedEmailDisabled) {
		t.Fatalf("expected current_state disabled, got %v", payload["current_state"])
	}
	if payload["new_state"] != string(jmap.MaskedEmailEnabled) {
		t.Fatalf("expected new_state enabled, got %v", payload["new_state"])
	}
}

func TestPrintMaskedDryRunBulk_JSON(t *testing.T) {
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), outputModeKey, outfmt.JSON)
	cmd.SetContext(ctx)

	toUpdate := []jmap.MaskedEmail{
		{Email: "a@example.com", State: jmap.MaskedEmailDisabled},
		{Email: "b@example.com", State: jmap.MaskedEmailDisabled},
	}

	out := captureStdout(t, func() {
		err := printMaskedDryRunBulk(cmd, "example.com", jmap.MaskedEmailEnabled, toUpdate)
		if err != nil {
			t.Fatalf("printMaskedDryRunBulk error: %v", err)
		}
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if payload["dry_run"] != true {
		t.Fatalf("expected dry_run true, got %v", payload["dry_run"])
	}
	if payload["domain"] != "example.com" {
		t.Fatalf("expected domain, got %v", payload["domain"])
	}
	if payload["count"] != float64(2) {
		t.Fatalf("expected count 2, got %v", payload["count"])
	}
	aliases, ok := payload["aliases"].([]any)
	if !ok || len(aliases) != 2 {
		t.Fatalf("unexpected aliases payload: %#v", payload["aliases"])
	}
}
