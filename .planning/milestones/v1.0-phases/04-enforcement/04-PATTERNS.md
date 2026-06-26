# Phase 4: Enforcement-слой (тулинг) - Pattern Map

**Mapped:** 2026-06-17
**Files analyzed:** 15 (7 NEW config/build files + 8 EDITED knowledge docs)
**Analogs found:** 8 / 15 (all 8 edited docs have a strong in-repo analog; the 7 new config files have NO in-repo code analog — their only "analog" is the verbatim RESEARCH.md snippet)

> **Honesty note (per boundaries.md no-phantom):** This phase splits cleanly into two
> classes. The **config/build files** (`.golangci.yml`, `lefthook.yml`, `package.json`,
> `commitlint.config.*`, `buf.yaml`, `buf.gen.yaml`, `Makefile`) are **net-new tooling
> with no existing analog in this repo** — there is no prior `.golangci.yml`, no `lefthook.yml`,
> no `package.json`, no `Makefile` (verified: none exist). The planner MUST copy their content
> from the RESEARCH.md §Code Examples (lines cited per file below), not from a fictional repo
> analog. The **edited knowledge docs** have a real, strong analog: the existing doc content
> itself plus the Phase 1–3 authoring standard. Do not invent a code analog for the config files.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `.golangci.yml` (NEW) | config (lint/format) | transform (lint rules over source) | RESEARCH.md §Code Examples L335-393 | research-snippet (no code analog) |
| `lefthook.yml` (NEW) | config (git hooks) | event-driven (pre-commit/pre-push/commit-msg) | RESEARCH.md §Code Examples L397-434 | research-snippet (no code analog) |
| `package.json` (NEW) | config (dep manifest) | batch (npm install) | RESEARCH.md §Code Examples L437-447 | research-snippet (no code analog) |
| `commitlint.config.js`/`.mjs`/`.cjs` (NEW) | config (commit lint) | transform (validate commit msg) | RESEARCH.md §Code Examples L451-458 | research-snippet (no code analog) |
| `buf.yaml` (NEW, skeleton) | config (proto lint/breaking) | transform (proto lint) | RESEARCH.md §Code Examples L462-475 | research-snippet (no code analog) |
| `buf.gen.yaml` (NEW, skeleton) | config (codegen) | batch (codegen, inert) | RESEARCH.md §Code Examples L478-492 | research-snippet (no code analog) |
| `Makefile` (NEW) | build (tool pinning) | batch (`make tools` install) | RESEARCH.md §Code Examples L496-509 | research-snippet (no code analog) |
| `knowledge/authoring.md` (EDIT) | docs (canon legend) | transform (status legend rewrite) | `knowledge/authoring.md` §"Статус enforcement" L64-68 (self) + Phase 1–3 authoring std | exact (self-edit) |
| `knowledge/style.md` (EDIT) | docs (mark flip) | transform (ENF-05 flips) | existing marks L14, L53, L72, L90 (self) | exact (self-edit) |
| `knowledge/architecture.md` (EDIT) | docs (mark flip) | transform (ENF-05 flips) | existing marks L26-27, L30, L151, L154 (self) | exact (self-edit) |
| `knowledge/testing.md` (EDIT) | docs (mark flip) | transform (ENF-05 flip) | existing marks L111, L114, L126 (self) | exact (self-edit) |
| `knowledge/patterns.md` (EDIT, likely no-op) | docs (mark verify) | transform (ENF-05 verify) | `convention-only` marks L31, L87, L118, L154 (self) | exact (self-edit) |
| `knowledge/boundaries.md` (EDIT) | docs (ownership map) | transform (register enforcement facts) | ownership-map table L61-70 (self) | exact (self-edit) |
| `knowledge/build.md` (EDIT) | docs (bootstrap) | transform (add bootstrap recipe) | recipe sections L16-45 (self) | exact (self-edit) |
| `knowledge/README.md` (EDIT) | docs (index) | transform (index/status update) | index table L23-35 (self) | exact (self-edit) |

**Match quality legend:**
- `research-snippet` = no in-repo code analog exists; planner copies the exact target content from RESEARCH.md §Code Examples (honest: these files are net-new).
- `exact (self-edit)` = the file being edited is its own best analog — the existing content + mark sites are already present; ENF-05 flips them in place per the authoring standard.

## Pattern Assignments

### NEW config/build files — no code analog; copy from RESEARCH.md verbatim

> The seven files below have **no precedent in this repo**. The planner MUST NOT search for
> a code analog (there is none). The authoritative source content is the RESEARCH.md §Code
> Examples block, already version-verified this session. Lines cited are in
> `.planning/phases/04-enforcement/04-RESEARCH.md`.

