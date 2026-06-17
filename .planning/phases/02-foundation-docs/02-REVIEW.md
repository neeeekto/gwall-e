---
phase: 02-foundation-docs
reviewed: 2026-06-17T00:00:00Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - knowledge/README.md
  - knowledge/boundaries.md
  - knowledge/build.md
  - knowledge/git.md
  - knowledge/structure.md
findings:
  critical: 1
  warning: 2
  info: 2
  total: 5
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-06-17
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Documentation-only phase: five Russian-language knowledge-base docs establishing project
conventions (README index, boundaries, build/test, git, structure). I verified each
factual claim against the live repository: `go.work` membership, module paths, Go version,
directory layout, relative links, and — critically — the actual build/test commands.

Overall the docs are well-structured, internally consistent, and free of leaked
secrets/IPs/internal hostnames. All relative markdown links resolve to existing files.
Workspace membership (`pkg`, `services/analytics`, `services/audit`), the Go 1.24.6
requirement, the `inventory` WIP-outside-workspace status, the empty `inventory/internal/*`
scaffolds, the `gateway`/`outgate` README+go.mod-only stubs, and the `analytics`
"no packages yet" claim all match reality.

However, the central promise of `build.md` — "реально проверенные рецепты" (only
verified, runnable recipes), with an explicit **WON'T** against documenting commands that
do not run — is violated: **two of the documented build commands fail with exit code 1**.
This is the headline finding. The `git.md` remote and `origin/HEAD → main` claims are
accurate.

## Critical Issues

### CR-01: Documented `go build ./...` commands fail (exit 1) — phantom recipes

**File:** `knowledge/build.md:26-27` (audit) and `knowledge/build.md:38-40` (inventory)

**Issue:** `build.md` states it documents "реально проверенные рецепты" and explicitly
forbids phantom commands ("**WON'T** документировать команды, которые не запускаются").
Two documented recipes do not run:

- `cd services/audit && go build ./...` — claimed "сборка проходит" (line 26-27).
  Actual result:
  ```
  go: build output "cmd" already exists and is a directory
  exit status 1
  ```
- `cd services/inventory && GOWORK=off go build ./...` (line 40) — same failure mode
  (`go: build output "cmd" ... exit status 1`).

Root cause: each module's only package lives in a directory named `cmd`. `go build ./...`
with a single resolved package named `cmd` tries to emit a binary file `cmd` into the cwd,
which collides with the existing `cmd/` directory. This is deterministic, not a
WIP/flakiness artifact — it fails every time on a clean tree.

Note: `inventory`'s failure is independently expected (build.md correctly frames inventory
as WIP that "не гарантированно компилируется"), so the WIP framing partially covers it.
But for **audit** the doc makes a hard, falsifiable claim ("модуль содержит пакеты ... сборка
проходит") that is wrong. This is the load-bearing example of the whole "build modules"
section, so it must be correct.

**Fix:** Replace the failing invocations with verified ones. Confirmed-working
alternatives on this tree (all exit 0):

```
# audit — verify compilation without emitting a colliding binary:
cd services/audit && go build ./cmd          # exit 0
# or, to compile all packages without an output collision:
cd services/audit && go vet ./...            # exit 0
# or send the binary elsewhere:
cd services/audit && go build -o /tmp/audit ./...

# inventory (WIP, GOWORK=off) — same adjustment:
cd services/inventory && GOWORK=off go build ./cmd
# or: cd services/inventory && GOWORK=off go vet ./...
```

Update the prose accordingly so the "сборка проходит" claim matches a command that
actually exits 0. Re-verify after editing.

## Warnings

### WR-01: README index omits the root-level Go module from the repo map

**File:** `knowledge/structure.md:17-25`, `knowledge/README.md:10-19`

**Issue:** The repository root contains a real Go module — root `go.mod` declares
`module github.com/gwall-e` (plus `go.sum`) — that is not part of `go.work` and is not
mentioned anywhere in `structure.md`'s module/workspace map. `structure.md` positions
itself as the canon for "что собирается вместе, а что живёт отдельно" and lists exactly
three workspace modules, but a reader navigating the repo will hit a fourth, unexplained
`go.mod` at the root. Leaving it undocumented invites exactly the "is this stale? do I
fix it?" confusion the boundaries doc tries to prevent.

**Fix:** Add one line to `structure.md` clarifying the root module's role (e.g., tooling /
aggregator / stale), or — if it is a stale scaffold — note it under the
boundaries/stale-files rule. At minimum acknowledge its existence so the "exactly three
modules" framing is not contradicted by an unmentioned root `go.mod`.

### WR-02: "audit содержит пакеты (есть `cmd/main.go`)" overstates the module's contents

**File:** `knowledge/build.md:30-31`

**Issue:** The doc presents `services/audit` as the positive counter-example to the
"analytics is an empty scaffold" case, implying it is a real buildable service. In fact
`services/audit` contains only `go.mod`, `README.md`, and a single `cmd/main.go` — there
is no `internal/` or domain code, and (per CR-01) `go build ./...` does not even succeed.
The contrast drawn against `analytics` is weaker and more fragile than the prose suggests,
and combined with CR-01 it risks misleading a reader into treating audit as a fully wired
service.

**Fix:** Soften the claim to match reality (e.g., "audit содержит лишь `cmd/main.go` —
точку входа-заготовку; сборка точки входа: `go build ./cmd`"), and ensure it pairs with
the corrected command from CR-01.

## Info

### IN-01: Commit-message example uses a scope form the doc later restricts

**File:** `knowledge/git.md:23` vs `knowledge/git.md:57`

**Issue:** The Conventional Commits example shows `docs(02-01): ...` (a sub-phase scope),
while the GSD section (line 57) prescribes phase scopes of the form `docs(NN):` (e.g.
`docs(02):`). The two scope shapes (`02-01` vs `02`) are not reconciled, which could
confuse a reader about the canonical scope format.

**Fix:** Either use a matching `docs(02): ...` example in line 23, or add a half-sentence
noting that sub-phase scopes (`NN-MM`) are also acceptable in the GSD flow.

### IN-02: "snesённые/пустые stale-леса" wording for inventory/internal could be tightened

**File:** `knowledge/structure.md:45-48`

**Issue:** All six `services/inventory/internal/*` directories (`repository`, `app`, `api`,
`usecase`, `domain`, `cron`) are present but empty — confirmed. The phrase "снесённые/пустые
stale-леса, оставшиеся от старого кода" is accurate, but "снесённые" (torn down) plus
"оставшиеся от старого кода" (left over from old code) is slightly contradictory phrasing
for what are simply empty placeholder directories. Minor clarity nit, not a factual error.

**Fix:** Optional — simplify to "пустые каталоги-заглушки `internal/*`" to avoid the
torn-down/leftover ambiguity.

---

_Reviewed: 2026-06-17_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
