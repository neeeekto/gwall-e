# Feature Research

**Domain:** AI-agent / team conventions knowledge base (FOUNDATION milestone) + DC Hardware-as-a-Service platform (future vision, background context)
**Researched:** 2026-06-17
**Confidence:** HIGH (foundation layer, corroborated by AGENTS.md / Claude Code / Cline guidance); MEDIUM (platform-vision layer, based on domain knowledge of DCIM/HaaS/bare-metal-provisioning ecosystems)

> **Scope note.** The FIRST milestone is ONLY the `knowledge/` (memory-bank-style) convention base for AI agents and the team. That is the PRIMARY focus below. The HaaS/DC platform features are FUTURE epics, included as background to size the knowledge base and avoid over-documenting features that don't exist yet.

---

## PART A — FOUNDATION: AI-Conventions Knowledge Base (PRIMARY)

The deliverable is a set of versioned Markdown rule documents that (a) load into AI coding agents (Claude Code reads `CLAUDE.md`; the broader ecosystem reads `AGENTS.md`) and (b) serve as the team's onboarding/convention reference. The existing root `CLAUDE.md` already points at `memory-bank/` (now `knowledge/`), so the knowledge base is the detailed body that `CLAUDE.md` links into.

### Core design principles (constrain every doc below)

- **Lean entry point, progressive disclosure.** Root agent file (`CLAUDE.md`/`AGENTS.md`) stays high-signal (~80–200 lines); detail lives in linked `knowledge/*.md` files loaded on demand. Models reliably follow ~100–150 instruction "slots"; over-stuffing dilutes attention.
- **Describe capabilities and stable domain concepts, not brittle file paths.** Paths drift fast in an AI-assisted codebase; domain language and rules are stable.
- **Each doc is rule-shaped and verifiable** ("X because Y", "do A not B"), not narrative prose.
- **One canonical source per fact** — link, don't duplicate (the existing `CLAUDE.md` already imports memory-bank docs via `@`-style references).

### Table Stakes (the foundation is incomplete without these)

| Feature (rule doc) | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Entry point / index** (`knowledge/README.md`) | Agents and humans need a map; root `CLAUDE.md` must link here | LOW | Index + 1-line purpose per doc; defines reading order. Dependency root for all others |
| **Build / run / test commands** (`build.md` or in structure doc) | The single most-used agent content; project has non-obvious `GOWORK=off` for `inventory` and `cd pkg && go test` | LOW | Must capture the workspace quirk, frontend `npx nx` commands, per-module build/vet/test |
| **Repo structure & workspace layout** (`structure.md`) | Multi-module `go.work`; which modules are/aren't in workspace; where services/pkg/agents live | MEDIUM | Already referenced by root `CLAUDE.md`. Keep capability-level, avoid enumerating every file |
| **Testing conventions** (`testing.md`) | Hard constraint: Ginkgo + Gomega; test comments in English | LOW | Already a named doc in `CLAUDE.md`. Include how to structure specs, naming, table-stakes coverage expectations |
| **Code style & language conventions** (`style.md` / `conventions.md`) | Constraint: domain comments/terminology in code language per CLAUDE.md; typed IDs; error conventions (sentinel vs wrapped) | MEDIUM | Encodes the agreed comment-language rule, naming, error handling, DTO-mapping-in-handler rule |
| **Architecture rules** (`architecture.md`) | DDD + hexagonal (NO CQRS bus) is the declared standard; agents must respect layer/port boundaries | MEDIUM-HIGH | The flow `inbound adapter (gRPC/cron) → use case (Execute) → domain → outbound port`; usecases-interactor + query-lite + UnitOfWork + transactional-outbox rules; aggregate factory + `PullEvents()`; what each layer may/may not import |
| **Git / workflow conventions** (`git.md`) | Constraint: remote, `main` branch, commit/PR norms; agents must not push/branch wrong | LOW | Branch model, commit message style, when to commit (parallel-agent safety), PR expectations |
| **Glossary / ubiquitous language** (`glossary.md`) | DDD demands a shared domain vocabulary (host, VM, owner, SRE, ITDC, namespace, project, shadow host) | MEDIUM | Most stable, most reusable doc; anchors agent reasoning. Maps EN/RU terms if dual-language |
| **Boundaries / "do not" rules** (in root file + `conventions.md`) | Agents need explicit guardrails: don't mass-fix `nil`-deps/`TODO` scaffolding; don't trust stale README/Makefile/compose | LOW | Project already has known scaffolding traps; codifying prevents agents "helpfully" breaking WIP |
| **Libraries reference** (`libraries.md`) | Shared `pkg/` packages need documented purpose/usage so agents reuse not reinvent | LOW-MEDIUM | Already referenced by `CLAUDE.md`; one row per shared package |