#### `.golangci.yml` (config, transform) — ENF-01

**Analog:** RESEARCH.md §Code Examples L335-393 (full v2 config, version-verified).

**Source content to copy:** `version: "2"`; `linters.default: standard` + `enable: [errorlint, depguard]`; `depguard.rules` with **biting** `no-cqrs-bus` / `no-tx-manager` deny blocks and **dormant** `domain-imports-inward-only`; `formatters.enable: [gofumpt, gci]` with `gci.sections` `prefix(github.com/gwall-e)`.

**Critical constraints (RESEARCH Pitfalls):**
- Install path MUST include `/v2/` (Pitfall 1, L282-286).
- v2 schema only — `linters.default`, NOT `enable-all`/`disable-all`; formatters in their own top-level block (Pitfall 2, L288-292).
- gofumpt runs **embedded** in `formatters`, never standalone (Pitfall 4, L300-304).
- depguard layer rules are **dormant** (path selectors match no files yet) — document in rule `desc`, do NOT claim active enforcement (D-05; Pattern 2 L227-231).
- `no-cqrs-bus` denies `github.com/gwall-e/pkg/mediatr`; `no-tx-manager` path is a forward best-guess (Assumption A3 L557) — confirm against git history at plan time.

#### `lefthook.yml` (config, event-driven) — ENF-02

**Analog:** RESEARCH.md §Code Examples L397-434.

**Source content to copy:** `pre-commit` (`parallel: false`, `lint-workspace` per-module loop over `pkg services/analytics services/audit` + `lint-inventory` with `GOWORK=off`, both `stage_fixed: true`); `pre-push` (`test-pkg`/`test-audit`/`test-analytics`, **inventory explicitly excluded** with NOTE comment); `commit-msg` (`npx --no-install commitlint --edit {1}`).

**Critical constraints:**
- Lefthook **v2.1.9**, NOT v1.x — STACK.md drift corrected (Pitfall 3, L294-298). Schema unchanged from v1.
- `inventory` excluded from pre-push by design (D-03) — comment it as intentional, not a gap (Pitfall 5, L306-310).
- `analytics` `go test ./...` matches 0 packages → no-op, not an error (L426, build.md L28-31).
- buf is NOT wired into any hook (D-10; anti-pattern L248).
- `inventory` pre-commit lint is best-effort/advisory WIP (Open Question 1 L564-567) — planner decides blocking vs advisory.

#### `package.json` (config, batch) — ENF-03

**Analog:** RESEARCH.md §Code Examples L437-447.

**Source content to copy:** `{ "name": "gwall-e-tooling", "private": true, "devDependencies": { "@commitlint/cli": "21.0.2", "@commitlint/config-conventional": "21.0.2" } }`.

**Critical constraints:**
- `private: true` + **exact pins** (no `^`/`~`) — supply-chain hardening (Security L641, Audit L128-138).
- commitlint is the **single** Node dependency — keep isolated, do not expand Node tooling (D-04).
- Confirm `21.0.2` is still latest stable at plan time; `21.0.1` acceptable as a settling buffer (Assumption A5 L559).

#### `commitlint.config.js` / `.mjs` / `.cjs` (config, transform) — ENF-03

**Analog:** RESEARCH.md §Code Examples L451-458.

**Source content to copy:** `extends: ['@commitlint/config-conventional']`.

**Critical constraints:**
- Filename + module format is discretion (D-04). Node 22, no `"type": "module"` in package.json → `.js` may resolve as CJS and break `export default`. Use `.mjs` (force ESM) or `.cjs` with `module.exports`; validate at execution (Open Question 2 L569-572).
- Config-conventional type-enum already covers the GSD `docs(NN):` commits used in git.md L23 / L57-58.

#### `buf.yaml` (config, transform — SKELETON) — ENF-04

**Analog:** RESEARCH.md §Code Examples L462-475.

**Source content to copy:** `version: v2`; `modules: [{ path: proto }]`; `lint.use: [STANDARD]`; `breaking.use: [FILE]`.

**Critical constraints:**
- **Skeleton** — no `.proto` exist; NOT wired into any hook (D-10). Header comment MUST say so.
- Phrase docs/commit as "buf config **seeded**; codegen **activates when `.proto` are added**" — never "working codegen" (Pitfall 7, L318-322).

#### `buf.gen.yaml` (config, batch — SKELETON) — ENF-04

**Analog:** RESEARCH.md §Code Examples L478-492.

