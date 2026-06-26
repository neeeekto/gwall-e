---
phase: 04-enforcement
plan: 04
subsystem: docs
tags: [enforcement-status, knowledge-base, depguard, errorlint, gofumpt, commitlint, lefthook, bootstrap, no-phantom, ENF-05]

# Dependency graph
requires:
  - phase: 04-enforcement
    plan: 01
    provides: ".golangci.yml (errorlint + depguard biting/dormant + gofumpt/gci formatters) + Makefile make tools — the hook mechanism the marks flip to"
  - phase: 04-enforcement
    plan: 02
    provides: "package.json + commitlint.config.mjs (commit-msg) + buf skeleton — backing the commitlint hook status and the honest buf-skeleton ownership row"
  - phase: 04-enforcement
    plan: 03
    provides: "lefthook.yml — the live pre-commit/pre-push/commit-msg orchestration that makes 'hook' statuses true today"
provides:
  - "ENF-05 complete: every Phase-3 forward enforcement mark in knowledge/*.md flipped to a truthful status (no planned: remains)"
  - "authoring.md §Статус enforcement — single canon legend defining hook/convention-only/CI-gated precisely + MUST NOT mark CI-gated without CI"
  - "build.md §Bootstrap тулинга — one-time make tools / npm install / lefthook install recipe (D-11)"
  - "boundaries.md ownership-map rows registering the enforcement taxonomy + bootstrap + inventory exclusion + honest buf-skeleton note"
affects: [enforcement, knowledge-base, future-CI-epic, layer-code-restoration]

# Tech tracking
tech-stack:
  added: []
  patterns: ["truthful-status reconciliation (hook=committed git hook today, convention-only=review, CI-gated=reserved-no-CI)", "depguard subnotes biting (fires today) vs dormant (forward, matches nothing yet)", "pointer-over-copy: one status canon (authoring.md), topic docs reference it"]

key-files:
  created: [.planning/phases/04-enforcement/04-04-SUMMARY.md]
  modified: [knowledge/authoring.md, knowledge/style.md, knowledge/architecture.md, knowledge/testing.md, knowledge/build.md, knowledge/boundaries.md, knowledge/README.md]

key-decisions:
  - "typed-IDs -> convention-only (A1): no off-the-shelf linter enforces 'use a named type for IDs'; a custom rule is out of scope — honest default over a phantom hook"
  - "mockery -> convention-only (A2): mockery harness (.mockery.yaml/go:generate) deferred; kept the convention, dropped 'planned: Phase 4' phrasing that implied this phase wired it"
  - "CI-gated never applied as a rule status: confined to authoring.md legend (definition) + architecture.md 'becomes CI-gated when CI lands' note + boundaries.md taxonomy row — D-07/D-08, Pitfall 8"
  - "depguard layer marks flipped to hook (lint: depguard, dormant) with the Manual-Only review note refined to state the rule is dormant because no layer code exists yet"
  - "patterns.md left untouched — verified no planned: marks (all already convention-only); no needless churn"

patterns-established:
  - "Forward-then-flip closed out: Phase-3 D-11 forward marks become live statuses in Phase 4 without rewriting any rule body — status text only"
  - "Honest dormancy: a depguard rule written against not-yet-existing layer paths is documented hook (..., dormant), not claimed as active enforcement"

requirements-completed: [ENF-05]

# Metrics
duration: ~5min
completed: 2026-06-17
---

# Phase 4 Plan 04: ENF-05 truthful enforcement-status reconciliation Summary

**Flipped every Phase-3 forward enforcement mark in `knowledge/*.md` to its truthful status now that the 04-01/04-02/04-03 configs exist — gofumpt/errorlint/depguard-biting become `hook`, layer rules `hook (..., dormant)`, typed-IDs/mockery `convention-only`, and nothing is `CI-gated` (no CI) — refined the single canon status legend in `authoring.md`, documented the one-time tooling bootstrap in `build.md`, and registered the new enforcement facts honestly in `boundaries.md`/`README.md`.**

## Performance

- **Duration:** ~5 min
- **Tasks:** 2 (both auto)
- **Files modified:** 7 knowledge docs

## Accomplishments
- Closed out ENF-05: `grep -rn 'planned:' knowledge/` returns nothing — every forward mark now carries its real enforcement mechanism, backed by the configs shipped in waves 1–3.
- Made `authoring.md` §«Статус enforcement» the single canon legend: defines the 3 statuses precisely (`hook` = committed git hook today, `convention-only` = review-enforced, `CI-gated` = reserved until a CI pipeline exists), adds the **MUST NOT** mark `CI-gated` without CI rule, and documents the `biting`/`dormant` depguard subnotes. Stale "Phase 1 fixes only the standard" wording replaced with "статусы проставлены (Phase 4, ENF-05)".
- Kept honesty airtight: typed-IDs → `convention-only` (A1 — no off-the-shelf linter), mockery → `convention-only` (A2 — harness deferred), layer-direction depguard rules → `hook (lint: depguard, dormant)` (they match no Go files yet), and no rule is ever labeled `CI-gated`.
- Documented the one-time `make tools` + `npm install` + `lefthook install` bootstrap in `build.md` (D-11) in the existing recipe style, pointer-over-copy to the `Makefile` version table, with the `inventory` pre-push exclusion stated as intentional and buf kept skeleton.
- Registered the new enforcement facts in `boundaries.md`'s ownership map (taxonomy → authoring.md, bootstrap/commands+inventory exclusion → build.md) and added an honest root-configs note to `README.md` — link integrity across all `knowledge/*.md` intact.

## Task Commits

Each task was committed atomically:

