---
phase: 04-enforcement
verified: 2026-06-17T15:44:34Z
status: human_needed
score: 14/14
overrides_applied: 0
human_verification:
  - test: "Run `make tools` then `lefthook validate`"
    expected: "lefthook exits 0 — config schema is valid"
    why_human: "lefthook binary absent on verification machine; requires `make tools` to install v2.1.9"
  - test: "Run `make tools` then `golangci-lint config verify`"
    expected: "exits 0 — .golangci.yml is a valid golangci-lint v2 schema"
    why_human: "golangci-lint binary absent; structural YAML validation passed but config verify requires the tool"
  - test: "Run `make tools` + `npm install` + `lefthook install`, then `git commit --allow-empty -m 'bad message'`"
    expected: "commit-msg hook fires and rejects the commit (commitlint error shown)"
    why_human: "lefthook install not run on this machine; hook firing through git requires activated .git/hooks"
  - test: "Stage a .go file under pkg/ with a bad import order, then `git add` and attempt commit"
    expected: "pre-commit hook runs golangci-lint; gci auto-corrects import order and re-stages via stage_fixed, or lint blocks the commit"
    why_human: "golangci-lint absent; live pre-commit firing requires make tools + lefthook install"
  - test: "Run `lefthook run pre-push` after bootstrap"
    expected: "`cd pkg && go test ./...` runs green; no inventory test runs"
    why_human: "lefthook absent; requires make tools + lefthook install to activate"
  - test: "Run `make tools` then `buf build` from repo root"
    expected: "exits 0 — empty proto module produces no output but is not an error"
    why_human: "buf binary absent; buf build must be verified after make tools installs buf v1.71.0"
---

# Phase 4: Enforcement-слой (тулинг) — Verification Report

**Phase Goal:** Механизируемые правила базы знаний подкреплены тулингом, а каждое правило помечено статусом enforcement — база перестаёт быть только декларативной.

**Verified:** 2026-06-17T15:44:34Z
**Status:** human_needed
**Re-verification:** No — initial verification

All 14 automated must-haves verified. 6 human verification items remain (live hook firing after `make tools` + `lefthook install`). Accepting tools absent on machine as valid per the no-phantom rule — the phase correctly deferred live-tool checks to the bootstrap step.

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `.golangci.yml` exists with v2 schema (`version: "2"`, `linters.default: standard`) | VERIFIED | File exists; `version: "2"` and `default: standard` confirmed |
| 2 | gofumpt + gci in `formatters` block (top-level, NOT as standalone linters) | VERIFIED | `formatters:` top-level block with `enable: [gofumpt, gci]` present; no standalone gofumpt install in Makefile |
| 3 | depguard `no-cqrs-bus` ban on `github.com/gwall-e/pkg/mediatr` is biting (`$all`) | VERIFIED | Rule present with `list-mode: lax`, `files: ["$all"]`, deny `github.com/gwall-e/pkg/mediatr` |
| 4 | depguard layer-direction rule is explicitly DORMANT (no matching files yet) | VERIFIED | `domain-imports-inward-only` rule documented "DORMANT — matches nothing today" in desc |
| 5 | errorlint enabled with `errorf`, `asserts`, `comparison: true` | VERIFIED | Present under `linters.settings.errorlint` |
| 6 | `Makefile` has `make tools` pinning golangci v2.12.2 / lefthook v2.1.9 / buf v1.71.0 at exact versions (no `@latest`) | VERIFIED | `GOLANGCI_VERSION := v2.12.2`, `LEFTHOOK_VERSION := v2.1.9`, `BUF_VERSION := v1.71.0`; all `go install` lines use pinned versions via `/v2/` paths |
| 7 | `lefthook.yml` exists with pre-commit (lint+format), pre-push (in-ws tests), commit-msg (commitlint); buf in NO hook | VERIFIED | All three hook groups present; `golangci-lint run` in pre-commit; `npx --no-install commitlint --edit {1}` in commit-msg; no buf run command in any hook |
| 8 | `inventory` excluded from pre-push with explicit intentional NOTE | VERIFIED | NOTE comment explicitly states "INTENTIONALLY NOT tested ... deliberate, NOT an oversight (D-03)" |
| 9 | `package.json` is `private: true`, exact-pinned commitlint 21.0.2, NO other deps | VERIFIED | `"private": true`, `@commitlint/cli: "21.0.2"`, `@commitlint/config-conventional: "21.0.2"`, no caret/tilde, no other devDeps |
| 10 | commitlint round-trip: bad message rejected (exit 1), valid GSD-style accepted (exit 0) | VERIFIED | `printf 'bad message' \| npx --no-install commitlint` → exit 1 (2 errors); `printf 'docs(04): wire commitlint' \| npx --no-install commitlint` → exit 0 |
| 11 | `buf.yaml` + `buf.gen.yaml` v2 skeletons with SKELETON headers; pinned plugins; not hooked | VERIFIED | Both files: `version: v2`, SKELETON header comments, pinned `protocolbuffers/go:v1.36.5` + `grpc/go:v1.5.1`; buf wired into no hook |
| 12 | `grep -rn 'planned:' knowledge/` returns nothing — all Phase-3 forward marks flipped | VERIFIED | Command returns exit 1 (no matches) across all 10 knowledge files |
| 13 | No rule STATUS labeled `CI-gated`; `CI-gated` confined to authoring.md legend + architecture.md forward note + boundaries.md taxonomy row | VERIFIED | `CI-gated` appears in: authoring.md (legend definition), architecture.md L27 ("становится CI-gated при появлении CI" — a forward note, not a rule status), boundaries.md (taxonomy reference row). No rule carries `⟶ CI-gated` as its enforcement status. |
| 14 | `knowledge/build.md` documents `make tools` + `npm install` + `lefthook install` bootstrap (D-11) | VERIFIED | §Bootstrap тулинга section present with all three steps; pointer-over-copy to Makefile for versions |