**Source content to copy:** `version: v2`; pinned remote plugins `buf.build/protocolbuffers/go:v1.36.5` (protoc-gen-go) and `buf.build/grpc/go:v1.5.1` (protoc-gen-go-grpc), each `out: gen/go`, `opt: [paths=source_relative]`; optional protovalidate.

**Critical constraints:**
- Plugin versions pinned exact; confirm at plan time (root go.mod shows protobuf runtime v1.36.5) (Assumption A4 L558).
- Inert until `.proto` exist — no-phantom (D-10).

#### `Makefile` (build, batch) — D-11

**Analog:** RESEARCH.md §Code Examples L496-509.

**Source content to copy:** version vars `GOLANGCI_VERSION := v2.12.2`, `LEFTHOOK_VERSION := v2.1.9`, `BUF_VERSION := v1.71.0`; `.PHONY: tools` target installing golangci-lint (`/v2/` path), lefthook (`/v2`), buf; echo next-steps for `npm install` + `lefthook install`.

**Critical constraints:**
- Do NOT add a `tool` block to the root `github.com/gwall-e` go.mod (rotten leftover, ginkgo v1/mongo v1, outside go.work) — Makefile or dedicated `tools/go.mod` instead (anti-pattern L249, Alternatives L100).
- Do NOT `go install gofumpt` standalone (Go 1.25 requirement) — embedded in golangci (Pitfall 4).
- If a root `Makefile` conflicts with a stale one, place the target in a clearly-owned location and point build.md at the live one (RESEARCH note L510). **Verified: no root `Makefile` exists currently.**
- Mechanism (Makefile vs `go.mod tool` vs mise) is D-11 discretion.

---

### EDITED knowledge docs — analog is the existing content + Phase 1–3 authoring standard

> **Shared analog for ALL doc edits:** `knowledge/authoring.md` (the authoring standard itself).
> Every flip MUST preserve: the `⟶ ` mark prefix, MUST/SHOULD/WON'T force tags, the
> "запрет → do" pairing, and pointer-over-copy (no duplicated facts). The marks already exist
> (D-11 forward marks from Phase 3) — ENF-05 **flips the status in place**, it does not rewrite
> the rules. The ENF-05 Mapping Table (RESEARCH L512-533) is the authoritative per-mark instruction.

#### `knowledge/authoring.md` (docs, transform) — ENF-05 / D-09 — CANON LEGEND

**Analog:** the existing §"Статус enforcement" (self), `knowledge/authoring.md` L64-68.

**Existing content to rewrite** (L66-68):
```text
**SHOULD**: где правило механизируемо, помечать его статусом проверки — `CI-gated`,
`hook` или `convention-only`. В Phase 1 фиксируется только сам стандарт пометок; сами
статусы проставляются в topic-доках и enforcement-конфигах (Phase 4, задел под ENF-05).
```

**Flip instruction** (RESEARCH Mapping L531, D-09 L100-103):
- Define the **3 statuses precisely**: `hook` (checked by a committed git hook today), `convention-only` (review-enforced), `CI-gated` (**reserved** until a CI pipeline exists).
- Add the rule: **do NOT mark `CI-gated` without a CI pipeline** (Pitfall 8 L324-328).
- Change "В Phase 1 фиксируется только сам стандарт" → "статусы проставлены (Phase 4)".
- This is the **single canon** legend; topic docs reference it, do not redefine (progressive disclosure, D-09). Keep the **MUST/SHOULD** force-tag style of the surrounding section.

#### `knowledge/style.md` (docs, transform) — ENF-05

**Analog:** existing mark sites (self). Four marks to act on:

| Line | Existing mark (verbatim) | Flip to (RESEARCH Mapping L518-522) |
|------|--------------------------|-------------------------------------|
| L14 | `gofumpt **(planned: hook Phase 4)**` (inline in prose) | `hook (format: gofumpt)` |
| L53 | `⟶ planned: CI-gated Phase 4 (linter)` (typed IDs) | **`convention-only (review-enforced)`** — no off-the-shelf linter for typed-ID rule (Assumption A1 L555); flag to user |
| L72 | `⟶ planned: CI-gated Phase 4 (linter, напр. errorlint)` (`%w`) | `hook (lint: errorlint)` |
| L90 | `⟶ planned: CI-gated Phase 4 (depguard)` (DTO→домен) | `hook (lint: depguard, dormant)` — note dormancy |

**Mark-format excerpt to match** (existing style, L52-53):
```text
  именованный тип, который компилятор проверяет. ⟶ planned: CI-gated Phase 4 (linter)
```
Keep the trailing `⟶ <status>` form; only the status text changes. The `convention-only (review-enforced)` marks at L30, L33-37, L68 **stay unchanged**. Also update the prose preamble L10-11 ("Phase 4 переключит её на фактический") if needed to read as done.

