# Agent-Friendly CLI Refactor Plan (fastmail-cli)

Goal: minimize agent friction cost (retries, navigation, ambiguity) while keeping the human CLI usable.

## Principles (compressed)

- Prefer “desire paths”: implement the commands/aliases an agent naturally tries.
- Keep JSON mode machine-stable:
  - stdout is JSON only
  - errors are structured
  - include IDs and follow-up metadata in responses
- Flatten navigation: accept the resource identifier the agent already has, do lookups internally.
- Non-interactive by default when asked: a single global `--yes/-y` should remove prompts.

## Work Items

### 1. Global Non-Interactive Switch

- [x] Add persistent `--yes/-y` to skip confirmations.
- [x] Remove per-command `--yes/-y` flags (they shadow the persistent flag).
- [x] Route all confirmations through `app.Confirm(...)` (no ad-hoc `confirmPrompt(...)` calls).

Acceptance:
- `fastmail -y <any command that prompts>` never blocks on stdin.

### 2. JSON Mode Contract

- [x] When `--output=json`, print structured errors to stderr (instead of free-form text).
- [ ] Audit commands for “JSON contamination” (any stdout output when in JSON mode).
- [ ] Ensure every mutating command returns a JSON object in JSON mode (not silence).
  - Done for: `files upload|download|mkdir|delete|move`, `contacts delete`, `calendar event-delete`.

Acceptance:
- `FASTMAIL_OUTPUT=json fastmail ...` emits exactly one JSON document on stdout for success cases.

### 3. Desire Paths (Aliases + Shortcuts)

- [x] Add plural/synonym aliases for core nouns (email/emails/messages, calendar/cal, etc).
- [x] Add common verb aliases (`ls`, `rm`, `mv`, etc) where sensible.
- [x] Make `fastmail auth` run the recommended login flow.
- [x] Add top-level email shortcuts:
  - `fastmail search ...` -> `fastmail email search ...`
  - `fastmail ls` / `fastmail list` -> `fastmail email list`
  - `fastmail show/get/cat <id>` -> `fastmail email get <id>`
  - `fastmail send ...` -> `fastmail email send ...`

Acceptance:
- An agent guessing “common Unix-ish” commands succeeds without reading `--help`.

### 4. “Flat Access” Helpers (Reduce Navigation)

- [ ] Where an operation currently requires “find parent IDs then act”, add resolver logic:
  - mailbox: accept ID or name (already exists in several places)
  - files: accept path; return normalized remote path + local path in JSON
  - future: accept full URLs and normalize IDs (email IDs, thread IDs, etc)

Acceptance:
- Agent can act directly on the identifier it has, without manual multi-step discovery.

### 5. Tests (Interface-Level)

- [ ] Add CLI-level tests for the desire paths and the JSON contract (stdout-only JSON).
- [ ] Add a regression test for `email download --all` in JSON mode.

## Implementation Order

1. Fix `--yes` shadowing by removing command-local `--yes` flags and refactoring to `app.Confirm`.
2. Add missing JSON outputs for mutating commands (start with `files *`, `contacts delete`, `calendar event-delete`).
3. Add root-level shortcut commands for high-frequency email actions.
4. Add CLI contract tests.