1. **Task 1: Refine authoring.md legend + flip marks in style/architecture/testing; verify patterns** — `5bf7c89` (docs)
2. **Task 2: Document tooling bootstrap in build.md + register enforcement facts in boundaries.md/README.md** — `9868758` (docs)

**Plan metadata:** committed separately with SUMMARY.md / STATE.md / ROADMAP.md / REQUIREMENTS.md.

## Files Created/Modified
- `knowledge/authoring.md` — §«Статус enforcement» rewritten as the canon 3-status legend: `hook`/`convention-only`/`CI-gated`-reserved defined precisely, MUST NOT mark CI-gated without CI, biting/dormant depguard subnotes, cross-refs to `lefthook.yml`/`.golangci.yml`/`build.md`; "statuses set (Phase 4, ENF-05)".
- `knowledge/style.md` — gofumpt prose → `hook (format: gofumpt)`; typed-IDs L53 → `convention-only (review-enforced)`; `%w` → `hook (lint: errorlint)`; DTO→domain → `hook (lint: depguard, dormant)`; preamble now points to the authoring legend instead of "Phase 4 переключит".
- `knowledge/architecture.md` — 2 domain-direction marks → `hook (lint: depguard, dormant)` (one with "становится CI-gated при появлении CI"); 2 resurrection bans → `hook (lint: depguard, biting)` (no-cqrs-bus on `pkg/mediatr`, no-tx-manager); Manual-Only review note refined to state the rule is dormant because layer code does not exist yet.
- `knowledge/testing.md` — mockery → `convention-only (review-enforced)` + "mockery-обвязка отложена"; "planned: Phase 4" phrasing (prose + illustration comment) de-phantomed to "отложены/отложена".
- `knowledge/build.md` — new §«Bootstrap тулинга (разово на клон)»: `make tools` (pointer to Makefile versions), `npm install` (single Node dep), `lefthook install`; post-bootstrap hook behavior + intentional `inventory` pre-push exclusion + buf-skeleton no-phantom note.
- `knowledge/boundaries.md` — ownership-map rows: enforcement-status taxonomy → authoring.md, bootstrap/commands+inventory exclusion → build.md; honest note that root configs exist after Phase 4 while buf codegen does not.
- `knowledge/README.md` — "Что где живёт" note that enforcement configs live in repo root, with facts canonized in authoring/build/boundaries.

## Decisions Made
- **A1 — typed-IDs `convention-only`:** there is no off-the-shelf linter that enforces "use a named type for IDs"; a custom rule is out of scope. Flipped to `convention-only (review-enforced)` rather than inventing a `hook` for an unconfigured linter (no-phantom).
- **A2 — mockery `convention-only`:** the full mockery harness (`.mockery.yaml` + `go:generate` + install) is Deferred per CONTEXT. Kept the tool-choice convention; removed every "planned: Phase 4" phrasing that implied this phase wired generation.
- **`CI-gated` is never a rule status:** per D-07/D-08/Pitfall 8 it appears only as a *definition* (authoring.md legend), a *forward note* (architecture.md "becomes CI-gated when CI lands"), and a *taxonomy reference* (boundaries.md ownership row, `CI-gated`-reserved). The gate `grep -rnE '⟶[^—]*CI-gated'` confirms no rule carries it as a status.
- **patterns.md untouched:** verified via grep that all its marks are already `convention-only` and no `planned:` remains — no edit, no churn.

## Deviations from Plan

None — plan executed exactly as written. The plan flagged the architecture.md Manual-Only review note as one to keep; it was refined (not removed) to truthfully state the depguard layer rule is dormant because layer code does not yet exist, which strengthens the no-phantom posture without changing the rule.

## Issues Encountered
None. All grep gates and the markdown link-integrity scan passed on the first run.

## User Setup Required
None for this plan (docs only). The bootstrap it documents — `make tools` → `npm install` → `lefthook install` — is the one-time per-clone setup that activates the hooks the flipped `hook` statuses now reference; the live end-to-end hook-firing checks remain carried forward from 04-01/04-02/04-03 (tools were absent on the machine during those runs).

## Next Phase Readiness
- ENF-05 done — the knowledge base is no longer purely declarative: each mechanizable rule now carries its real enforcement mechanism, with dormant/biting/skeleton states all stated honestly.
- The dormant depguard layer rules (`hook (lint: depguard, dormant)`) activate automatically when layer code (`domain/usecases/...`) lands — that restoration is a separate implementation epic.
- `CI-gated` is reserved: when a CI pipeline is added (future epic), the same `.golangci.yml`/`buf` are reused unchanged and the relevant `hook` rules can be promoted per the authoring.md legend rule.

## Threat Flags

None — no new security surface beyond the plan's `<threat_model>`. The doc-only mitigations are implemented as specified: T-4-09 (no `CI-gated` rule status; legend forbids it without CI), T-4-10 (dormant depguard / mockery convention-only / buf skeleton all stated honestly), T-4-11 (global `grep -rn 'planned:' knowledge/` gate returns nothing). T-4-SC is accepted at phase level (this plan ran no installs).

## Self-Check: PASSED
- FOUND: knowledge/authoring.md (legend rewritten)
- FOUND: knowledge/build.md (bootstrap section)
- FOUND: knowledge/boundaries.md (ownership rows)
- FOUND: .planning/phases/04-enforcement/04-04-SUMMARY.md
- FOUND commit 5bf7c89 (Task 1)
- FOUND commit 9868758 (Task 2)
- GATE: `grep -rn 'planned:' knowledge/` returns nothing
- GATE: no rule status `CI-gated` (confined to legend/note/taxonomy reference)

---
*Phase: 04-enforcement*
*Completed: 2026-06-17*
