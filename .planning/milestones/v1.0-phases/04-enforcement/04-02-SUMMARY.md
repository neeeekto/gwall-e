---
phase: 04-enforcement
plan: 02
subsystem: infra
tags: [commitlint, conventional-commits, buf, protobuf, codegen-skeleton, npm, tooling]

# Dependency graph
requires:
  - phase: 04-enforcement
    plan: 01
    provides: "Makefile pinning buf v1.71.0 (make tools); .golangci.yml live lint/format gate"
provides:
  - "Root private package.json — the single Node devDep footprint, exact-pinned @commitlint/cli + @commitlint/config-conventional 21.0.2 (ENF-03, D-04)"
  - "commitlint.config.mjs extending @commitlint/config-conventional (ESM under Node 22)"
  - "buf.yaml v2 lint+breaking SKELETON (module path proto, no .proto yet, NOT hooked) (ENF-04, D-10)"
  - "buf.gen.yaml v2 codegen SKELETON with exact-pinned remote plugins (inert until .proto land)"
  - "package-lock.json — reproducible commitlint install"
affects: [04-03, 04-04, enforcement, lefthook, commit-msg, proto-codegen]

# Tech tracking
tech-stack:
  added: ["@commitlint/cli 21.0.2", "@commitlint/config-conventional 21.0.2", "buf v2 config schema", "protocolbuffers/go v1.36.5 (pinned, inert)", "grpc/go v1.5.1 (pinned, inert)"]
  patterns: ["single-private-Node-dep (private:true, exact pins, no other devDep)", ".mjs to force ESM where package.json lacks type:module", "honest buf skeleton (seeded + marked, codegen-activates-later, no-phantom)"]

key-files:
  created: [package.json, package-lock.json, commitlint.config.mjs, buf.yaml, buf.gen.yaml]
  modified: []

key-decisions:
  - "Exact-pin commitlint 21.0.2 (latest stable, confirmed via npm view at execution) — no caret/tilde (D-04 supply-chain hardening)"
  - "commitlint.config.mjs (not .js) — .mjs forces ESM since package.json has no type:module; avoids CJS-resolution breaking export default under Node 22"
  - "buf skeleton validated structurally only (buf absent on machine) — no buf build claimed (no-phantom); buf build to be re-run after make tools"
  - "Legitimacy checkpoint (Task 1) auto-approved on RESEARCH evidence: too-new flag is a verified false-positive; exact-pin applied"

patterns-established:
  - "Single isolated Node dependency: package.json is private:true with commitlint as the ONLY devDep tree; package-lock committed for reproducibility"
  - "Honest codegen skeleton: pinned plugins + SKELETON header comments stating not-hooked / activates-when-proto-added"

requirements-completed: [ENF-03, ENF-04]

# Metrics
duration: ~4min
completed: 2026-06-17
---

# Phase 4 Plan 02: commitlint config + buf skeleton Summary

**Private root `package.json` exact-pinning `@commitlint/cli` + `@commitlint/config-conventional` 21.0.2 (the single Node devDep) with a `.mjs` config extending config-conventional — round-trip verified — plus an honest buf v2 SKELETON (`buf.yaml` lint+breaking, `buf.gen.yaml` with pinned codegen plugins) marked not-hooked / activates-when-`.proto`-land.**

## Performance

- **Duration:** ~4 min
- **Tasks:** 3 (1 checkpoint auto-approved + 2 auto)
- **Files created:** 5 (package.json, package-lock.json, commitlint.config.mjs, buf.yaml, buf.gen.yaml)

## Accomplishments
- Made Conventional Commits mechanically enforceable: commitlint installed and round-trip verified — rejects `bad message` (exit 1), accepts the repo's GSD style `docs(04):` / `feat(04):` (exit 0). Plan 04-03 wires it into the `commit-msg` hook via `npx --no-install commitlint`.
- Kept commitlint the ONLY Node dependency: `package.json` is `private: true`, exact-pinned (no caret/tilde), no other devDep; `node_modules/` stays gitignored, `package-lock.json` committed for reproducible installs (76 packages, 0 vulnerabilities).
- Cleared the blocking legitimacy gate (T-4-SC) before any install with live evidence (see Checkpoints).
- Seeded proto codegen honestly: `buf.yaml` (v2, `lint.use: [STANDARD]`, `breaking.use: [FILE]`, module path `proto`) and `buf.gen.yaml` (v2, exact-pinned `protocolbuffers/go:v1.36.5` + `grpc/go:v1.5.1`, `out: gen/go`, `paths=source_relative`) — both carrying SKELETON headers stating no `.proto` exist, NOT wired into any hook, codegen activates when schemas land (D-10, no-phantom). Not wired into any lefthook hook.

## Task Commits

Each task committed atomically:

1. **Task 1: Legitimacy gate (checkpoint:human-verify)** — auto-approved per `--auto` policy; no commit (gate only).
2. **Task 2: package.json + commitlint.config.mjs (ENF-03), npm install bootstrap** — `88aa153` (feat)
3. **Task 3: buf.yaml + buf.gen.yaml skeleton (ENF-04)** — `393853e` (feat)

**Plan metadata:** committed separately with SUMMARY.md / STATE.md / ROADMAP.md.