#### `knowledge/architecture.md` (docs, transform) — ENF-05

**Analog:** existing mark sites (self). Four marks to flip:

| Line | Existing mark (verbatim) | Flip to (RESEARCH Mapping L523-526) |
|------|--------------------------|-------------------------------------|
| L26-27 | `⟶ planned: CI-gated Phase 4 (depguard)` (domain inward-only) | `hook (lint: depguard, dormant)` + **one line** "становится CI-gated при появлении CI" (D-08) |
| L30 | `⟶ planned: CI-gated Phase 4 (depguard)` (domain holds only ports) | `hook (lint: depguard, dormant)` |
| L151 | `⟶ planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)` (CQRS bus) | `hook (lint: depguard, biting)` — fires today |
| L154 | `⟶ planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)` (TxManager) | `hook (lint: depguard, biting)` — fires on reintroduction |

**Biting-ban mark excerpt to match** (existing, L150-151):
```text
  inbound-адаптер (`api`/`cron`) зовёт нужный use case **напрямую**.
  ⟶ planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)
```
The `convention-only (review-enforced)` marks (L65, L97, L106, L116, L122, L138) **stay unchanged**. The Manual-Only review note at L55-57 stays.

#### `knowledge/testing.md` (docs, transform) — ENF-05

**Analog:** existing mark sites (self). One mark to flip + phrasing guard:

| Line | Existing mark (verbatim) | Flip to (RESEARCH Mapping L528) |
|------|--------------------------|----------------------------------|
| L111 | `⟶ planned: Phase 4 (go:generate)` (mockery) | **`convention-only (review-enforced)`** + note "mockery harness deferred" (Assumption A2 L556) |

**Mark excerpt to match** (existing, L108-111):
```text
- **SHOULD** мокать порты через **mockery** ... ⟶ planned: Phase 4 (go:generate)
```
- Full mockery wiring is **out of scope / Deferred** (CONTEXT L37-38, L242-243). Keep the convention; do NOT claim `go:generate` works.
- L112-116 already enforce no-phantom ("не выдавать mockery за уже настроенный") and L126 illustration comment ("planned: Phase 4") — keep these honest; update "planned: Phase 4" phrasing so it does not imply this phase wired it. The `convention-only` marks at L21, L24, L28, L31, L56, L58, L63 **stay unchanged** (pre-push runs the suite via `go test`, but structure rules stay review-enforced).

#### `knowledge/patterns.md` (docs, transform — likely NO-OP) — ENF-05

**Analog:** existing marks (self). All four marks (L31, L87, L118, L154) are already `⟶ convention-only (review-enforced)` and **stay unchanged** (RESEARCH Mapping L530 — recipes reference architecture/style). Listed for completeness; planner should **verify no `planned:` remains** (grep gate) and otherwise leave the file untouched.

#### `knowledge/build.md` (docs, transform) — D-11 BOOTSTRAP

**Analog:** existing recipe sections (self), L16-45 (the `cd pkg && go test`, `GOWORK=off` recipes).