### Differentiators (high-value, not strictly required to "complete" the foundation)

| Feature (rule doc) | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Architecture Decision Records** (`decisions/ADR-*.md`) | Captures *why* (e.g. inventory out of `go.work`, restart from scratch, dropping CQRS in favor of usecases-interactor, UnitOfWork+outbox for events) — prevents agents re-litigating settled calls | MEDIUM | PROJECT.md "Key Decisions" table is the seed; lightweight ADR format. High leverage for a long multi-epic project |
| **Reference-implementation walkthrough** (`reference-service.md`) | `inventory` is meant to be the canonical DDD example; a guided tour teaches the pattern by example | MEDIUM | Caveat: service currently does not build; doc must mark it WIP/aspirational, not "copy this verbatim" |
| **Pattern catalog / recipes** ("how to add a use case / query / aggregate / repository") | Turns architecture rules into copy-able procedures; massively speeds correct agent output | MEDIUM-HIGH | Step-by-step: new use case = input/output DTO + use-case struct + `Execute` (wrap write in `uow.Do`) + wiring in `app/` + tests. Pairs with architecture doc |
| **Anti-pattern catalog** | Names the traps (anemic domain, leaking transport DTOs into domain, fat handlers, skipping `PullEvents` after save) | MEDIUM | Complements style/architecture; cheap insurance against systemic drift |
| **Onboarding guide** (human-oriented quickstart) | First-day path for new humans; reduces tribal knowledge | LOW-MEDIUM | Largely a curated reading order over the table-stakes docs |
| **`AGENTS.md` alongside `CLAUDE.md`** | Cross-tool portability (Codex, Cursor, Copilot, Aider, Gemini, Windsurf read `AGENTS.md`) | LOW | Make one the source of truth, the other a thin pointer, to avoid divergence |
| **Self-update / maintenance protocol** | A rule for *how the rules change* (update in PRs, audit cadence, keep root file synced) — stale instructions are worse than none | LOW | Keeps the base alive across many epics |
| **Conventions enforcement linkage** | Note where a rule is machine-enforced (linter, `go vet`, CI, Claude Code hooks) vs convention-only | MEDIUM | Raises trust: enforced rules > documented rules |

### Anti-Features (deliberately do NOT put in the knowledge base — over-documentation traps)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Exhaustive file-path map / per-file inventory** | "Help the agent navigate" | Paths drift constantly in AI-assisted code; stale paths send agents confidently to wrong place | Describe capabilities + stable domain concepts; let agents discover paths just-in-time |
| **Duplicating commands/conventions into every service's local doc** | Consistency everywhere | Multiple sources diverge; updates miss copies | Single canonical doc + nested docs only for genuine local overrides |
| **Documenting the future platform features as if they exist** | "Capture the vision" | Foundation milestone explicitly excludes business features; documenting non-existent SSH/inventory/auto-repair behavior creates phantom rules | Keep platform vision in PROJECT.md/roadmap; knowledge base documents only what's real now |
| **Mega single file with everything** | "One place to look" | Blows the context budget; dilutes the rules that matter; hard to maintain | Lean root file + progressive-disclosure linked docs (split when >~150–200 lines) |
| **Generated/auto-summarized API docs in the base** | Completeness | Goes stale instantly; not rule-shaped; noise for agents | Generate from code on demand; keep base to rules and decisions |
| **Prose essays / philosophy without actionable rules** | Sounds thorough | Agents can't act on vague "follow best practices"; burns context | Every entry must be a verifiable rule or command |
| **Personal/tool-specific preferences in shared base** | "It's how I work" | Pollutes team conventions; causes churn | Personal prefs in user-level (`~/.claude`) or git-ignored local files |
| **Codifying the `nil`/`TODO` scaffolding as if intentional design** | Documenting current state | Encourages agents to imitate or "complete" broken scaffolding | A single "this is WIP scaffolding, do not extend/fix" boundary rule |

### Foundation feature dependencies

```
knowledge/README.md (index, reading order)
    └──anchors──> all other docs

glossary.md (ubiquitous language)
    └──prerequisite for──> architecture.md ──> pattern catalog ──> anti-pattern catalog
                                  └──illustrated by──> reference-service.md

structure.md ──pairs with──> build.md (commands rely on knowing workspace layout)

architecture.md + style.md ──feed──> conventions/"do not" boundaries
decisions/ADR-*.md ──explains "why" behind──> architecture.md & structure.md & git.md
CLAUDE.md (root) ──imports/links──> every knowledge/*.md   (must stay lean)
AGENTS.md ──mirrors/points-to──> CLAUDE.md
```

