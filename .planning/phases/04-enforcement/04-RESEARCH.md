# Phase 4: Enforcement-слой (тулинг) - Research

**Researched:** 2026-06-17
**Domain:** Go multi-module enforcement tooling (golangci-lint v2, Lefthook git hooks, commitlint, buf proto skeleton) + truthful enforcement-status reconciliation in a knowledge base
**Confidence:** HIGH (tool versions + config schemas cross-checked vs official release pages/docs), MEDIUM (multi-module orchestration mechanics — community-corroborated, no single authoritative source), HIGH (repo facts — direct observation)

## Summary

Phase 4 wires **local** enforcement tooling (no CI) to back the `knowledge/*.md` rules and flips each Phase-3 forward enforcement mark (`planned: CI-gated Phase 4 (...)`) to a **truthful** status. The toolchain is already pinned by `.planning/research/STACK.md` and re-verified here against official release pages: golangci-lint **v2.12.2** (May 2026), buf **v1.71.0** (Jun 2026), commitlint **v21.x**. Two STACK.md figures drifted and are corrected below: **Lefthook is now v2.x (latest v2.1.9)**, not v1.x — and golangci-lint's actual latest is v2.12.2 (STACK.md said "v2.12.x" — consistent). Lefthook v2 keeps the `lefthook.yml` `pre-commit`/`pre-push`/`commit-msg` + `commands`/`run`/`glob`/`stage_fixed` schema; v1 examples transfer directly.

The hard parts are **honesty constraints**, not tooling: (1) the repo has **no layer code** (`domain/usecases/...` are empty `internal/` dirs in WIP `inventory`), so depguard import-direction rules are authored **forward-compatible and dormant**; (2) there is **no CI runner this phase**, so nothing may be labeled `CI-gated` — Phase-3 marks that said `CI-gated` flip to **`hook`**; (3) there are **no `.proto`**, so buf ships as a clearly-marked **skeleton not wired into any failing hook**; (4) `inventory` is excluded from `pre-push` deliberately and that exclusion is documented as intentional, not a gap.