**Edit:** Add a **bootstrap section** documenting the one-time tooling setup, matching the existing recipe-level style (MUST/WON'T, no fragile listings, pointer-over-copy):
- `make tools` (Go tools — golangci-lint v2.12.2, lefthook v2.1.9, buf v1.71.0; cross-ref Makefile)
- `npm install` (commitlint)
- `lefthook install` (activate git hooks)

**Style excerpt to match** (existing build.md recipe form, L18-22):
```text
**MUST** прогонять тесты общей библиотеки из её модуля:

`cd pkg && go test ./...`

Это рабочий рецепт: в `pkg` есть тесты (`pkg/http`), прогон зелёный.
```
- Keep build.md the **canon** for commands (boundaries.md ownership L64). Document `inventory` exclusion from pre-push as intentional (D-03). The bootstrap install commands are in RESEARCH §Installation L105-119.

#### `knowledge/README.md` (docs, transform) — INDEX

**Analog:** existing index table (self), L23-35, and "Что где живёт" L10-19.

**Edit:** If the planner decides enforcement configs warrant index visibility, update status/notes consistent with the existing table. Index links MUST point only to existing files (L37-38 rule). The new root config files (`.golangci.yml` etc.) live in repo root, NOT `knowledge/` — reference them only if it adds agent value, honoring "ссылки только на реально существующие файлы".

---

## Shared Patterns

### Authoring standard (applies to ALL doc edits)

**Source:** `knowledge/authoring.md` (the canon standard, L1-68).
**Apply to:** authoring.md, style.md, architecture.md, testing.md, patterns.md, boundaries.md, build.md, README.md.

```text
# from authoring.md — every rule carries exactly one force tag; every ban pairs with a "do"
- **MUST** — твёрдый инвариант.
- **SHOULD** — сильная рекомендация (отклонение обосновано в PR).
- **WON'T** — явный запрет.
Формула: не «не делай X», а «делай Y; X — WON'T, потому что Z».
```
Preserve force tags, `⟶ <status>` mark form, "запрет → do" pairing, pointer-over-copy. ENF-05 flips **status text only**, never rewrites the rule body.

### No-phantom / pointer-over-copy (cross-cutting honesty)

**Source:** `knowledge/boundaries.md` (§Никаких phantom-фич L31-38; ownership-map L55-70) + `knowledge/authoring.md` §"Никаких phantom-правил" L53-62.
**Apply to:** buf skeleton (D-10), depguard dormant rules (D-05), mockery (A2), every `hook` vs `CI-gated` decision (D-07).

```text
# from authoring.md L53-62
Описывать поведение несуществующих подсистем, фич или файлов — **WON'T**; ...
база документирует только то, что реально есть в репозитории сейчас.
```
Concretely: never label `CI-gated` (no CI exists); never present buf codegen as working; document `inventory` exclusion as intentional, not a gap.

### Ownership-map registration (boundaries.md)

**Source:** `knowledge/boundaries.md` §"Карта владения фактами" L55-70 (table `| Факт | Канон | Статус |`).
**Apply to:** register the new enforcement facts honestly in the table.

```text
| Факт | Канон | Статус |
|------|-------|--------|
| WIP-статус `inventory`, членство `go.work`, раскладка модулей | structure.md | существует |
| Команды сборки/запуска/тестов (вкл. `GOWORK=off`) | build.md | существует |
```
Add rows for: enforcement-status taxonomy → authoring.md; tool bootstrap/commands → build.md; the `inventory` pre-push exclusion fact. Keep "один факт = один канон-док"; reference, do not copy. New rows MUST carry honest "существует" status (the config files do exist after this phase; buf codegen does NOT — phrase accordingly).

### Conventional Commits scope (for the phase's own commits)

**Source:** `knowledge/git.md` §Conventional Commits L18-58.
**Apply to:** all commits in this phase + the commitlint config consistency check.

```text
- **MUST** ... формате Conventional Commits: `type(scope): subject`.
- В GSD-коммитах scope — phase-scoped вида `docs(NN):` (напр. `docs(02):`).
```
GSD phase commits use `docs(04): ...` / `feat(04): ...` / `chore(04): ...`; config-conventional's type-enum already accepts these. The commitlint config (ENF-03) MUST NOT reject the repo's own commit style.

## No Analog Found

The seven NEW config/build files have **no in-repo code analog** — by design, this phase introduces the tooling for the first time. Their authoritative source is RESEARCH.md §Code Examples (verbatim, version-verified), NOT a repo precedent.

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `.golangci.yml` | config | transform | No prior golangci config; v2 schema is net-new. Copy RESEARCH L335-393. |
| `lefthook.yml` | config | event-driven | No prior git-hook manager config. Copy RESEARCH L397-434. |
| `package.json` | config | batch | No prior Node manifest in this Go repo. Copy RESEARCH L437-447. |
| `commitlint.config.*` | config | transform | No prior commit-lint config. Copy RESEARCH L451-458. |
| `buf.yaml` | config | transform | No `.proto` / prior buf config. Skeleton — copy RESEARCH L462-475. |
| `buf.gen.yaml` | config | batch | No prior codegen config. Skeleton — copy RESEARCH L478-492. |
| `Makefile` | build | batch | No prior root Makefile (verified absent). Copy RESEARCH L496-509. |

**For these, the planner uses RESEARCH.md patterns directly** (this is the intended path — RESEARCH §Code Examples are "target artifacts to create", L332).

## Metadata

**Analog search scope:** `knowledge/*.md` (10 docs), repo root (config/build file presence), `go.work`, RESEARCH.md §Code Examples + §ENF-05 Mapping Table.
**Files scanned:** 10 knowledge docs (read: authoring, style, architecture marks, testing marks, build, README, boundaries ownership-map; grepped: patterns), repo-root config presence check, git.md commit format.
**Key verification:** No root `.golangci.yml`/`lefthook.yml`/`package.json`/`Makefile`/`buf.*`/`commitlint.config.*` currently exist (confirmed). No `.claude/skills/` or `.agents/skills/` present.
**Pattern extraction date:** 2026-06-17