**Dependency notes:**
- **README/index first:** every other doc plugs into it; defines reading order. Cheapest, highest-leverage starting point.
- **Glossary before architecture/patterns:** DDD rules are unusable without shared terms; the glossary is the most stable, reusable artifact.
- **Structure before commands:** build/test commands (`GOWORK=off`, `cd pkg`) only make sense once the workspace layout is documented.
- **ADRs explain, don't gate:** they reference the rule docs but no rule doc requires them — that's why ADRs are a differentiator, not table stakes.
- **Reference-service walkthrough depends on `inventory` building** — currently it does not; either mark aspirational or defer until the service compiles.

### MVP Definition (FOUNDATION milestone)

**Launch With (v1 — the milestone is "done" when these exist and the root `CLAUDE.md` links them):**
- [ ] `knowledge/README.md` index + reading order — anchors everything
- [ ] Build/run/test commands (incl. `GOWORK=off`, `cd pkg` test, `npx nx` frontend) — highest-frequency agent content
- [ ] Repo structure & `go.work` layout — without it, agents misnavigate the multi-module repo
- [ ] Testing conventions (Ginkgo + Gomega, English test comments) — hard project constraint
- [ ] Code style & language/error conventions (typed IDs, sentinel vs wrapped, DTO-mapping-in-handler) — hard constraint
- [ ] Architecture rules (DDD/hexagonal layer & port boundaries — NO CQRS bus; usecases-interactor, query-lite, UnitOfWork, transactional-outbox, aggregate/event rules) — the declared standard
- [ ] Git/workflow conventions — prevents wrong branch/push behavior
- [ ] Glossary / ubiquitous language — DDD prerequisite
- [ ] Boundaries / "do not" rules (don't fix scaffolding; don't trust stale README/Makefile/compose) — known traps already exist
- [ ] Root `CLAUDE.md` trimmed to a lean index that links the above

**Add After Validation (v1.x — once the base is in daily use and gaps surface):**
- [ ] ADRs seeded from PROJECT.md "Key Decisions" — trigger: a settled decision gets re-litigated by an agent/human
- [ ] Pattern catalog / "how to add a command/query/aggregate" — trigger: agents repeatedly get the layering wrong
- [ ] Anti-pattern catalog — trigger: recurring systemic mistakes observed
- [ ] `AGENTS.md` cross-tool mirror — trigger: team adopts a non-Claude agent
- [ ] Maintenance/self-update protocol — trigger: first stale-doc incident

**Future Consideration (later epics):**
- [ ] Reference-service walkthrough — defer until `inventory` builds and the application layer exists
- [ ] Per-service nested `CLAUDE.md`/`AGENTS.md` — defer until services diverge enough to need local overrides
- [ ] Enforcement linkage (CI/lint/hooks mapping) — defer until CI/lint pipeline exists

### Foundation prioritization matrix

| Feature (rule doc) | Agent/Team Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| README/index | HIGH | LOW | P1 |
| Build/run/test commands | HIGH | LOW | P1 |
| Repo structure / workspace | HIGH | MEDIUM | P1 |
| Testing conventions | HIGH | LOW | P1 |
| Code style & error/lang conventions | HIGH | MEDIUM | P1 |
| Architecture rules (DDD/hex, no CQRS bus) | HIGH | MEDIUM-HIGH | P1 |
| Git/workflow | MEDIUM | LOW | P1 |
| Glossary / ubiquitous language | HIGH | MEDIUM | P1 |
| "Do not" boundaries | HIGH | LOW | P1 |
| Libraries (`pkg/`) reference | MEDIUM | LOW-MEDIUM | P2 |
| ADRs | MEDIUM-HIGH | MEDIUM | P2 |
| Pattern catalog / recipes | HIGH | MEDIUM-HIGH | P2 |
| Anti-pattern catalog | MEDIUM | MEDIUM | P2 |
| `AGENTS.md` cross-tool mirror | MEDIUM | LOW | P2 |
| Onboarding guide | MEDIUM | LOW-MEDIUM | P2 |
| Maintenance protocol | MEDIUM | LOW | P2 |
| Reference-service walkthrough | MEDIUM | MEDIUM | P3 |
| Enforcement linkage | MEDIUM | MEDIUM | P3 |

---

## PART B — PLATFORM VISION: DC Host-Management / HaaS (BACKGROUND CONTEXT)

