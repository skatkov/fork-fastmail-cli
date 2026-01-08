# Fastmail CLI Refactor Plan

## Goals
- Reduce `internal/cmd/email.go` size and improve command-focused readability.
- Centralize repeated output, prompt, and formatting helpers.
- Keep behavior stable; only mechanical refactors.

## Steps
1. Baseline review
   - Identify duplicated output/prompt/format patterns across commands.
   - Decide target helper modules and APIs.

2. Extract shared helpers
   - Output helpers (JSON/text switching, tabwriter setup, common table rendering utilities).
   - Prompt helpers (confirmation + dry-run patterns).
   - Format helpers (size/date/address formatting used across email/files/quota).

3. Split email commands
   - Move subcommands into focused files (list/search/get/send/move/mark/delete/bulk/attachments/mailboxes/identities).
   - Keep `newEmailCmd` as assembly point.

4. Update call sites
   - Replace local helpers with shared helpers.
   - Ensure all commands compile with the new helpers.

5. Validate
   - Run relevant tests if available (or at least `go test ./internal/cmd` if fast).

## Deliverables
- New helper files under `internal/cmd` and/or `internal/outfmt`.
- `internal/cmd/email.go` slimmed to command wiring.
- No behavior changes.
