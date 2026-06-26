---
status: testing
phase: 04-enforcement
source: [04-VERIFICATION.md]
started: 2026-06-17
updated: 2026-06-17
---

## Current Test

number: 1
name: lefthook validate exits 0
expected: |
  After `make tools` + `lefthook install`, running `lefthook validate` at the
  repo root parses lefthook.yml and exits 0 with no errors.
awaiting: user response

## Tests

### 1. `lefthook validate` exits 0
expected: After `make tools` + `lefthook install`, `lefthook validate` exits 0 (lefthook.yml is well-formed).
result: [pending]

### 2. `golangci-lint config verify` exits 0
expected: After `make tools`, `golangci-lint config verify` accepts `.golangci.yml` (v2 schema, formatters block, depguard/errorlint) and exits 0.
result: [pending]

### 3. Bad commit message rejected through the git hook
expected: With hooks installed, `git commit -m "bad message"` is rejected by the commit-msg hook (commitlint, non-zero exit) — the commit does not land.
result: [pending]

### 4. Valid Conventional Commit accepted through the git hook
expected: With hooks installed, a message like `docs(04): wire commitlint` passes the commit-msg hook and the commit lands.
result: [pending]

### 5. pre-commit fires on a staged .go file
expected: Staging a `.go` file and committing runs golangci-lint v2 (lint + embedded gofumpt/gci format) via the pre-commit hook across in-workspace modules + the `GOWORK=off` inventory pass.
result: [pending]

### 6. `lefthook run pre-push` runs in-workspace tests; inventory excluded
expected: `lefthook run pre-push` runs `go test` for pkg/analytics/audit only; `inventory` is intentionally NOT tested (documented NOTE, D-03).
result: [pending]

### 7. `buf build` exits 0 on the empty skeleton
expected: After `make tools`, `buf build` succeeds on the skeleton buf.yaml/buf.gen.yaml (no `.proto` yet) and exits 0; buf is wired into no hook.
result: [pending]

## Summary

total: 7
passed: 0
issues: 0
pending: 7
skipped: 0
blocked: 0

## Gaps

None — all automated checks passed (14/14 must-haves). These 7 items verify live
git-hook firing, which requires the one-time tooling bootstrap (`make tools` +
`lefthook install`; `npm install` already done) that could not be performed in the
autonomous run because the Go tool binaries (golangci-lint, lefthook, buf) are not
installed on this machine. They are honest no-phantom carry-forwards, not defects.
