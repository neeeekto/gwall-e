# Project Research Summary

**Project:** gwall-e — DC Hardware-as-a-Service Platform
**Domain:** AI-agent / team conventions knowledge base (milestone 1) + Go DDD/hexagonal microservices platform (future epics)
**Researched:** 2026-06-17
**Confidence:** HIGH

## Executive Summary

The first milestone of gwall-e is not a business feature — it is a **conventions and knowledge base** (`knowledge/` directory) for AI coding agents and the team. This knowledge base is the foundation every future service epic depends on: it encodes architecture decisions, coding conventions, build commands, testing standards, and domain language so agents and humans work consistently. Without it, each AI-assisted epic risks reinventing or contradicting settled decisions. The knowledge base is structured as a lean entry-point file (`CLAUDE.md`) that points to per-topic Markdown files in `knowledge/` — progressive disclosure that avoids blowing the model's context budget.

The target service architecture for the platform (which the knowledge base documents and future epics implement) is DDD + hexagonal (ports and adapters), WITHOUT a CQRS bus or dispatcher. The old `pkg/mediatr` / `CommandDispatcher` / `QueryDispatcher` / `TxManager` have been deleted. Write-side use cases are plain structs with an `Execute` method; read-side uses query services that read MongoDB directly into DTOs (CQRS-lite); transactions are managed via a `UnitOfWork` port; domain events reach downstream via a transactional outbox inside the UnitOfWork transaction, with an async relay. gRPC (grpc-go) is the inbound transport; MongoDB driver v2 is the persistence layer.

The primary risk for milestone 1 is over-documentation: stuffing too much into a single file (context budget exhaustion), encoding stale file paths (they drift in an AI-assisted codebase), documenting the current broken scaffolding as intended design, or writing phantom rules for features that do not yet exist. The mitigation is strict scope discipline: the knowledge base documents only what is real and stable now, every rule carries explicit MUST/SHOULD/WON'T strength, every prohibition is paired with the prescribed alternative, and the root `CLAUDE.md` stays under ~150 lines.

## Key Findings

### Recommended Stack

The platform is a Go multi-module workspace (`go.work`, Go 1.24.6) with per-service `go.mod` files. The `inventory` service is intentionally excluded from `go.work` and must be built with `GOWORK=off`. The milestone-1 deliverable is the `knowledge/` directory and enforcement tooling; the platform stack below is what the conventions standardize for all future services.

**Core technologies:**
- **Go 1.24.6 + go.work workspace**: multi-module monorepo; `inventory` excluded by design
- **grpc-go v1.81.x + go-grpc-middleware/v2 v2.3.3**: inbound/outbound RPC; mandate interceptor chain order (logging first, recovery last)
- **mongo-driver v2 (v2.7.x)**: persistence outbound adapter — MUST migrate off v1 (deprecated since Jan 2025); repo currently pins v1.17.9
- **buf v1.71.x + protovalidate-go**: Protobuf build/lint/codegen; replaces raw `protoc`; protovalidate replaces deprecated `protoc-gen-validate`
- **golangci-lint v2.12.x + gofumpt + gci**: primary convention enforcer; v2 config schema (`linters.default`); do not port v1 config
- **Lefthook v1.x**: Git hooks manager; preferred over Husky (no Node runtime); pre-commit: lint/format; pre-push: tests; commit-msg: commitlint
- **commitlint + config-conventional**: Conventional Commits; the only Node tooling in the repo
- **Ginkgo v2.23.4 + Gomega v1.38.0**: mandated test framework; bump root module's stale Ginkgo v1 + Gomega v1.37.0
- **log/slog (stdlib)**: structured logging interface; handler chosen at composition root only
- **CLAUDE.md + AGENTS.md**: both thin pointers to `knowledge/`; AGENTS.md is the cross-tool open standard (AAIF/Linux Foundation, 2025)

**What NOT to use:** mongo-driver v1, protobuf APIv1, grpc-middleware v1, protoc-gen-validate, Ginkgo v1, go-chi/chi as primary transport, raw protoc + shell scripts, golangci-lint v1 config.

### Expected Features