**Score:** 14/14 truths verified

---

## Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `.golangci.yml` | v2 schema, errorlint, depguard, gofumpt+gci formatters | VERIFIED | v2, `linters.default: standard`, errorlint+depguard enabled, `formatters` top-level block |
| `Makefile` | `make tools` with pinned versions, `/v2/` paths | VERIFIED | 3 exact-pinned installs; echo mentions npm install + lefthook install next steps |
| `lefthook.yml` | pre-commit/pre-push/commit-msg; no buf | VERIFIED | All 3 hooks present; inventory NOTE; `npx --no-install`; no buf command |
| `package.json` | private, exact-pinned commitlint only | VERIFIED | `private: true`, 21.0.2 exact pins, no caret/tilde, single devDep |
| `commitlint.config.mjs` | extends config-conventional, ESM | VERIFIED | `.mjs` forces ESM; `export default { extends: ['@commitlint/config-conventional'] }` |
| `buf.yaml` | v2, lint STANDARD, breaking FILE, SKELETON | VERIFIED | All required fields; SKELETON header |
| `buf.gen.yaml` | v2, pinned remote plugins, SKELETON | VERIFIED | Both plugins exact-pinned; SKELETON header |
| `knowledge/authoring.md` | 3-status legend canon with MUST NOT CI-gated rule | VERIFIED | §Статус enforcement rewritten; 3 statuses defined; MUST NOT CI-gated without CI; biting/dormant subnotes |
| `knowledge/style.md` | gofumpt→hook(format), %w→hook(errorlint), DTO→hook(depguard,dormant), typed-IDs→convention-only | VERIFIED | All 4 marks confirmed flipped |
| `knowledge/architecture.md` | 2 layer marks→hook(depguard,dormant), 2 bans→hook(depguard,biting) | VERIFIED | L27+L30 dormant; L152+L155 biting |
| `knowledge/testing.md` | mockery→convention-only; no phantom go:generate claim | VERIFIED | `convention-only (review-enforced) — mockery-обвязка отложена` |
| `knowledge/build.md` | make tools + npm install + lefthook install bootstrap | VERIFIED | §Bootstrap тулинга section with all 3 steps |
| `knowledge/boundaries.md` | ownership map rows: enforcement taxonomy + bootstrap + buf-skeleton | VERIFIED | Rows present; honest buf-skeleton note |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `.golangci.yml` | `github.com/gwall-e/pkg/mediatr` | depguard no-cqrs-bus deny | VERIFIED | `deny: pkg: github.com/gwall-e/pkg/mediatr` in biting rule |
| `Makefile` | `github.com/golangci/golangci-lint/v2/cmd/golangci-lint` | `go install` pinned | VERIFIED | `/v2/` path confirmed in Makefile |
| `lefthook.yml` | `.golangci.yml` | pre-commit `golangci-lint run` | VERIFIED | `golangci-lint run ./...` in lint-workspace and lint-inventory commands |
| `lefthook.yml` | `commitlint.config.mjs` | commit-msg `npx --no-install commitlint --edit {1}` | VERIFIED | Exact command present; `--no-install` confirmed |
| `lefthook.yml` | pkg ginkgo suite | pre-push `cd pkg && go test` | VERIFIED | `run: cd pkg && go test ./...` confirmed |
| `knowledge/style.md` | `.golangci.yml` errorlint | ENF-05 mark flip %w→hook(lint: errorlint) | VERIFIED | `⟶ hook (lint: errorlint)` at L74 |
| `knowledge/architecture.md` | `.golangci.yml` depguard no-cqrs-bus | ENF-05 mark flip→hook(lint: depguard, biting) | VERIFIED | `⟶ hook (lint: depguard, biting)` at L152+L155 |
| `knowledge/build.md` | `Makefile` make tools | bootstrap recipe cross-ref | VERIFIED | "Версии — канон в корневом `Makefile`" pointer-over-copy |

---

## Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| commitlint rejects bad message | `printf 'bad message' \| npx --no-install commitlint` | exit 1, "type may not be empty" | PASS |
| commitlint accepts GSD-style message | `printf 'docs(04): wire commitlint' \| npx --no-install commitlint` | exit 0 | PASS |
| `golangci-lint config verify` | Not runnable (tool absent) | golangci-lint not installed | SKIP — human verify |
| `buf build` | Not runnable (tool absent) | buf not installed | SKIP — human verify |
| `lefthook validate` | Not runnable (tool absent) | lefthook not installed | SKIP — human verify |
| Git commits exist for all deliverables | `git log --oneline` | 7 commits verified: 81a0239, 864127c, 88aa153, 393853e, b0ffc76, 5bf7c89, 9868758 | PASS |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|---------|
| ENF-01 | 04-01 | `.golangci.yml` golangci-lint v2 + gofumpt + gci | SATISFIED | File exists with v2 schema, formatters block, errorlint, depguard |
| ENF-02 | 04-03 | `lefthook.yml` — pre-commit / pre-push / commit-msg | SATISFIED | All 3 hooks present and correctly wired; live firing is human_needed |
| ENF-03 | 04-02, 04-03 | commitlint config wired to commit-msg | SATISFIED | package.json + commitlint.config.mjs exist; commit-msg hook wires `npx --no-install commitlint --edit {1}`; round-trip verified |
| ENF-04 | 04-02 | buf.yaml + buf.gen.yaml v2 skeleton | SATISFIED | Both files v2, SKELETON headers, pinned plugins, not wired into any hook |
| ENF-05 | 04-04 | Every mechanizable rule in knowledge/ has factual enforcement status | SATISFIED | `grep -rn 'planned:' knowledge/` → empty; all marks verified |

**Note on ENF-03 checkbox:** REQUIREMENTS.md v1 list shows `[ ] ENF-03` (unchecked) while the traceability table correctly reads "Config delivered (04-02); commit-msg wiring 04-03". This is a documentation cosmetic gap — the actual deliverables (package.json, commitlint.config.mjs, commit-msg hook wiring) all exist and the round-trip works. The checkbox should be updated to `[x]`.

---

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `knowledge/README.md` traceability in REQUIREMENTS.md | 37 | `[ ] ENF-03` checkbox unchecked despite completion | Info | Documentation cosmetic only; traceability row confirms complete; actual artifacts verified |

No TBD/FIXME/XXX debt markers found in any phase-modified file.
No placeholder/stub implementations found.
No phantom claims found — honest carry-forwards documented (tools absent → `golangci-lint config verify` and `buf build` explicitly deferred to post-`make tools`).

---

## Human Verification Required

### 1. `lefthook validate` after `make tools`

**Test:** `make tools && lefthook validate`
**Expected:** lefthook exits 0 — `lefthook.yml` schema is valid
**Why human:** lefthook binary absent on verification machine; installed by `make tools`

### 2. `golangci-lint config verify` after `make tools`

**Test:** `make tools && golangci-lint config verify`
**Expected:** exits 0 — `.golangci.yml` is a valid golangci-lint v2 schema
**Why human:** golangci-lint binary absent; structural YAML validation passed but config verify requires the actual binary

### 3. commit-msg hook fires on bad message

**Test:** After `make tools && npm install && lefthook install`: `git commit --allow-empty -m "bad message"`
**Expected:** commit-msg hook activates and rejects the commit (commitlint error); commit does not land
**Why human:** lefthook install not run; hook firing through git requires activated `.git/hooks/`

### 4. commit-msg hook accepts valid GSD-style message

**Test:** `git commit --allow-empty -m "docs(04): smoke test hooks"`
**Expected:** commit-msg hook runs and accepts the message; commit lands
**Why human:** Same bootstrap requirement as above

### 5. pre-commit hook fires on staged .go file

**Test:** Stage a .go file with bad import order under `pkg/`, attempt commit
**Expected:** pre-commit runs golangci-lint; gci auto-corrects import order and re-stages via `stage_fixed`, or lint blocks commit
**Why human:** golangci-lint absent; requires make tools + lefthook install

### 6. pre-push runs in-workspace tests; inventory excluded

**Test:** `lefthook run pre-push`
**Expected:** `cd pkg && go test ./...` runs green; inventory test does NOT run
**Why human:** lefthook absent; requires make tools + lefthook install

### 7. `buf build` on empty skeleton

**Test:** `make tools && buf build`
**Expected:** exits 0 — empty proto module produces no output, not an error
**Why human:** buf binary absent; installed by `make tools`

---

## Gaps Summary

No gaps found. All 14 must-haves are verified. The 7 human verification items above are carry-forwards from the no-phantom policy: the enforcement tools (golangci-lint, lefthook, buf) were absent on the executor machine during plan runs, and the phase correctly documented this rather than fabricating tool-run results. Live hook firing requires `make tools` + `npm install` + `lefthook install` (one-time bootstrap).

The ENF-03 checkbox discrepancy in REQUIREMENTS.md is a cosmetic documentation issue, not a functional gap — the deliverables exist and the commitlint round-trip passes.

---

_Verified: 2026-06-17T15:44:34Z_
_Verifier: Claude (gsd-verifier)_