## Files Created/Modified
- `package.json` — `name: gwall-e-tooling`, `private: true`, devDependencies ONLY: `@commitlint/cli` `21.0.2` + `@commitlint/config-conventional` `21.0.2` (exact pins, no caret/tilde). The single Node footprint.
- `package-lock.json` — lockfile for the 76-package commitlint dependency tree (reproducible install).
- `commitlint.config.mjs` — `export default { extends: ['@commitlint/config-conventional'] }`; `.mjs` forces ESM under Node 22 (RESEARCH Open Question 2). config-conventional's type-enum already accepts GSD `docs(NN):`/`feat(NN):` style.
- `buf.yaml` — v2; `modules: [{ path: proto }]`; `lint.use: [STANDARD]`; `breaking.use: [FILE]`. SKELETON header: no `.proto` yet, not hooked, activates-later.
- `buf.gen.yaml` — v2; pinned remote plugins `buf.build/protocolbuffers/go:v1.36.5` + `buf.build/grpc/go:v1.5.1`, each `out: gen/go`, `opt: [paths=source_relative]`. SKELETON header: inert, not hooked, codegen activates when `.proto` are added. No protovalidate (optional, deferred).

## Checkpoints

**Task 1 — Legitimacy gate (`checkpoint:human-verify`, gate=blocking-human): AUTO-APPROVED.**
This run is `--auto`. Per the auto-mode package-legitimacy policy the gate was auto-approved on the pre-vetted RESEARCH evidence, re-confirmed live at execution via `npm view`:
- `@commitlint/cli` and `@commitlint/config-conventional` latest stable = `21.0.2` (matches RESEARCH recommendation and dist-tag `latest`).
- Repository: `git+https://github.com/conventional-changelog/commitlint.git` (real, well-known source).
- No `postinstall`/`install` script — only dev `deps`/`pkg` scripts (no supply-chain install-time execution).
- Not deprecated (empty `deprecated` field).
- The `too-new` SUS verdict is a verified false-positive (keyed on latest-release date, not package age — both are mature, 8M+ weekly downloads).
Disposition: exact-pin `21.0.2`. `npm install` ran after approval: 76 packages added, 0 vulnerabilities.

## Decisions Made
- **Exact pin `21.0.2`** confirmed as current latest stable via `npm view @commitlint/cli version` and `dist-tags` at execution time. No settling-buffer downgrade needed.
- **`.mjs` over `.js`** for the commitlint config to force ESM unconditionally (package.json has no `type: module`), per RESEARCH Open Question 2.
- **buf validated structurally only:** `buf` is not installed on this machine (RESEARCH §Environment Availability consistent). YAML parses cleanly (ruby `YAML.load_file`) and all required v2 keys are present (`version: v2`, `lint.use: [STANDARD]`, `breaking.use: [FILE]`, module path `proto`; `buf.gen.yaml` two exact-pinned plugins). Per no-phantom, NO `buf build` pass is claimed — it must be re-run after `make tools` installs buf v1.71.0.

## Deviations from Plan

None — plan executed exactly as written. Verification fallbacks for absent tooling (buf) were anticipated by the plan's `<verify>` block ("MISSING — run 'make tools' first") and are not deviations.

## Issues Encountered
- `buf` binary absent on the machine (expected — installed by `make tools`, plan 04-01's Makefile). Resolved by structural YAML validation; `buf build` carried forward.
- `package-lock.json` was generated by `npm install` and committed alongside `package.json` (reproducible-install footprint); `node_modules/` confirmed already gitignored — not committed.

## User Setup Required
One-time bootstrap per clone (documented in build.md by plan 04-04): after `make tools` (golangci/lefthook/buf) → `npm install` (commitlint, done here for this machine) → `lefthook install` (git hooks, plan 04-03). After `make tools`, honestly verify the buf skeleton: `buf build` should exit 0 on the empty module.

## Next Phase Readiness
- `commitlint.config.mjs` is the live config plan 04-03's `lefthook.yml` `commit-msg` hook will invoke via `npx --no-install commitlint --edit {1}` (RESEARCH §lefthook.yml).
- ENF-05 (plan 04-04) will flip the git.md Conventional-Commits enforcement mark to reflect the live `commit-msg` hook, and any proto/buf forward marks toward this seeded skeleton (kept phrased as "activates when `.proto` are added").
- **Carry-forward (no-phantom):** `buf build` has NOT been executed (buf absent); the phase gate must run it after `make tools`. buf is wired into NO hook this phase by design (D-10).

## Threat Flags

None — no new security surface beyond the plan's `<threat_model>`. Mitigations implemented as specified: T-4-SC (legitimacy gate cleared with live evidence + exact pins, `private: true`), T-4-04 (buf plugins exact-pinned, inert), T-4-05 (SKELETON headers prevent phantom-codegen claims), T-4-06 (`commit-msg` hook will use `npx --no-install` — set up by 04-03).

## Self-Check: PASSED
- FOUND: package.json
- FOUND: package-lock.json
- FOUND: commitlint.config.mjs
- FOUND: buf.yaml
- FOUND: buf.gen.yaml
- FOUND commit 88aa153 (Task 2)
- FOUND commit 393853e (Task 3)

---
*Phase: 04-enforcement*
*Completed: 2026-06-17*