> Background only. These are FUTURE epics, explicitly out of scope for milestone 1. Listed so the knowledge base (Part A) covers the right domain language and architecture, and so the roadmap can sequence epics. Categories drawn from DCIM, bare-metal provisioning (e.g. MaaS, Tinkerbell/Foreman), and access-management ecosystems.

### Table Stakes (expected of any HaaS/host-management platform)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Host & VM inventory / single source of truth | Core promise; everything else references it | HIGH | Aggregate model already sketched in `inventory` (host, project, namespace, datacenter, rack) |
| Host state/health monitoring (health, checks, CPU) | Operators must see status before acting | HIGH | Drives auto-repair and safe actions |
| SSH access-rights granting & management | A core stated value (controlled, non-conflicting access) | HIGH | Must integrate with ownership/consistency model |
| Host actions (reboot, reimage/reprovision) | Baseline lifecycle control | HIGH | Needs external bot/hardware integration; idempotency + audit |
| Ownership / authorization model (owner vs SRE vs ITDC) | Core value: "nobody can take a host out of process" | HIGH | The consistency guarantee is the differentiating safety property |
| Audit log of actions | Compliance + traceability; `audit` service already exists | MEDIUM | Cross-cutting; needed early for any destructive action |

### Differentiators (where gwall-e competes)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Consistency/ownership safety across roles | The headline value — coordinated, conflict-free actions between owners and SRE/ITDC | HIGH | Domain-logic differentiator, not a commodity |
| Bulk / mass operations over host groups | Operate on fleets safely with the same consistency guarantees | HIGH | Hard part is doing it *safely* (partial failure, rollback, concurrency) |
| Auto-repair / self-healing | Reduce toil; act on health signals automatically | HIGH | Depends on monitoring + actions + strong guardrails |
| Unified owner/SRE/ITDC experience | Single tool replacing fragmented scripts/spreadsheets | MEDIUM-HIGH | UX + role-aware views |

### Anti-Features (platform, for future roadmap discipline)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Real-time everything (live streams of all host metrics) | "Dashboards feel alive" | Massive infra cost; rarely needed for control-plane decisions | Poll/event on meaningful state changes; reserve realtime for active operations |
| Generic "run arbitrary command on any host" | Maximum flexibility | Bypasses the ownership/consistency model — the product's whole point | Curated, audited, role-checked actions only |
| Building own metrics/monitoring stack | End-to-end ownership | Reinvents Prometheus/observability; huge scope | Integrate existing monitoring; consume signals |
| Bypass/override of ownership "for emergencies" without audit | Operator convenience | Erodes the core consistency guarantee | Break-glass flow that is explicit, time-boxed, and fully audited |

### Platform dependency sketch

```
Inventory (source of truth)
    └──required by──> Monitoring ──> Auto-repair
    └──required by──> Ownership/authorization model ──> SSH access grants
                                                    └──> Host actions ──> Bulk operations
Audit ──cross-cuts──> all destructive actions
```

---

## Sources

- GitHub Blog — *How to write a great agents.md: lessons from over 2,500 repositories* (https://github.blog/ai-and-ml/github-copilot/how-to-write-a-great-agents-md-lessons-from-over-2500-repositories/) — HIGH
- *AGENTS.md: The Open Standard for Guiding AI Coding Agents (2026)* — devtk.ai (https://devtk.ai/en/blog/what-is-agents-md-guide/) — MEDIUM
- *AGENTS.md Complete Guide for Engineering Teams (2026)* — buildbetter.ai (https://blog.buildbetter.ai/agents-md-complete-guide-for-engineering-teams-in-2026/) — MEDIUM
- Claude Code Docs — *How Claude remembers your project (memory)* (https://code.claude.com/docs/en/memory) — HIGH
- Cline Docs — *Memory Bank best practices* (https://docs.cline.bot/best-practices/memory-bank) — HIGH
- *The Complete Guide to CLAUDE.md: Memory, Rules, Loading, and Cross-Tool Compression* — Medium/Bijit Ghosh (https://medium.com/@bijit211987/the-complete-guide-to-claude-md-memory-rules-loading-and-cross-tool-compression-97cc12ed037b) — MEDIUM
- Project context: `.planning/PROJECT.md`, root `CLAUDE.md` (gwall-e) — HIGH
- Platform-vision categorization: domain knowledge of DCIM / bare-metal provisioning (MaaS, Tinkerbell, Foreman) and access-management patterns — MEDIUM

---
*Feature research for: AI-conventions knowledge base (foundation) + DC HaaS platform (vision)*
*Researched: 2026-06-17*