**Must have (table stakes — milestone 1 is incomplete without these):**
- `knowledge/README.md` index + reading order
- Build/run/test commands (incl. `GOWORK=off`, `cd pkg && go test`, `npx nx`)
- Repo structure / `go.work` layout
- Testing conventions (Ginkgo + Gomega; test comments in English)
- Code style and conventions (typed IDs; sentinel vs wrapped errors; DTO-mapping-in-handler; MUST/SHOULD/WON'T on every rule; canonical home for Russian-comment rule)
- Architecture rules (DDD + hexagonal; NO CQRS bus; usecases-interactor, query-lite, UnitOfWork, transactional outbox, aggregate/event rules)
- Git/workflow conventions
- Glossary / ubiquitous language (EN/RU mapping)
- Boundaries / "do not" rules (WIP scaffolding; stale README/Makefile/compose not authoritative)
- Root `CLAUDE.md` trimmed to lean index (~80–150 lines)

**Should have (add after v1 validated):**
- ADRs seeded from PROJECT.md key decisions (dropped CQRS, UnitOfWork+outbox, Russian comments, inventory out of go.work)
- Pattern catalog / recipes ("how to add a use case / query / aggregate")
- Anti-pattern catalog (anemic domain, transport DTOs in domain, reviving CQRS dispatcher)
- `AGENTS.md` cross-tool mirror
- Maintenance/self-update protocol

**Defer (future platform epics):** host inventory, SSH access grants, host lifecycle actions, bulk operations, auto-repair, monitoring, reference-service walkthrough (blocked until `inventory` builds)

**Anti-features (never add to knowledge base):** exhaustive file-path inventories, phantom rules for unbuilt features, prose philosophy without actionable rules, mega single-file.

### Architecture Approach

Two-layer architecture: (1) the knowledge base itself uses thin entry-points + progressive disclosure into `knowledge/*.md`; (2) the service architecture the base documents is DDD + hexagonal WITHOUT a CQRS bus. Absolute dependency rule: everything points inward to `domain/`; `domain/` imports nothing from the project. Inbound adapters call use cases directly via `Execute`. Writes flow through `uow.Do()` which atomically saves the aggregate and appends domain events to the transactional outbox; an async relay publishes them. Reads bypass aggregates entirely (CQRS-lite).

**Major components:**
1. `domain/` — aggregates (private fields, factory `NewX`, invariants), value objects, typed IDs, domain events, port interfaces (`Repository`, `UnitOfWork`, `EventPublisher`)
2. `usecases/` — write-side; 1 use case = 1 struct + `Execute(ctx, in) (out, error)`; wraps all writes in `uow.Do()`
3. `query/` — read-side; query services read MongoDB directly into DTOs; no aggregate hydration
4. `repositories/` — Mongo implementations of domain ports + `UnitOfWork` (transaction in `ctx`) + transactional outbox; relay
5. `api/grpc/` — inbound adapter; protovalidate at the edge
6. `cron/` — inbound adapter; scheduled jobs calling use cases directly
7. `app/` — composition root; manual DI
8. `cmd/` — thin `main`, graceful shutdown

### Critical Pitfalls

1. **Mega-file blowing the context budget** — root `CLAUDE.md` under ~150 lines; split topic files at ~200 lines; progressive disclosure. Models reliably follow ~100–150 instruction slots; rules beyond that are silently dropped.
2. **Comment-language rule not encoded canonically** — RESOLVED: Russian comments + domain terminology; English identifiers + test comments. Prevention: encode in exactly ONE doc (`knowledge/style.md`) as MUST; all others link to it; capture as ADR.
3. **Brittle file-path inventories that drift** — describe stable capabilities and layout patterns, not specific source paths. Agents discover paths just-in-time.
4. **Codifying `nil`/`TODO` scaffolding as intentional design** — `inventory` deliberately does not build; stale README/Makefile/docker-compose.yml are not authoritative. One explicit boundary rule: do NOT extend, complete, or fix the scaffolding.
5. **Ambiguous requirement strength and bare prohibitions** — every normative rule MUST carry MUST/SHOULD/WON'T; every prohibition MUST be paired with the prescribed alternative.
6. **Reviving the CQRS bus/dispatcher** — `pkg/mediatr`, `CommandDispatcher`, `QueryDispatcher`, `TxManager` are deleted and invalid. Settled; encode as MUST NOT and an ADR.
7. **Phantom rules for unbuilt features** — knowledge base documents only what exists now plus the stable glossary.

## Implications for Roadmap

### Phase 1: Knowledge Base Layout and Entry Points

**Rationale:** The structural decision (lean root file + progressive-disclosure `knowledge/` graph) must come first; getting it wrong requires painful restructuring. The authoring standard (MUST/SHOULD/WON'T, pair prohibitions with dos) must be set here so all subsequent docs follow it.

**Delivers:** `knowledge/` directory; `knowledge/README.md` index; root `CLAUDE.md` trimmed to lean index; authoring standard encoded; `AGENTS.md` stub
**Avoids:** context budget exhaustion; monolith CLAUDE.md; ambiguous rule strength across all later docs

### Phase 2: Stable Foundation Docs (Glossary, Structure, Build, Git, Boundaries)

**Rationale:** Glossary is the DDD prerequisite — architecture and style docs are incoherent without shared terms. Build commands are the highest-frequency agent content. Boundaries prevent the most expensive agent mistakes and are cheap to write.

**Delivers:** `knowledge/glossary.md` (ubiquitous language, EN/RU); `knowledge/structure.md` (`go.work`, in/out modules, `inventory` WIP status); `knowledge/build.md` (`GOWORK=off`, `cd pkg`, `npx nx`); `knowledge/git.md` (branches, Conventional Commits, PRs); `knowledge/boundaries.md` (do-not rules)
**Avoids:** agents navigating to stale paths; agents "fixing" intentional WIP; wrong branch/push behavior

### Phase 3: Convention and Architecture Docs

**Rationale:** Style (Russian-comment rule, error conventions, typed IDs) and architecture (layer rules, UnitOfWork, outbox) are the most complex docs and depend on the glossary being settled. Architecture must explicitly state NO CQRS bus.

**Delivers:** `knowledge/style.md` (canonical Russian-comment MUST; typed IDs; sentinel vs wrapped; DTO-mapping-in-handler); `knowledge/testing.md` (Ginkgo v2 + Gomega; English test comments); `knowledge/architecture.md` (DDD + hexagonal; layer import rules; use-case interactor; CQRS-lite query; UnitOfWork; transactional outbox; aggregate factory + PullEvents; MUST NOT revive CQRS bus); ADR for Russian-comment decision
**Avoids:** comment-language drift; architecture violations; test framework mismatch; CQRS bus revival

### Phase 4: Enforcement Layer (Tooling and CI)

**Rationale:** Documented rules without machine enforcement drift. Each mechanizable rule must name its tool or be tagged "convention-only." This is what makes the knowledge base reliably effective rather than aspirational.

**Delivers:** `.golangci.yml` (golangci-lint v2; gofumpt formatter; gci import ordering); `lefthook.yml` (pre-commit: lint+format; pre-push: tests; commit-msg: commitlint); `commitlint.config.js`; `buf.yaml` + `buf.gen.yaml` skeleton; CI step documentation; each `knowledge/*.md` rule tagged with enforcement status
**Avoids:** format/style drift despite documented rules; false confidence in unenforced invariants

### Phase Ordering Rationale

- Layout before content: structural split must be decided before writing content
- Glossary before architecture: DDD rules require shared terms
- Structure before build: build commands only make sense once workspace layout is documented
- Conventions before enforcement: you can only write lint rules for conventions that are defined
- Boundaries early: "do not fix scaffolding" is cheap and prevents expensive mistakes from day 1

### Research Flags

**All Phase 1–4 patterns are standard (skip research-phase during planning).** Architecture is settled in PROJECT.md Key Decisions; tool choices are verified in STACK.md; pitfalls are corroborated by high-confidence sources. Future platform epics (completing `inventory`, SSH access, bulk ops, auto-repair) will each need a research-phase during their planning.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Official release pages verified; repo directly inspected; versions cross-checked |
| Features | HIGH | Foundation layer grounded in Anthropic docs, GitHub 2,500-repo study, Cline docs; platform vision layer is MEDIUM |
| Architecture | HIGH | Grounded in locked PROJECT.md Key Decisions; textbook hexagonal + transactional outbox |
| Pitfalls | HIGH | Most pitfalls corroborated by multiple HIGH sources |

**Overall confidence:** HIGH

### Gaps to Address

- **`AGENTS.md` content:** Empirical evidence of agent-behavior improvement is early/mixed. Treat as LOW-cost addition (thin pointer) until team adopts a second agent.
- **Reference-service walkthrough:** Deferred — `inventory` does not build. Unlock when the service compiles.
- **Enforcement completeness:** CI pipeline runner not yet defined. Tag unmechanizable rules as "convention-only + review-enforced."
- **ADR format:** No template yet. Seed from PROJECT.md Key Decisions in Phase 3; keep lightweight.

## Sources

### Primary (HIGH confidence)
- `.planning/PROJECT.md` — Key Decisions, architecture constraints, comment-language resolution
- Root `CLAUDE.md` — current conventions, memory-bank references
- Repo inspection (`go.work`, all `go.mod`, deleted CQRS artifacts) — current state
- pkg.go.dev / GitHub official releases — grpc-go v1.81.1, golangci-lint v2.12.2, buf v1.71.0, ginkgo v2.23.4, mongo-driver v2 GA
- Anthropic Claude Code docs — memory model, CLAUDE.md best practices, context budget
- Cline Docs — Memory Bank best practices
- GitHub Blog — *How to write a great agents.md: lessons from 2,500+ repos*

### Secondary (MEDIUM confidence)
- Augment Code — don'ts-without-dos degrade behavior; bloat hurts
- agentsmd.io / AAIF / Linux Foundation — AGENTS.md standard
- Bijit Ghosh / Medium — CLAUDE.md guide (~150 instruction slots, pruning test)
- buf.build/blog — protovalidate v1.0 GA, PGV deprecation
- Community consensus on Wire vs Fx vs manual DI; Lefthook vs Husky

---
*Research completed: 2026-06-17*
*Ready for roadmap: yes*