**Primary recommendation:** Ship a single root `.golangci.yml` (v2 schema: `linters.default: standard` + opt-in `depguard`/`errorlint`; `formatters: [gofumpt, gci]`); a root `lefthook.yml` (pre-commit = `golangci-lint run` per-module incl. `GOWORK=off` inventory lint; pre-push = `go test` for in-workspace modules only; commit-msg = `npx --no-install commitlint --edit {1}`); a private root `package.json` exact-pinning commitlint; a skeleton `buf.yaml`+`buf.gen.yaml` (v2) **referenced from no hook**; pin Go tools via a `Makefile` `make tools` target (recommended over `go.mod tool` to keep the polluted root module out of it); and flip every forward mark per the **ENF-05 mapping table** below.

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Single root `.golangci.yml` (v2 schema: `linters.default: standard` + opt-in, `formatters` block for gofumpt+gci) — one source of truth. Run must cover BOTH in-workspace modules (`./pkg`, `./services/analytics`, `./services/audit` via `go.work`) AND `inventory` (outside workspace, separate pass with `GOWORK=off` from its module dir). Exact run mechanic (single workspace run vs per-module loop) is planner/research discretion, but **both** module sets must be linted.
- **D-02:** `pre-commit` = lint + format. Formatting (gofumpt + gci) wired **through golangci-lint v2 `formatters`** (format == lint == future CI identical), NOT a separate parallel gofumpt call. Keep hook fast (prefer staged files); CI is future source of truth (out of scope).
- **D-03:** `pre-push` = tests **for in-workspace modules only** (`pkg`, `analytics`, `audit`). `inventory` EXCLUDED (WIP, outside workspace, `internal/` deleted — boundaries.md forbids running/fixing WIP scaffolding). Exclusion documented explicitly so it is not read as an oversight. Ginkgo where a suite exists (real reference: `pkg/http`).
- **D-04:** Follow STACK.md: minimal root `package.json` (`private: true`, only `devDependencies`, exact-pin `@commitlint/cli` + `@commitlint/config-conventional`) + `commitlint.config.js` (extends `config-conventional`). `commit-msg` hook calls `npx --no-install commitlint --edit {1}`. `npm install` documented next to `lefthook install` as one-time bootstrap. commitlint stays the **single** Node dependency, isolated to dev tooling. Go-native alternative → Deferred.
- **D-05:** depguard configured **forward-compatible**: import-direction rules from architecture.md (domain imports nothing outward; `usecases → domain`; `api`/`repositories → usecases`/`domain`) written against **target layer paths** (`.../internal/domain`, `.../usecases` etc.). No layer code now → these rules **lie dormant** (match nothing) and activate as code lands — no phantom claim they already check anything.
- **D-06:** In parallel — **immediately-biting** "resurrection bans" (concrete WON'T from architecture.md/PROJECT.md): ban importing removed `pkg/mediatr` (CQRS bus), `TxManager`/`tx` dispatcher. These guards are real **today** (fire on any reintroduction). Plus `errorlint` for sentinel-vs-wrapped (`%w`/`errors.Is`) — flips style.md "planned: CI-gated Phase 4 (errorlint)". Exact linter set (depguard, errorlint, + optional for typed-IDs) is planner discretion within Phase-3 forward marks.
- **D-07:** Phase 4 wires local lefthook hooks; full CI out of scope. Statuses defined honestly: `hook` = rule checked by a local git hook via committed config (available today); `convention-only` = review-enforced; `CI-gated` = reserved until a CI pipeline exists (same `.golangci.yml`/`buf` reused in CI unchanged).
- **D-08:** Forward marks D-11 flip accordingly: `planned: CI-gated Phase 4 (depguard/errorlint/...)` → **`hook (lint: depguard/...)`** (NOT `CI-gated` — no CI yet); `planned: hook (gofumpt)` → `hook (format)`; `convention-only (review-enforced)` stays. architecture.md CI-gated-depguard marks relabel to `hook`, with **one line** noting the config becomes CI-gated when CI lands.
- **D-09:** Status legend refined in **one canon** — `knowledge/authoring.md` §"Статус enforcement" (define 3 statuses precisely + rule "don't mark CI-gated without CI"). Topic docs reference the legend, don't redefine. Change wording "Phase 1 fixes only the standard" to "statuses set (Phase 4)".
- **D-10:** `buf.yaml` (v2: `lint` + `breaking` config, module points at future proto-root) and `buf.gen.yaml` (v2: plugins with **version pinning** — `protoc-gen-go`, `protoc-gen-go-grpc`; optional protovalidate per STACK). Marked as **skeleton**: no `.proto` yet, so buf is **NOT** wired into failing lefthook hooks (proto codegen/lint manual/opt-in until schemas exist). no-phantom: do not present codegen as "working".
- **D-11:** All tool versions (golangci-lint v2.12.x, gofumpt, gci, lefthook, buf, commitlint exact) pinned **reproducibly**: install targets in `Makefile` (`make tools`) referencing a version table + documented in `knowledge/build.md` (one-time bootstrap: `make tools` + `lefthook install` + `npm install`). Exact mechanism (Makefile vs `go.mod tool` Go 1.24 vs `.tool-versions`/mise) is planner/research discretion; commitlint pinned via `package.json` exact. Versions from STACK.md (HIGH-confidence, June 2026).

### Claude's Discretion

- Exact golangci run mechanic: workspace run vs per-module loop (D-01).
- Concrete linter set beyond gofumpt/gci/depguard/errorlint (D-06), within Phase-3 forward marks.
- Tool version pinning mechanism (Makefile vs `go.mod tool` vs mise) (D-11).
- commitlint config filename (`commitlint.config.js` vs `.commitlintrc.*`) and exact `commit-msg` invocation syntax (D-04).
- Exact proto-root shape in `buf.yaml` and codegen plugin list (D-10).
- Wording of the refined status legend in `authoring.md` (D-09) within the 3-status model.

### Deferred Ideas (OUT OF SCOPE)

- **Full CI pipeline** (runner/matrices/workflow files, promoting `hook` rules to `CI-gated`) — future epic; same `.golangci.yml`/`buf` reused unchanged. **Do NOT create `.github/workflows` or any CI config.**
- **Real `.proto` + working codegen** (filling the buf skeleton, wiring buf into hooks) — unblocked when schemas/a compilable service exist.
- **Go-native commitlint** (lefthook-regex / commitlint-rs / conform) as a Node replacement — future option; now follow research (Node devDep).
- **Full mockery harness** (`.mockery.yaml` + `go:generate` + install) beyond minimal ENF seed — matures with layer/inventory code restoration.
- **Layer code restoration** (`domain/usecases/...`), after which dormant depguard rules (D-05) go active — separate implementation epic.
- **ADR docs, `anti-patterns.md`, `libraries.md`, onboarding, maintenance protocol** — v2.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ENF-01 | `.golangci.yml` (golangci-lint v2; gofumpt formatter; gci import order), workspace-consistent | §Standard Stack (golangci v2.12.2), §Code Examples (full `.golangci.yml` with `linters.default`, `formatters`, `depguard`, `errorlint`), §Architecture Patterns (multi-module lint mechanic), §Pitfall 1/2/4 |
| ENF-02 | `lefthook.yml` — pre-commit (lint+format), pre-push (tests), commit-msg (commitlint) | §Standard Stack (Lefthook v2.1.9 — **drift corrected**), §Code Examples (full `lefthook.yml`), §Architecture (per-module loop incl. `GOWORK=off` inventory lint; in-workspace-only tests), §Pitfall 3/5 |
| ENF-03 | commitlint config (Conventional Commits) wired to commit-msg | §Standard Stack (commitlint v21.x), §Package Legitimacy Audit, §Code Examples (`package.json`, `commitlint.config.js`, commit-msg run), §Pitfall 6 |
| ENF-04 | Skeleton `buf.yaml` + `buf.gen.yaml` (lint/breaking/codegen) | §Standard Stack (buf v1.71.0, v2 config), §Code Examples (skeleton `buf.yaml`/`buf.gen.yaml`), §Pitfall 7 (phantom-codegen phrasing) |
| ENF-05 | Each mechanizable rule in `knowledge/*.md` gets a real enforcement status | §ENF-05 Mapping Table (every current `planned:` mark → flipped status), §Architecture (authoring.md legend refinement D-09), §Pitfall 8 (CI-gated honesty) |

## Architectural Responsibility Map

This phase has no application tiers (it is build/dev tooling + docs), so the map covers **enforcement tiers** instead — which mechanism owns enforcing each rule class.

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Format (gofumpt) + import order (gci) | golangci-lint v2 `formatters` (via pre-commit hook) | future CI (same config) | D-02: format==lint==CI must be one config; gofumpt embedded in golangci avoids a divergent standalone binary |
| Lint rules (errorlint, govet, staticcheck, depguard) | golangci-lint v2 `linters` (via pre-commit hook) | future CI | D-01/D-06: single root config, hook is the live gate today |
| Import-direction layer rules | depguard (dormant — no layer code) | review | D-05: rules authored forward, fire only when layer code lands |
| Resurrection bans (mediatr/TxManager) | depguard (biting today) | review | D-06: any reintroduction is an importable path → depguard catches it now |
| Commit message format | commitlint (via commit-msg hook) | review | D-04: Conventional Commits, single Node dep |
| Test execution gate | `go test`/ginkgo (via pre-push hook) | future CI | D-03: in-workspace modules only; inventory excluded by design |
| Proto lint/breaking/codegen | buf (skeleton, NOT hooked) | future (when `.proto` exist) | D-10: no schemas → manual/opt-in, no phantom codegen |
| Enforcement-status taxonomy | `knowledge/authoring.md` legend (canon) | topic docs reference it | D-09: one canon, pointer-over-copy |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| golangci-lint | **v2.12.2** (2026-05-06) | Single lint+format runner; v2 schema (`linters.default`, `formatters` block) | [CITED: github.com/golangci/golangci-lint/releases] Latest stable; v2 is the only schema new projects should adopt (v1 EOL'd at v1.64.8). Install path **must** include `/v2/`: `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`. |
| Lefthook | **v2.1.9** (2026-05-29) — *STACK.md said v1.x; corrected* | Git hooks manager (pre-commit/pre-push/commit-msg) | [CITED: github.com/evilmartians/lefthook/releases] Single dependency-free Go binary; `lefthook.yml` `pre-commit`/`pre-push`/`commit-msg` + `commands`/`run`/`glob`/`stage_fixed` schema unchanged from v1 — v1 examples transfer. Now installable as a Go tool: `go get -tool github.com/evilmartians/lefthook/v2@v2.1.9`. |
| @commitlint/cli | **v21.x** (pin exact — see audit) | Conventional Commits validation in commit-msg | [CITED: npmjs.com/@commitlint/cli] 8.5M weekly downloads; de-facto Conventional Commits standard. `--edit {file}` reads the commit message file. |
| @commitlint/config-conventional | **v21.x** (pin exact, match cli) | Conventional Commits ruleset | [CITED: npmjs.com/@commitlint/config-conventional] type-enum: build, chore, ci, docs, feat, fix, perf, refactor, revert, style, test. |
| buf (CLI) | **v1.71.0** (2026-06-16) | Proto lint/breaking/codegen — **skeleton only** this phase | [CITED: github.com/bufbuild/buf/releases] v2 config schema (`buf.yaml` + `buf.gen.yaml` version: v2) is current; CLI is still major v1. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gofumpt | embedded in golangci-lint v2.12.2 (gofumpt **0.9.2**) | Stricter gofmt formatter | [CITED: golangci-lint changelog] Use **via golangci `formatters`**, NOT standalone — see Pitfall 4 (standalone gofumpt 0.10.0 requires Go 1.25; repo is Go 1.24.6). |
| gci | embedded in golangci-lint v2 `formatters` | Deterministic import grouping (stdlib / default / `github.com/gwall-e/...`) | [CITED: golangci-lint.run/docs/formatters] Configure under `formatters.settings.gci`. |
| protoc-gen-go | latest stable (pin in `buf.gen.yaml`) | Go message codegen plugin (skeleton) | [ASSUMED] Pinned remote plugin in `buf.gen.yaml`; not executed until `.proto` exist. |
| protoc-gen-go-grpc | latest stable (pin in `buf.gen.yaml`) | gRPC stub codegen plugin (skeleton) | [ASSUMED] Same — skeleton only. |
| protovalidate (buf.build/bufbuild/protovalidate) | optional | Declarative validation rules in proto | [CITED: STACK.md] Optional per STACK.md; skeleton may reference, not required. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Makefile `make tools` pinning | `go.mod tool` directive (Go 1.24) | `go tool` is idiomatic + reproducible, but the **root `github.com/gwall-e` module is a polluted leftover** (stale ginkgo v1, mongo v1) NOT in `go.work` — adding a `tool` block there entangles dev tools with rotten deps. A standalone `tools/go.mod` or a Makefile sidesteps this. **Recommend Makefile** (D-11 discretion); if `go.mod tool` chosen, use a dedicated `tools/go.mod`, not the root. |
| Makefile pinning | `.tool-versions` / mise | mise/asdf pins non-Go tools too (node, buf) in one file; adds an external runtime dependency devs must install. Fine if team already uses mise; heavier otherwise. |
| commitlint (Node) | commitlint-rs / conform / lefthook regex | Go-native, zero node_modules — but loses the canonical config-conventional ruleset + ecosystem. D-04 locks Node; this is Deferred. |
| Lefthook `commands` | Lefthook `jobs` (newer syntax) | `jobs` supports groups/sequential control; `commands` is simpler and fully sufficient here. Use `commands`. |

**Installation (bootstrap — document in `knowledge/build.md`):**
```bash
# Go dev tools (via make tools — pinned versions in Makefile)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install github.com/evilmartians/lefthook/v2@v2.1.9
go install github.com/bufbuild/buf/cmd/buf@v1.71.0
# NOTE: do NOT `go install mvdan.cc/gofumpt@latest` — gofumpt runs embedded inside
#       golangci-lint v2 formatters (avoids the standalone Go 1.25 requirement). See Pitfall 4.

# commitlint (only Node tooling; exact-pin via package.json devDependencies)
npm install            # installs @commitlint/cli + @commitlint/config-conventional (exact-pinned)

# activate git hooks (once per clone)
lefthook install
```

**Version verification (performed this session):**
- `golangci-lint` v2.12.2 — confirmed latest via GitHub releases (2026-05-06). [CITED]
- `lefthook` v2.1.9 — confirmed latest via GitHub releases (2026-05-29). [CITED] **STACK.md drift (said v1.x).**
- `buf` v1.71.0 — confirmed latest via GitHub releases (2026-06-16). [CITED]
- `@commitlint/cli` / `config-conventional` v21.0.2 — confirmed latest via npm (published ~2026-05-29). [CITED]
- `gofumpt` standalone 0.10.0 (2026-05-04) requires Go ≥1.25 — **do not pin standalone**; use golangci-embedded 0.9.2. [CITED: mvdan/gofumpt releases]

## Package Legitimacy Audit

> Only external packages installed this phase are the two commitlint npm packages (Go tools are pinned but installed from their canonical, in-repo-referenced module paths). Verdict from `gsd-tools query package-legitimacy check --ecosystem npm`.

| Package | Registry | Age / Latest publish | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|----------------------|-----------|-------------|---------|-------------|
| `@commitlint/cli` | npm | latest v21.0.2 ~7 days old (pkg itself years old) | 8.55M/wk | github.com/conventional-changelog/commitlint | SUS (`too-new`) | **Approved** — false-positive: `too-new` fires on the *latest release date*, not package age. 8.5M downloads, real repo, no postinstall, not deprecated. **Pin exact** to avoid bleeding edge. |
| `@commitlint/config-conventional` | npm | latest v21.0.2 ~7 days old | 8.12M/wk | github.com/conventional-changelog/commitlint | SUS (`too-new`) | **Approved** — same false-positive. Pin exact, matching cli major. |

**Packages removed due to SLOP verdict:** none.
**Packages flagged as suspicious (SUS):** `@commitlint/cli`, `@commitlint/config-conventional` — both are mature, high-trust packages; the SUS verdict is a `too-new` false-positive driven by a recent release date. **Recommendation for planner:** pin to an exact stable (e.g. `21.0.1` or `21.0.2` — confirm the latest at plan time) in `package.json` with `private: true`; this is a standard, low-risk dependency. A `checkpoint:human-verify` task is **optional** given the 8M-download/real-repo/no-postinstall profile, but acceptable if the workflow mandates one for any SUS verdict.

*Go tools (golangci-lint v2, lefthook v2, buf, embedded gofumpt/gci) are installed from canonical module paths verified against official GitHub release pages this session; treated as VERIFIED via official release pages.*

## Architecture Patterns

### System Architecture Diagram

```text
  git commit ─────────────► lefthook (commit-msg) ──► npx --no-install commitlint --edit {1}
                                                          │  (reads .git/COMMIT_EDITMSG)
                                                          ▼  reject if not Conventional Commits

  git commit (staged .go) ─► lefthook (pre-commit) ──► golangci-lint run  (per module)
                                │                          ├─ formatters: gofumpt + gci  (--fix, stage_fixed)
                                │                          └─ linters: standard + errorlint + depguard
                                │                                         ├─ depguard: layer rules (DORMANT — no code)
                                │                                         └─ depguard: ban mediatr/TxManager (BITING)
                                │   in-workspace modules:  golangci-lint run ./...        (go.work active)
                                │   inventory (out of ws): cd services/inventory && GOWORK=off golangci-lint run ./...
                                ▼
  git push ───────────────► lefthook (pre-push) ─────► tests, in-workspace ONLY:
                                                          ├─ cd pkg && go test ./...        (ginkgo suite exists)
                                                          ├─ cd services/audit && go test ./...
                                                          └─ cd services/analytics && go test ./...   (0 pkgs ok)
                                                          └─ inventory: EXCLUDED (WIP, boundaries.md)

  buf (SKELETON — referenced by NO hook):
     buf.yaml (v2: lint + breaking, module → future proto-root)
     buf.gen.yaml (v2: pinned protoc-gen-go / protoc-gen-go-grpc)
     └─ manual/opt-in only; inert until .proto exist (no-phantom: codegen NOT "working")

  knowledge/*.md  ◄── ENF-05 flips forward marks ──  .golangci.yml / lefthook.yml are the live config
     authoring.md §"Статус enforcement" = canon legend (hook / convention-only / CI-gated-reserved)
```

### Component Responsibilities

| Artifact | Path | Responsibility |
|----------|------|----------------|
| Lint+format config | `.golangci.yml` (repo root) | v2 schema; `linters.default: standard` + errorlint + depguard; `formatters: [gofumpt, gci]` |
| Hook orchestration | `lefthook.yml` (repo root) | pre-commit (lint+format), pre-push (in-ws tests), commit-msg (commitlint) |
| Commit lint deps | `package.json` (repo root, `private: true`) | exact-pinned commitlint devDependencies only |
| Commit lint config | `commitlint.config.js` (repo root) | `extends: ['@commitlint/config-conventional']` |
| Proto skeleton | `buf.yaml` + `buf.gen.yaml` (repo root) | v2 lint/breaking + pinned codegen plugins; NOT hooked |
| Tool pinning | `Makefile` (`make tools` target) | reproducible install of golangci-lint/lefthook/buf at pinned versions |
| Status legend | `knowledge/authoring.md` §"Статус enforcement" | canon 3-status taxonomy (D-09) |
| Bootstrap docs | `knowledge/build.md` | `make tools` + `lefthook install` + `npm install` one-time steps (D-11) |

### Recommended Project Structure (new files this phase)
```
/                       # repo root
├── .golangci.yml       # ENF-01 — v2 schema, single source of truth
├── lefthook.yml        # ENF-02 — git hooks
├── package.json        # ENF-03 — private, commitlint devDeps only
├── commitlint.config.js# ENF-03 — extends config-conventional
├── buf.yaml            # ENF-04 — skeleton (v2, lint+breaking)
├── buf.gen.yaml        # ENF-04 — skeleton (v2, pinned codegen plugins)
├── Makefile            # D-11 — `make tools` pinned install target
└── knowledge/          # ENF-05 — flip forward marks; refine authoring.md legend
```

### Pattern 1: Multi-module lint — per-module loop (recommended over single workspace run)
**What:** golangci-lint operates on one module at a time. With `go.work` active it can lint the workspace, but `inventory` is deliberately outside `go.work` and built with `GOWORK=off`. A per-module loop is the clearest, most honest mechanic and naturally handles both sets.
**When to use:** Always here — it makes the `inventory` `GOWORK=off` pass explicit and keeps the in-workspace vs out-of-workspace boundary visible (matches build.md's existing recipes).
**Example:**
```yaml
# lefthook.yml — pre-commit lint, per-module (covers both go.work and GOWORK=off inventory)
# Source: lefthook.dev configuration + golangci-lint v2 install path; multi-module mechanic = MEDIUM confidence
pre-commit:
  parallel: false   # lint passes touch shared caches; keep deterministic
  commands:
    lint-workspace:
      glob: "*.go"
      # in-workspace modules build with go.work active; lint each module root
      run: |
        set -e
        for m in pkg services/analytics services/audit; do
          (cd "$m" && golangci-lint run ./...)
        done
      stage_fixed: true
    lint-inventory:
      glob: "services/inventory/**/*.go"
      # inventory is outside go.work → must lint with GOWORK=off from its dir
      run: cd services/inventory && GOWORK=off golangci-lint run ./...
      stage_fixed: true
```
> **Honesty note:** `services/analytics` currently has **no Go packages** (build matches 0 packages — see build.md). `golangci-lint run ./...` printing "no packages" there is expected, not an error. `inventory` is WIP and not guaranteed to compile; its lint pass may surface compile-stage issues — that is acceptable for a WIP module and **must not** trigger "fixing" the scaffolding (boundaries.md). Consider `lint-inventory` advisory (non-blocking) or scoped to only fire when inventory has compilable code; planner discretion. The simplest honest default: keep it, document that inventory lint is best-effort WIP.

### Pattern 2: depguard — dormant forward rules + biting resurrection bans (single linter, two rule sets)
**What:** depguard v2 uses named `rules`, each with `files`/`list-mode`/`allow`/`deny`. Layer-direction rules target paths that don't exist yet (dormant); resurrection bans target removed package paths (bite on any reintroduction).
**When to use:** ENF-01 + sets up ENF-05 flip for architecture.md/style.md import-direction marks.
**Example:** see §Code Examples `.golangci.yml`.
> **Dormancy is real, not phantom:** depguard matches files by path prefix. With no `.../internal/domain` Go files, the `domain` rule's `files` selector matches nothing → the rule is inert. This is **not** a phantom claim because the config does not assert anything runs; it is a forward declaration that activates when code lands. Document this in the rule `desc` and in the ENF-05 status note.

### Pattern 3: commit-msg → commitlint with `--no-install`
**What:** Lefthook passes the commit-message file path as the positional `{1}` template variable. `npx --no-install` runs the locally-installed commitlint without attempting a network fetch.
**When to use:** ENF-03.
**Example:**
```yaml
commit-msg:
  commands:
    commitlint:
      run: npx --no-install commitlint --edit {1}
```
> `--no-install` (not bare `npx`) guarantees the hook fails loudly if `npm install` was skipped, rather than silently pulling a different commitlint version from the network. Document `npm install` as a bootstrap step.

### Anti-Patterns to Avoid
- **Separate standalone `gofumpt` call in pre-commit:** violates D-02 (format must == lint == CI). Run gofumpt **inside** golangci `formatters`. Also avoids the gofumpt-0.10.0 Go-1.25 requirement (Pitfall 4).
- **Marking anything `CI-gated` this phase:** there is no CI runner. Violates D-07/no-phantom. Use `hook` or `convention-only` only.
- **Wiring buf into a lefthook hook:** no `.proto` exist → the hook would either no-op confusingly or fail. Keep buf manual/opt-in (D-10).
- **Adding a `tool` block to the root `github.com/gwall-e` go.mod:** that module is a rotten leftover (ginkgo v1, mongo v1) outside `go.work`. Use a Makefile or a dedicated `tools/go.mod`.
- **Running `inventory` tests in pre-push:** boundaries.md forbids running/fixing WIP scaffolding (D-03).
- **golangci-lint v1 config schema** (`enable-all`/`disable-all`): v2 only. Use `linters.default`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Import-direction enforcement | Custom AST script / grep checking layer imports | depguard (in golangci v2) | Path-prefix rules, list-modes, per-file selectors — handles dormant-then-active cleanly |
| `%w`/`errors.Is` checks | Regex grep for `%v` on errors | errorlint | Understands wrapping semantics, type-asserted errors, switch type-switches |
| Commit message validation | Hand-rolled regex in a shell hook | commitlint + config-conventional | Canonical ruleset, helpful messages, ecosystem (changelog/versioning later) |
| Git hook installation/orchestration | Hand-written `.git/hooks/*` scripts | Lefthook | One `lefthook.yml`, parallel/staged-file support, `lefthook install` per clone |
| Proto codegen | Raw `protoc` + shell + manual plugin versions | buf generate + `buf.gen.yaml` | Pins plugin versions, breaking-change detection, monorepo-aware (skeleton now) |
| Import ordering | Custom goimports wrapper | gci (in golangci v2 formatters) | Deterministic, configurable sections, integrated with the one lint config |

**Key insight:** Every mechanizable rule in this phase already has a best-in-class Go-ecosystem tool that **reuses the same config in future CI unchanged** — hand-rolling would fork format/lint behavior between the local hook and future CI, exactly the divergence D-02 exists to prevent.

## Runtime State Inventory

> This is a config/docs phase, not a rename/migration — but it touches enforcement state. Inventory of state that the new configs interact with:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no datastore keys/collections touched by enforcement tooling. | none |
| Live service config | None — no external service config. **Git hooks are installed into `.git/hooks/` by `lefthook install`** (per-clone local state, not committed). | document `lefthook install` as bootstrap (D-11); not a migration |
| OS-registered state | None. | none |
| Secrets/env vars | `GOWORK=off` is an invocation-time env var (not a stored secret); used for the inventory lint/test pass. | none — already documented in build.md |
| Build artifacts | `services/audit/cmd` exists as a directory (benign `go build ./...` "output already exists" warning — not a failure). No egg-info/compiled-tool artifacts to migrate. **Go tools are NOT yet installed** (golangci-lint/lefthook/buf/gofumpt absent on this machine — verified). | `make tools` bootstrap installs them; node v22.20.0 / npm 10.9.3 present |

**Nothing found requiring data migration.** The only "state" is per-clone git-hook installation and per-clone tool installation, both covered by the documented bootstrap.

## Common Pitfalls

### Pitfall 1: golangci-lint v2 install path missing `/v2/`
**What goes wrong:** `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.12.2` fails or installs v1.
**Why it happens:** v2 moved the main package under `/v2/`. Many blog snippets show the v1 path.
**How to avoid:** Use `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2` (note `/v2/`).
**Warning signs:** `go install` resolves an old version or "module ... found, but does not contain package".

### Pitfall 2: Mixing v1 and v2 `.golangci.yml` schema
**What goes wrong:** `enable-all`/`disable-all` keys silently ignored or rejected.
**Why it happens:** v2 replaced them with `linters.default: standard|all|none|fast` and moved formatters out of `linters`.
**How to avoid:** Start clean on v2 (the repo has no existing `.golangci.yml`). Set `version: "2"`, `linters.default: standard`, opt-in extras, and a separate top-level `formatters:` block.
**Warning signs:** gofumpt/gci configured under `linters` instead of `formatters` → not applied.

### Pitfall 3: Lefthook major version drift (STACK.md said v1.x)
**What goes wrong:** Pinning `lefthook` "v1.x" per STACK.md installs an outdated major; or planner assumes a v1-only install path.
**Why it happens:** STACK.md (June 2026) predates/mis-states the current v2.1.9 line.
**How to avoid:** Pin **v2.1.9** (`github.com/evilmartians/lefthook/v2@v2.1.9`). The `lefthook.yml` schema (pre-commit/pre-push/commit-msg, commands/run/glob/stage_fixed) is unchanged from v1, so config carries over.
**Warning signs:** `go install github.com/evilmartians/lefthook@...` (no `/v2`) resolves a stale binary.

### Pitfall 4: Standalone gofumpt 0.10.0 requires Go 1.25 (repo is Go 1.24.6)
**What goes wrong:** `go install mvdan.cc/gofumpt@latest` pulls 0.10.0, which requires Go ≥1.25; install or run breaks on the Go 1.24.6 toolchain.
**Why it happens:** gofumpt 0.10.0 (2026-05-04) became a fork of gofmt as of Go 1.26 and bumped its minimum to Go 1.25.
**How to avoid:** **Do not install gofumpt standalone.** Run it through golangci-lint v2's `formatters` (golangci v2.12.2 embeds gofumpt 0.9.2, compatible). This also satisfies D-02 (format==lint==CI from one config).
**Warning signs:** `make tools` includes a `go install mvdan.cc/gofumpt` line; gofumpt configured as a separate pre-commit command.

### Pitfall 5: `inventory` linted/tested as if it were healthy
**What goes wrong:** pre-push runs `inventory` tests, or pre-commit blocks on `inventory` WIP compile errors → devs `--no-verify`.
**Why it happens:** `inventory` is outside `go.work`, WIP, `internal/` empty — but a naive loop includes it.
**How to avoid:** Exclude `inventory` from pre-push entirely (D-03). For pre-commit lint, run inventory with `GOWORK=off` and treat it as **best-effort/advisory** WIP; document the exclusion as intentional (boundaries.md), not an oversight.
**Warning signs:** pre-push step `cd services/inventory && go test`; hook failing on known-WIP code.

### Pitfall 6: commitlint runs from the wrong directory / pulls from network
**What goes wrong:** `npx commitlint` (without `--no-install`) silently downloads a different version; or the config isn't found.
**Why it happens:** Bare `npx` auto-installs missing packages; commitlint config must be discoverable from repo root.
**How to avoid:** `npx --no-install commitlint --edit {1}`; keep `package.json` + `commitlint.config.js` at repo root; document `npm install` bootstrap.
**Warning signs:** commit-msg hook slow on first run (network fetch); "config not found".

### Pitfall 7: buf skeleton presented as working codegen (phantom)
**What goes wrong:** Docs/commit claim "proto codegen is set up" when no `.proto` exist and no plugin has run.
**Why it happens:** Shipping `buf.gen.yaml` looks like working codegen.
**How to avoid:** Mark both files **skeleton**; reference from NO hook; phrase docs as "buf config seeded; codegen activates when `.proto` are added" (no-phantom). See §Code Examples for exact wording.
**Warning signs:** `buf generate` mentioned as a verified working command; buf wired into lefthook.

### Pitfall 8: Flipping a mark to `CI-gated` when there is no CI
**What goes wrong:** ENF-05 sets `CI-gated` statuses, claiming a pipeline that doesn't exist.
**Why it happens:** Phase-3 forward marks literally say "CI-gated Phase 4".
**How to avoid:** Per D-07/D-08 those flip to **`hook`** (the live mechanism today), with a one-line note "becomes CI-gated when CI lands". Reserve `CI-gated` for the future epic.
**Warning signs:** any `CI-gated` status in `knowledge/*.md` after Phase 4 while `.github/workflows` is absent.

## Code Examples

> All snippets below are **target artifacts to create** (the repo has none of these files yet). Versions verified this session. Multi-module mechanic and exact rule paths are MEDIUM confidence — validate `golangci-lint run` locally during execution.

### `.golangci.yml` (ENF-01) — v2 schema, format+lint+depguard+errorlint
```yaml
# Source: golangci-lint.run v2 docs (linters.default, formatters, depguard, errorlint) — CITED
version: "2"

linters:
  default: standard            # v2: replaces enable-all/disable-all
  enable:
    - errorlint                # flips style.md "%w / errors.Is" mark → hook
    - depguard                 # layer-direction (dormant) + resurrection bans (biting)
  settings:
    errorlint:
      errorf: true             # require %w when wrapping with fmt.Errorf
      asserts: true
      comparison: true
    depguard:
      rules:
        # --- BITING TODAY: ban resurrecting removed subsystems (D-06) ---
        no-cqrs-bus:
          list-mode: lax
          files:
            - "$all"
          deny:
            - pkg: "github.com/gwall-e/pkg/mediatr"
              desc: "CQRS bus removed intentionally; inbound adapter calls the use case directly (architecture.md §MUST NOT CQRS)."
        no-tx-manager:
          list-mode: lax
          files:
            - "$all"
          deny:
            # match a removed tx dispatcher package by path fragment when it lands again
            - pkg: "github.com/gwall-e/services/inventory/internal/tx"
              desc: "TxManager/tx dispatcher removed; use the domain UnitOfWork port (architecture.md §UnitOfWork)."
        # --- DORMANT until layer code exists (D-05): domain must not import outward ---
        domain-imports-inward-only:
          list-mode: strict
          files:
            - "**/internal/domain/**"   # no such Go files yet → rule matches nothing today
          allow:
            - "$gostd"
            - "github.com/gwall-e/pkg"   # shared lib allowed; refine when layers land
          deny:
            - pkg: "github.com/gwall-e/services"
              desc: "domain imports nothing outward; declare ports, implement in adapters (architecture.md §Слои). DORMANT until layer code exists."

formatters:
  enable:
    - gofumpt                  # flips style.md "gofumpt planned: hook" → hook (format)
    - gci
  settings:
    gci:
      sections:
        - standard
        - default
        - prefix(github.com/gwall-e)   # project import group
      custom-order: true
    gofumpt:
      module-path: github.com/gwall-e
      extra-rules: false
```
> The dormant `domain-imports-inward-only` rule and the `no-tx-manager` path are **forward declarations**; adjust the exact paths when layer code lands. The two `no-*` resurrection bans bite immediately on any reintroduction of those import paths.

### `lefthook.yml` (ENF-02)
```yaml
# Source: lefthook.dev configuration docs (commands/run/glob/stage_fixed, {1}) — CITED
# pre-commit: lint+format (per module, incl. GOWORK=off inventory)
pre-commit:
  parallel: false
  commands:
    lint-workspace:
      glob: "*.go"
      run: |
        set -e
        for m in pkg services/analytics services/audit; do
          (cd "$m" && golangci-lint run ./...)
        done
      stage_fixed: true
    lint-inventory:
      glob: "services/inventory/**/*.go"
      # inventory is outside go.work (WIP) → GOWORK=off; best-effort, do not "fix" scaffolding
      run: cd services/inventory && GOWORK=off golangci-lint run ./...
      stage_fixed: true

# pre-push: tests for IN-WORKSPACE modules only; inventory excluded by design (D-03)
pre-push:
  parallel: false
  commands:
    test-pkg:
      run: cd pkg && go test ./...           # real ginkgo suite (pkg/http)
    test-audit:
      run: cd services/audit && go test ./...
    test-analytics:
      run: cd services/analytics && go test ./...   # 0 packages currently → no-op, not an error
    # NOTE: services/inventory intentionally NOT tested here — WIP outside go.work (boundaries.md).

# commit-msg: Conventional Commits via commitlint (D-04)
commit-msg:
  commands:
    commitlint:
      run: npx --no-install commitlint --edit {1}
```

### `package.json` (ENF-03) — private, commitlint devDeps only
```json
{
  "name": "gwall-e-tooling",
  "private": true,
  "description": "Dev-only tooling (commitlint). Not a Node app — single isolated Node dependency.",
  "devDependencies": {
    "@commitlint/cli": "21.0.2",
    "@commitlint/config-conventional": "21.0.2"
  }
}
```
> Exact-pin (no `^`/`~`). Confirm `21.0.2` is still the latest stable at plan/execution time (it was published ~7 days before this research). `private: true` prevents accidental publish.

### `commitlint.config.js` (ENF-03)
```js
// Source: commitlint.js.org getting-started — CITED
export default {
  extends: ['@commitlint/config-conventional'],
  // Optional: align with git.md scopes (phase-scoped GSD commits like docs(04): ...).
  // Conventional types already include: feat, fix, docs, refactor, test, chore, style, perf, build, ci, revert, perf.
};
```
> Filename is discretion (D-04): `commitlint.config.js` (ESM `export default`) works with Node 22 (`"type"` absent → may need `.mjs` or `commitlint.config.mjs` if CJS resolution complains; `.cjs` with `module.exports` is the safe fallback). Validate during execution.

### `buf.yaml` (ENF-04) — SKELETON, v2, NOT hooked
```yaml
# Source: buf.build/docs/configuration/v2/buf-yaml — CITED
# SKELETON: no .proto exist yet. buf is NOT wired into any lefthook hook (no-phantom).
# Activates when real schemas land under the proto root below.
version: v2
modules:
  - path: proto            # future proto root — no files yet (skeleton)
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

### `buf.gen.yaml` (ENF-04) — SKELETON, v2, pinned plugins
```yaml
# Source: buf.build/docs/configuration/v2/buf-gen-yaml — CITED
# SKELETON: pins codegen plugin versions. NOT executed until .proto exist. Do not claim working codegen.
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go:v1.36.5   # protoc-gen-go (pin exact at plan time)
    out: gen/go
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go:v1.5.1               # protoc-gen-go-grpc (pin exact at plan time)
    out: gen/go
    opt:
      - paths=source_relative
  # Optional (STACK.md): protovalidate plugin — add when validation schemas exist.
```
> **No-phantom phrasing for docs/commit:** "buf config **seeded** (lint/breaking/codegen skeleton); proto codegen **activates when `.proto` are added** — not wired into hooks this phase." Pin the exact remote plugin versions at plan time (`buf.build/protocolbuffers/go`, `buf.build/grpc/go`).

### `Makefile` (D-11) — pinned tool install
```makefile
# Source: tool versions verified vs official release pages this session — CITED
GOLANGCI_VERSION := v2.12.2
LEFTHOOK_VERSION := v2.1.9
BUF_VERSION      := v1.71.0

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	# gofumpt runs embedded inside golangci-lint v2 formatters — NOT installed standalone (Go 1.25 req).
	@echo "Next: 'npm install' (commitlint) and 'lefthook install' (git hooks)."
```
> boundaries.md flags root `Makefile` as potentially stale/non-authoritative. If a root `Makefile` already conflicts, create the tools target in a clearly-owned location and reference it from `knowledge/build.md` so the canon points to the live one.

## ENF-05 Mapping Table — every current forward mark → flipped status

> Derived from grep of `knowledge/*.md`. Each row: current mark (verbatim) → flipped status per D-07/D-08. `convention-only (review-enforced)` marks **stay unchanged** (listed for completeness where a doc author might expect a flip).

| File / Rule | Current mark | Flips to | Mechanism (this phase) |
|-------------|--------------|----------|------------------------|
| `style.md` — gofumpt (general Go style line) | `planned: hook Phase 4` | **`hook (format: gofumpt)`** | golangci `formatters.gofumpt` via pre-commit |
| `style.md` — typed IDs | `planned: CI-gated Phase 4 (linter)` | **`hook (lint)`** *or* `convention-only` | No standard linter enforces "use a named type for IDs" out of the box. **Recommend `convention-only (review-enforced)`** unless a custom rule is added (out of scope). Flag to user — see Assumptions A1. |
| `style.md` — sentinel vs `%w` | `planned: CI-gated Phase 4 (linter, напр. errorlint)` | **`hook (lint: errorlint)`** | golangci `linters.errorlint` via pre-commit |
| `style.md` — DTO→domain mapping | `planned: CI-gated Phase 4 (depguard)` | **`hook (lint: depguard, dormant)`** | depguard layer rule — dormant until layer code lands; note dormancy |
| `style.md` — RU comments / EN identifiers / EN test comments | `convention-only (review-enforced)` | **unchanged** | review |
| `architecture.md` — domain import direction (inward only) | `planned: CI-gated Phase 4 (depguard)` | **`hook (lint: depguard, dormant)`** + 1-line "becomes CI-gated when CI lands" | depguard `domain-imports-inward-only` rule (dormant) |
| `architecture.md` — domain holds only ports (no Mongo/gRPC) | `planned: CI-gated Phase 4 (depguard)` | **`hook (lint: depguard, dormant)`** | depguard (dormant) |
| `architecture.md` — MUST NOT resurrect CQRS bus / `pkg/mediatr` | `planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)` | **`hook (lint: depguard, biting)`** | depguard `no-cqrs-bus` deny — fires today |
| `architecture.md` — MUST NOT resurrect `TxManager`/`tx.go` | `planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)` | **`hook (lint: depguard, biting)`** | depguard `no-tx-manager` deny — fires on reintroduction |
| `architecture.md` — Execute interactor, query-lite, UnitOfWork, outbox/relay, PullEvents, edge validation | `convention-only (review-enforced)` | **unchanged** | review |
| `testing.md` — mockery port mocking | `planned: Phase 4 (go:generate)` | **`convention-only (review-enforced)`** + note "mockery harness deferred" | No mockery wiring this phase beyond minimal seed (CONTEXT Out of Scope / Deferred). Keep convention; do NOT claim go:generate works. Flag — Assumptions A2. |
| `testing.md` — suite bootstrap, dot-imports, spec structure, Gomega asserts | `convention-only (review-enforced)` | **unchanged** | pre-push runs the suite (`go test`), but structure rules stay review-enforced |
| `patterns.md` — all recipe rules | `convention-only (review-enforced)` | **unchanged** | review (recipes reference architecture/style) |
| `authoring.md` §"Статус enforcement" — legend itself | "В Phase 1 фиксируется только сам стандарт пометок; сами статусы проставляются ... (Phase 4 ...)" | **rewrite**: define 3 statuses precisely + rule "don't mark `CI-gated` without a CI pipeline"; change "Phase 1 fixes only the standard" → "statuses set (Phase 4)" | D-09 — single canon legend |

**Net new live enforcement after Phase 4:** gofumpt (format), gci (imports), errorlint (`%w`), depguard resurrection bans (mediatr/TxManager — biting), commitlint (commit-msg), in-workspace tests (pre-push). **Dormant (forward):** depguard layer-direction rules. **Stays review-only:** RU/EN language, interactor/query-lite/UnitOfWork/outbox conventions, spec structure, recipes, typed-IDs (recommended), mockery (deferred).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| golangci-lint v1 (`enable-all`/`disable-all`, formatters as linters) | v2 (`linters.default`, separate `formatters` block) | v2.0 (2025) | Must author config in v2 schema; gofumpt/gci live under `formatters` |
| Standalone gofumpt binary in hooks | gofumpt embedded in golangci v2 formatters | v2 formatters GA | Avoids divergent format binary + gofumpt 0.10's Go 1.25 requirement |
| Husky + lint-staged (Node) for Go repos | Lefthook (single Go binary) | — | No node_modules for hooks; commitlint stays the only Node dep |
| `tools.go` blank-import convention | Go 1.24 `go.mod tool` directive | Go 1.24 (2025) | Optional here; Makefile chosen to avoid the rotten root module |
| Raw `protoc` + shell | `buf generate` + `buf.gen.yaml` v2 | buf v1.32 (2024) | Pinned plugins, breaking detection, monorepo-aware (skeleton now) |
| buf.yaml/buf.gen.yaml v1 | v2 config (`version: v2`, `modules:` list, inputs in gen file) | buf v1.32 (2024) | Use v2 schema for the skeleton |

**Deprecated/outdated:**
- golangci-lint v1 schema — do not adopt.
- Standalone gofumpt 0.10.0 on Go <1.25 — incompatible; use embedded.
- STACK.md "Lefthook v1.x" — superseded by v2.1.9 (pin v2).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | typed-IDs rule has no off-the-shelf linter → recommend `convention-only` rather than `hook` | ENF-05 Mapping | Phase-3 mark said "CI-gated (linter)"; if user wants a custom linter for typed IDs, that's added scope. User should confirm `convention-only` is acceptable. |
| A2 | mockery stays `convention-only` (go:generate harness deferred) | ENF-05 Mapping | testing.md mark said "planned: Phase 4 (go:generate)". CONTEXT Deferred excludes full mockery wiring; confirm minimal seed (or nothing) is intended for ENF-05. |
| A3 | `no-tx-manager` depguard path `.../inventory/internal/tx` is a reasonable guard path for the removed dispatcher | Code Examples / ENF-05 | The removed `tx.go`/`TxManager` exact import path is gone from the repo (grep found no remnants); the deny path is a best-guess forward guard. Refine to the actual historical path if known, or rely on the biting `no-cqrs-bus` + review. |
| A4 | protoc-gen-go `v1.36.5` / protoc-gen-go-grpc `v1.5.1` remote plugin pins | Code Examples (buf.gen.yaml) | Skeleton only; exact plugin versions should be confirmed at plan time (root go.mod shows protobuf runtime v1.36.5). No runtime impact until `.proto` exist. |
| A5 | commitlint `21.0.2` is the exact pin | package.json | Latest at research time; confirm latest stable at plan/execution (published ~7 days prior — consider `21.0.1` for a settling buffer). |
| A6 | Per-module lint loop is preferred over a single workspace `golangci-lint run` | Architecture Pattern 1 | If planner prefers a single workspace run + a separate inventory pass, that's equally valid (D-01 discretion). The loop is clearer for the inventory boundary. |

## Open Questions

1. **Should `inventory` lint be blocking or advisory in pre-commit?**
   - What we know: inventory is WIP, outside `go.work`, may not compile; boundaries.md forbids "fixing" it.
   - What's unclear: whether a failing inventory lint should block a commit that doesn't touch inventory.
   - Recommendation: scope `lint-inventory` to fire only on `services/inventory/**/*.go` changes (as in the example glob) and treat it as best-effort; document the WIP caveat. Planner discretion (D-01).

2. **commitlint config module format under Node 22 (no `package.json` "type": "module").**
   - What we know: `commitlint.config.js` with `export default` is ESM; the private `package.json` has no `"type"` field.
   - What's unclear: whether Node resolves `.js` as CJS (→ `export default` fails) in this setup.
   - Recommendation: use `commitlint.config.mjs` (force ESM) **or** `.cjs` with `module.exports`. Validate during execution (D-04 discretion).

3. **Exact removed `TxManager`/`tx` import path for the biting depguard ban.**
   - What we know: no remnants exist in code today; PROJECT.md records the removal.
   - What's unclear: the precise old import path to deny.
   - Recommendation: keep the `no-cqrs-bus` ban (path known: `pkg/mediatr`) as the primary biting guard; make `no-tx-manager` a forward guard and confirm the historical path against git history at plan time (A3).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | all (lint/test/tool install) | ✓ | go1.24.6 darwin/arm64 | — |
| node | commitlint | ✓ | v22.20.0 | — |
| npm | commitlint install | ✓ | 10.9.3 | — |
| golangci-lint | ENF-01 lint/format | ✗ | — | `make tools` installs v2.12.2 |
| lefthook | ENF-02 hooks | ✗ | — | `make tools` installs v2.1.9 |
| buf | ENF-04 (skeleton, manual) | ✗ | — | `make tools` installs v1.71.0; not needed until `.proto` exist |
| gofumpt (standalone) | — (runs embedded in golangci) | ✗ | — | embedded in golangci v2.12.2 — do NOT install standalone (Go 1.25 req) |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** golangci-lint, lefthook, buf — all installed by the documented `make tools` bootstrap (a planned phase deliverable, D-11). No blocker.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Ginkgo v2.23.4 + Gomega v1.38.0 (pinned in `pkg/go.mod`) |
| Config file | none (standard `go test` + Ginkgo bootstrap in `pkg/http/http_test.go`) |
| Quick run command | `cd pkg && go test ./...` |
| Full suite command | per-module: `cd pkg && go test ./...`; `cd services/audit && go test ./...`; `cd services/analytics && go test ./...` |

### Phase Requirements → Test Map
> This phase ships **config + docs**, not Go code. "Tests" are config-validity checks and doc-consistency checks, plus confirming the hooks fire.

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ENF-01 | `.golangci.yml` is valid v2 and runs | smoke | `golangci-lint config verify` then `cd pkg && golangci-lint run ./...` | ✅ after `make tools` |
| ENF-01 | depguard resurrection ban bites | smoke | add a throwaway import of `pkg/mediatr`, expect lint failure (manual/scripted) | manual |
| ENF-02 | lefthook hooks installed & fire | smoke | `lefthook install` then `lefthook run pre-commit` / `pre-push` | ✅ after install |
| ENF-03 | commitlint rejects bad / accepts good message | smoke | `echo "bad message" \| npx --no-install commitlint` (fails); `echo "docs(04): x" \| npx --no-install commitlint` (passes) | ✅ after `npm install` |
| ENF-04 | buf skeleton parses (no codegen) | smoke | `buf config ls-lint-rules` / `buf build` (no `.proto` → empty, not error) | ✅ after `make tools` |
| ENF-05 | every `planned:` mark flipped; no `CI-gated` remains | grep gate | `grep -rn "planned:" knowledge/` returns nothing; `grep -rn "CI-gated" knowledge/` only in the legend definition + "becomes CI-gated when CI lands" notes | manual/grep |

### Sampling Rate
- **Per task commit:** `golangci-lint config verify` (config tasks); `grep -rn "planned:" knowledge/` (doc-flip tasks).
- **Per wave merge:** run all smoke checks above (`lefthook run pre-commit`, commitlint round-trip, `buf build`).
- **Phase gate:** all ENF-01..05 smoke/grep checks green before `/gsd-verify-work`; in-workspace `go test` green (unaffected by config-only changes).

### Wave 0 Gaps
- [ ] No test-framework install needed — Ginkgo/Gomega already pinned in `pkg/go.mod`.
- [ ] `make tools` must run before ENF-01/02/04 smoke checks (golangci-lint/lefthook/buf absent on machine).
- [ ] `npm install` must run before ENF-03 commitlint round-trip.
- [ ] No new Go test files required (config/docs phase). ENF-05 verified by grep gate, not unit tests.

## Security Domain

> `security_enforcement: true`, ASVS level 1. This phase ships dev tooling + docs — no application code paths, no auth/session/crypto/data handling. ASVS application categories largely **do not apply**; the relevant surface is supply-chain integrity of the tools introduced.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V1 Architecture | partial | depguard enforces hexagon import boundaries (defense-in-depth via layering) — dormant until code lands |
| V2 Authentication | no | no auth code |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | no (deferred) | protovalidate referenced in buf skeleton only; not active |
| V6 Cryptography | no | — |
| V14 Configuration / Supply Chain | yes | exact-pin all tool versions (Makefile + `package.json` exact); legitimacy-audit npm deps (done); prefer `npx --no-install` to avoid network fetch of an unpinned commitlint |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Slopsquatted / typosquatted npm dep | Tampering | Legitimacy audit (done — both commitlint pkgs verified: real repo, 8M dl, no postinstall); exact-pin versions |
| Unpinned tool fetches differing version (config drift / supply chain) | Tampering | Pin golangci/lefthook/buf in Makefile; commitlint exact in `package.json`; `npx --no-install` |
| Malicious postinstall script | Tampering / Elevation | Verified `postinstall: null` for both commitlint packages this session |
| Phantom "working" codegen/CI misleading agents into unsafe assumptions | Repudiation / Info | no-phantom: buf marked skeleton, no `CI-gated` without CI (D-07/D-10) |

## Sources

### Primary (HIGH confidence — official release pages fetched this session)
- github.com/golangci/golangci-lint/releases — v2.12.2 (2026-05-06), recent v2.x list. Tool version + v2 install path.
- github.com/bufbuild/buf/releases — v1.71.0 (2026-06-16). buf CLI version.
- github.com/evilmartians/lefthook/releases — v2.1.9 (2026-05-29). Lefthook version (drift correction vs STACK.md).
- github.com/mvdan/gofumpt/releases — 0.10.0 (2026-05-04) requires Go 1.25; fork of gofmt as of Go 1.26.
- Repo inspection (`go.work`, all `go.mod`, `knowledge/*.md`, `services/*`, `.gitignore`, tool availability) — direct observation of module paths (`auditlogs`!), empty `inventory/internal/*` layer dirs, absent tools, node/npm present, no mediatr/tx remnants, no `.proto`.
- `gsd-tools query package-legitimacy check` — commitlint npm verdicts (SUS=too-new false-positive).

### Secondary (MEDIUM confidence — official docs via WebFetch/WebSearch, schema-level)
- golangci-lint.run/docs — v2 `linters.default`, `formatters` block, depguard `rules`/`list-mode`/`deny` schema, errorlint settings.
- buf.build/docs/configuration/v2 — `buf.yaml` v2 (`modules:` list, lint/breaking), `buf.gen.yaml` v2 (plugins, inputs).
- lefthook.dev/configuration — `commands`/`run`/`glob`/`stage_fixed`, `{1}` commit-msg var, `parallel`.
- commitlint.js.org / npm — `--edit {file}`, config-conventional type-enum, exact-pin guidance, v21.0.2.

### Tertiary (LOW confidence — community blogs, used only for corroboration; cross-checked vs primary/secondary)
- Multi-module golangci orchestration (per-module loop vs workspace) — community patterns; no single authoritative source → marked MEDIUM in body, validate locally.
- Go 1.24 `tool` directive guidance (itnext/rednafi/bytesizego) — corroborated the directive + caveats (separate go.mod for tools).

## Metadata

**Confidence breakdown:**
- Standard stack (versions): HIGH — all four core tool versions cross-checked against official GitHub/npm release pages this session.
- Config schemas (golangci v2, lefthook, buf v2, commitlint): MEDIUM-HIGH — official docs, schema-level; exact YAML to be validated by `golangci-lint config verify` / `buf build` at execution.
- Multi-module lint/test orchestration: MEDIUM — community-corroborated, no single authoritative source; D-01 leaves mechanic to planner; validate `golangci-lint run` per module locally.
- ENF-05 mapping: HIGH — derived directly from grep of `knowledge/*.md` + D-07/D-08 rules; two rows (typed-IDs, mockery) carry explicit assumptions A1/A2 for user confirmation.
- Pitfalls: HIGH — each tied to a verified version/schema fact (v2 install path, gofumpt Go-1.25, lefthook drift, no-phantom constraints).
- Repo facts: HIGH — direct observation.

**Research date:** 2026-06-17
**Valid until:** 2026-07-17 (tool versions move fast — re-verify commitlint/golangci/lefthook/buf latest before pinning; gofumpt/Go-1.25 interaction stable).
