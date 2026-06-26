# Pitfalls Research

**Domain:** AI-agent / team conventions knowledge base (FOUNDATION milestone, PRIMARY) + Go DDD/hexagonal microservices (NO CQRS bus) for DC Hardware-as-a-Service (background for future epics)
**Researched:** 2026-06-17
**Confidence:** HIGH (knowledge-base layer — corroborated by GitHub/Anthropic/Cline guidance and direct repo inspection); MEDIUM (platform layer — domain knowledge of DDD and DCIM/HaaS, not yet built)

> **Scope.** The first milestone ships ONLY the `knowledge/` (memory-bank-style) convention base. Sections 1–8 below ("Critical Pitfalls") are the PRIMARY, detailed focus and almost all map to milestone-1 phases. The platform/DDD/HaaS pitfalls are background for future epics; they are included so milestone-1 docs encode the *right* rules and so the roadmap can sequence later epics, but they are NOT milestone-1 deliverables.

---

## Critical Pitfalls

### Pitfall 1: The comment-language rule is not encoded canonically (decision MADE — Russian)

**Status: RESOLVED.** The team decided: Go code comments + domain terminology are **RUSSIAN** (identifier names stay English; test comments stay English). PROJECT.md has been reconciled (requirement block + Constraints + Key Decisions all say Russian) and matches root `CLAUDE.md`. The remaining risk is failing to encode this single rule canonically in the knowledge base.

**What goes wrong (if not encoded canonically):**
Even with the decision made, if the rule is restated in multiple docs (or left implicit), drift returns. Agents oscillate (some files RU, some EN), reviewers waste cycles, and the "ubiquitous language" that DDD depends on fractures.

**Why it happened originally:**
Conventions were captured in two places (CLAUDE.md and PROJECT.md) at different times and nobody reconciled them. Contradictory instructions across sources are the #1 cause of inconsistent agent behavior. (Now reconciled.)

**How to avoid:**
1. **Decision is made — Russian.** Do not re-litigate; record the rationale.
2. **Encode in exactly ONE canonical doc** (`knowledge/style.md` or `conventions.md`) as a hard **MUST**: "комментарии и доменная терминология — на русском; имена идентификаторов — английские; комментарии в тестах — английские."
3. **Every other file points to that canonical rule**, never restates it. Root `CLAUDE.md` keeps a pointer.
4. **Capture as an ADR** so it is not re-litigated by a future agent or contributor.
5. **Note enforcement reality**: comment language is hard to lint mechanically; flag it as convention-only + review-enforced.

**Warning signs:**
- The rule is restated (not linked) in more than one doc.
- Newly generated files mix RU and EN comments.
- Reviewers leave "wrong language" comments on PRs.

**Phase to address:**
Milestone-1 — the phase that produces `style.md`/`conventions.md` and the glossary. This is a **prerequisite/blocker**: resolve before writing the glossary (ubiquitous language) and before any reference-pattern docs. Also fix PROJECT.md in the same phase.

---

### Pitfall 2: Mega-file that blows the context budget (monolithic CLAUDE.md / memory-bank)

**What goes wrong:**
All rules are stuffed into one large root `CLAUDE.md` (or one giant `knowledge.md`). It is injected into the system prompt on **every** API call, consuming tokens constantly, and once it grows past the model's reliable instruction budget (~100–150 usable instruction "slots" after Claude's own system prompt takes ~50), the model silently drops rules — including the ones that matter. Over-specified instruction files also measurably *hurt*: unnecessary requirements increase reasoning tokens 14–22% and reduce completeness on complex tasks.

**Why it happens:**
"One place to look" feels tidy and authoritative. The cost (per-call token tax, attention dilution) is invisible until agents start ignoring rules.

**How to avoid:**
- **Lean root entry file (~80–150 lines), progressive disclosure.** Root `CLAUDE.md`/`AGENTS.md` carries only always-true, high-signal rules + pointers; topic detail lives in `knowledge/*.md` loaded on demand (e.g. "When touching the persistence layer, read `knowledge/architecture.md`").
- **One topic per file**, split any file that exceeds ~150–200 lines.
- **The pruning test for every line:** "Would removing this cause an agent to make a mistake?" If no, cut it.
- Use Claude Code's subdirectory-file mechanism and/or `@`-imports sparingly — reference, don't inline.

**Warning signs:**
- Root agent file > ~200 lines, or any topic file ballooning past ~200.
- Agents repeatedly ignore a rule that is "clearly written" near the bottom of a long file.
- The file contains content the model already knows ("be a senior engineer", generic Go style).

**Phase to address:**
Milestone-1 — the phase that designs the `knowledge/` layout and trims the existing root `CLAUDE.md` into a lean index. This is the structural decision the whole base hangs on; address early (right after README/index).

---

### Pitfall 3: Brittle file-path maps / per-file inventories that drift

**What goes wrong:**
The base enumerates concrete file paths ("the host aggregate is in `services/inventory/internal/domain/host.go`") and directory listings. In an AI-assisted, actively-refactored codebase (this one is mid-teardown of `pkg/mediatr` and `inventory`), paths move constantly. Stale paths send agents *confidently* to the wrong place — worse than no map, because the agent trusts it.

**Why it happens:**
"Help the agent navigate" instinct. Path maps look helpful and concrete. But GitHub's 2,500-repo study found directory listings don't help agents navigate faster — agents discover structure themselves.

**How to avoid:**
- **Describe stable capabilities and domain concepts, not paths.** "Aggregates live in the domain layer; one file per aggregate" is stable; the exact path is not.
- Document the *layout pattern* (the hexagon's layer→directory mapping) once, in `structure.md`, at capability level — not a per-file inventory.
- Let agents find paths just-in-time via search.
- If a path *must* be named (e.g. `go.work`, `lefthook.yml`), pick load-bearing, rarely-moving ones only.

**Warning signs:**
- A doc lists more than a handful of specific source-file paths.
- "File moved, doc not updated" appears in review.
- Agents cite paths that no longer exist.

**Phase to address:**
Milestone-1 — the `structure.md` / repo-layout phase. Verify by grepping the finished docs for source-file paths and challenging each one.

---

### Pitfall 4: Codifying the `nil`/`TODO` scaffolding as if it were intentional design

**What goes wrong:**
The repo deliberately contains broken scaffolding: `services/inventory` does not build (missing `application` root layer), `main.go` has intentional `nil`-dependencies and `// TODO`s, and the root `README.md`/`Makefile`/`docker-compose.yml` are explicitly stale. If the knowledge base documents the current state as "how it works," it (a) teaches agents to imitate broken patterns, and (b) invites agents to "helpfully" finish or fix the scaffolding — destroying work-in-progress that is meant to be torn down and rebuilt.

**Why it happens:**
A naive "document current reality" pass treats every existing artifact as ground truth. Auto-`/init` runs are especially prone to this — they describe whatever is on disk.

**How to avoid:**
- **One explicit boundary/"do not" rule:** "`inventory` is WIP reference scaffolding; it does not build; do NOT extend, complete, or fix its `nil`-deps / `TODO`s. Stale `README.md`/`Makefile`/`docker-compose.yml` are NOT authoritative — ignore them."
- **Never auto-generate the base** (`/init` or letting an agent write its own AGENTS.md) — generated files tend to enshrine current (broken) state and duplicate existing docs.
- If a reference-implementation walkthrough is written, mark it **aspirational/WIP**, not "copy this verbatim."

**Warning signs:**
- A doc describes `inventory` as a working example without a WIP caveat.
- Agents open PRs "fixing" the build or filling in `nil` deps unprompted.
- The base repeats anything from the stale README/Makefile.

**Phase to address:**
Milestone-1 — the "boundaries / do-not rules" phase (and any reference-service doc). High-leverage, low-cost; the traps already exist today.

---

### Pitfall 5: Ambiguous requirement strength (MUST vs SHOULD vs MAY) and "don'ts" without "dos"

**What goes wrong:**
Rules are written as prose ("we generally prefer typed IDs", "try to keep handlers thin") with no explicit strength. Agents and humans cannot tell a hard invariant from a soft preference, so hard rules get violated and soft preferences get over-applied. Separately, a base full of bare prohibitions ("don't do X") with no prescribed alternative makes agents over-explore and do less work — 15+ sequential "don'ts" with no "dos" measurably degraded agent behavior.

**Why it happens:**
Natural prose hides modality. Writers list dangers without the safe path.

**How to avoid:**
- **Phrase every rule as MUST / SHOULD / WON'T (MAY).** Reserve MUST for true invariants (e.g. "events MUST be pulled via `PullEvents()` only after a successful Save").
- **Pair every prohibition with the prescribed alternative:** not "don't put transport DTOs in the domain" but "map DTO→domain inside the handler; the domain MUST NOT import transport types."
- Lead docs with copy-pasteable commands and the *do* path; keep prohibitions few and always paired.

**Warning signs:**
- Rule docs use hedging verbs ("prefer", "try", "generally") without MUST/SHOULD tags.
- A doc has a long "don't" list and few "do" examples.
- Agents ask "is this a hard rule?" or violate a rule that was meant as MUST.

**Phase to address:**
Milestone-1 — applies as a cross-cutting authoring standard to every rule doc (testing, style, architecture, git). Set it in the README/authoring-conventions phase so all subsequent docs follow it.

---

### Pitfall 6: Stale / unmaintained rules with no update protocol

**What goes wrong:**
The base is written once and rots. Tech-stack facts drift (the repo *already* pins deprecated `mongo-driver` v1, `chi` despite a gRPC mandate, mixed Ginkgo v1/v2 — see STACK.md). Volatile state (project status, current WIP) is baked into a file that loads every session and is stale by next week. A stale base is **worse than none** — agents follow confidently wrong instructions.

**Why it happens:**
No owner, no cadence, no rule for "how the rules change." Rules and fast-moving project status get mixed into the same file.

**How to avoid:**
- **Write a maintenance/self-update protocol** (a differentiator doc): rules change via PR, reviewed like code; audit cadence; root file kept in sync with `knowledge/`.
- **Separate stable rules from volatile status.** Volatile/WIP status belongs in the moment (PR description, issue, roadmap), not in the always-loaded base.
- **"Update the base when an agent repeatedly makes the same mistake"** — grow through iteration, not big-bang.
- Map each rule to its enforcement (linter / `go vet` / CI / hook) so drift between doc and tooling is visible.

**Warning signs:**
- A rule references a version/library no longer used (already true today).
- "When did we last review this?" has no answer.
- The base contains "current status" that changes weekly.

**Phase to address:**
Milestone-1 for the protocol skeleton; ongoing thereafter. Pairs with the enforcement-linkage doc (deferrable until CI exists).

---

### Pitfall 7: Rules that aren't machine-enforced presented as if they were guaranteed

**What goes wrong:**
The base treats "documented" as "enforced." Code-style and import-ordering rules that *could* be a linter/formatter are left as prose; conversely, soft conventions are presented with the same authority as CI-gated invariants. Net effect: format/style drift despite the doc, and false confidence in unenforced rules. Empirical evidence that agent files improve outcomes is mixed — the *reliable* value is consistency, and consistency comes from enforcement.

**Why it happens:**
Writing a rule is cheaper than wiring a check; the gap between "stated" and "enforced" is invisible until violations accumulate.

**How to avoid:**
- **Move every mechanizable rule into tooling** and let the doc point to it: `gofumpt`/`gci` for formatting+import order, `golangci-lint` v2 for lint, `buf lint`/`buf breaking` for proto, commitlint for commit messages, `go test -race` in CI, Lefthook for fast pre-commit/pre-push (see STACK.md). "Use linters/formatters instead of code-style prose."
- **Tag each rule with its enforcement status** (CI-gated / hook / convention-only). Convention-only = trust accordingly.
- Keep hooks fast (lint/format/unit); heavy suites in CI — or devs use `--no-verify`.

**Warning signs:**
- A style rule exists in docs but no formatter enforces it.
- Formatting/import diffs appear in review despite a documented rule.
- Commits land that violate a "MUST" with no CI failure.

**Phase to address:**
Milestone-1 — the enforcement-layer phase (golangci-lint v2 config, gofumpt, Lefthook, commitlint, buf). STACK.md already specifies the toolchain; the knowledge base must link rule→tool.

---

### Pitfall 8: Documenting future platform features as if they exist (phantom rules)

**What goes wrong:**
Milestone 1 explicitly excludes business features (SSH grants, inventory, host actions, bulk ops, auto-repair, monitoring). If the base documents conventions/behavior for these non-existent subsystems, it creates "phantom rules" agents try to satisfy against code that isn't there — wasted budget, invented patterns, and contradictions when the real feature later lands differently.

**Why it happens:**
"Capture the vision" instinct; the line between roadmap and rules blurs.

**How to avoid:**
- **The base documents only what is real now.** Platform vision stays in PROJECT.md / FEATURES.md / roadmap.
- The base *may* establish stable **ubiquitous-language** terms (host, VM, owner, SRE, ITDC, namespace, project) since the glossary is stable — but it must NOT prescribe behavior for unbuilt features.
- Add feature-convention docs *when the feature is being built*, not before.

**Warning signs:**
- A `knowledge/*.md` describes "how SSH grants work" while no such code exists.
- Agents reference subsystems that aren't implemented.

**Phase to address:**
Milestone-1 scoping — the README/index and architecture phases. Verify by checking each doc names only existing capabilities (plus the stable glossary).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| One big `CLAUDE.md` instead of split `knowledge/*.md` | Fast to write; "one place" | Per-call token tax, dropped rules, unmaintainable | Never for the foundation; only a throwaway scratch note |
| Copy commands/conventions into every service doc | Local convenience | Divergent copies; updates miss some | Only for genuine *local overrides*, clearly marked |
| Leave comment-language rule ambiguous "for now" | Avoids a hard decision | RU/EN drift across whole codebase; broken ubiquitous language | Never — it is a milestone-1 blocker |
| Auto-`/init` to bootstrap the base | Instant draft | Enshrines broken scaffolding; duplicates existing docs | Only as a discard-after-review seed, never committed as-is |
| State rules as prose without MUST/SHOULD tags | Reads naturally | Ambiguous enforcement; inconsistent agent behavior | Never for normative rules |
| Document rule without wiring the linter/hook | Ship the doc faster | Style/format drift; false confidence | Acceptable temporarily IF tagged "convention-only, not yet enforced" |
| Keep `inventory` chi/Mongo-v1/Ginkgo-v1 as the "example" | No migration work | Teaches deprecated stack as canonical | Never document it as canonical; mark WIP + name target stack |

## Integration Gotchas

*(Background — future epics; the knowledge base should name these as target conventions, not implement them in milestone 1.)*

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| External bot / hardware service (outbound) | Leak the HTTP client / response shapes into the domain | Hide behind an outbound port; map to domain in the adapter; wrap errors with `%w` |
| MongoDB (outbound) | Start new services on deprecated driver **v1** (repo currently pins it) | Standardize `go.mongodb.org/mongo-driver/v2`; encode as a MUST so no service starts on v1 |
| gRPC transport | Put request validation in domain code; mismatch APIv1/APIv2 | Declarative `protovalidate` at the edge; standardize protobuf APIv2 (`google.golang.org/protobuf`) |
| Domain-event propagation across services | Reach for a broker before it's needed; couple domain to transport | Keep in-process `PullEvents()` intra-service; add a broker adapter only at the integration boundary |
| Cross-tool agent files (Claude + Codex/Cursor/etc.) | Maintain `CLAUDE.md` and `AGENTS.md` as two separate full copies → divergence | One canonical source; the other is a thin pointer/symlink |

## Performance Traps

*(The milestone-1 deliverable is documentation; "performance" here is the agent context budget. Platform rows are background for future epics.)*

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Context-budget exhaustion from a bloated base | Agents ignore rules; reasoning-token inflation (14–22%) | Lean root + progressive disclosure; prune test on every line | Past ~100–150 instruction slots / ~200-line root file |
| Unbounded bulk operation over host groups (future) | One mass action hits hundreds of hosts; partial failures, no rollback | Blast-radius limits, batching, dry-run, per-host idempotency, audit | When fleets grow beyond a handful of hosts |
| Auto-repair feedback loop (future) | Repair action triggers health signal that triggers more repairs (flapping) | Rate-limit/debounce, circuit breaker (`gobreaker`), max-attempts, require stable signal | Under correlated/flapping failures |
| Read-model fan-out / outbox relay lag (future) | Slow projections, stale reads at scale | Bounded projections, idempotent handlers, monitor relay lag | As query volume / event rate grows |

## Security Mistakes

*(Background — future epics. The headline platform value is consistency/ownership safety; these are the domain-specific hazards to encode when those epics start.)*

| Mistake | Risk | Prevention |
|---------|------|------------|
| Ownership "take a host arbitrarily" race | Two roles (owner vs SRE/ITDC) mutate the same host concurrently; the core consistency guarantee breaks | Enforce ownership/authorization as a domain invariant on the host aggregate; optimistic concurrency / versioning; serialize conflicting actions |
| Generic "run any command on any host" | Bypasses the entire ownership model (the product's whole point) | Curated, role-checked, audited actions only — no arbitrary exec |
| Unaudited emergency override / break-glass | Erodes the consistency guarantee; no traceability | Explicit, time-boxed, fully audited break-glass flow |
| Destructive action with no audit trail | Reboot/reimage with no record; compliance gap | Route all destructive actions through the existing `audit` service before/with execution |
| Over-broad SSH grant lifetime | Standing access accumulates | Scoped, expiring grants tied to the ownership model |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| New contributor can't find the rules | Onboarding drag; tribal knowledge persists | `knowledge/README.md` index defines a reading order; root `CLAUDE.md` links it |
| Rules read as vague philosophy | Humans (and agents) can't act on "follow best practices" | Every entry is a verifiable rule or copy-pasteable command |
| Mass-operation UI with no dry-run/scope preview (future) | Operator fat-fingers a fleet-wide destructive action | Show blast radius + require confirmation/dry-run before bulk execution |
| Role-blind views (future) | Owner/SRE/ITDC see actions they can't safely take | Role-aware views consistent with the ownership model |

## "Looks Done But Isn't" Checklist

- [ ] **Comment-language rule:** appears decided — verify a SINGLE language is stated in ONE canonical doc, that PROJECT.md (req block + Constraints) and root CLAUDE.md all agree, and an ADR records the choice.
- [ ] **Root `CLAUDE.md`:** looks lean — verify it actually links `knowledge/*.md` and doesn't restate detail (and is < ~150 lines).
- [ ] **`knowledge/` set:** looks complete — verify the table-stakes set exists (README/index, build/test incl. `GOWORK=off`+`cd pkg`, structure/`go.work`, testing Ginkgo+Gomega, style/error/lang, architecture DDD/hex — no CQRS bus; usecases-interactor, query-lite, UnitOfWork, transactional-outbox — git, glossary, do-not boundaries).
- [ ] **Boundary rules:** present — verify the "do NOT fix `inventory` scaffolding / `nil`-deps / `TODO`s; stale README/Makefile/compose are not authoritative" rule exists.
- [ ] **Requirement strength:** verify normative rules carry MUST/SHOULD/WON'T and prohibitions are paired with a prescribed "do."
- [ ] **Enforcement linkage:** verify each mechanizable rule names its tool (golangci-lint v2 / gofumpt / buf / commitlint / `go test -race`) or is tagged "convention-only."
- [ ] **No phantom features:** verify no `knowledge/*.md` prescribes behavior for unbuilt SSH/inventory/auto-repair/bulk/monitoring subsystems.
- [ ] **No brittle path map:** grep finished docs for source-file paths; each surviving path must be load-bearing and stable.
- [ ] **Stack rules current:** verify docs name target stack (mongo-driver **v2**, gRPC, Ginkgo **v2**), not the deprecated versions still pinned in the repo.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Comment-language contradiction shipped | MEDIUM | Decide; pick one doc as canonical; fix CLAUDE.md + both PROJECT.md spots; write ADR; sweep/normalize existing comments (and add a lint/review check) |
| Mega-file already in use | MEDIUM | Split into `knowledge/*.md` by topic; trim root to a pointer index; run the per-line pruning test |
| Brittle path map drifted | LOW | Delete path enumerations; replace with capability/layout descriptions |
| Scaffolding codified as design | LOW | Add the WIP/do-not boundary rule; mark reference-service doc aspirational |
| Stale rules accumulated | LOW–MEDIUM | Introduce the maintenance protocol; audit pass; move volatile status out of the always-loaded base |
| Ownership race shipped (future) | HIGH | Add aggregate-level invariant + optimistic concurrency; backfill audit; likely data reconciliation |
| Auto-repair flapping (future) | MEDIUM–HIGH | Add rate-limit/circuit-breaker/max-attempts; require stable signal; post-mortem the loop |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. EN-vs-RU comment contradiction | Milestone-1: style/conventions + glossary (BLOCKER, do first) | One language, one canonical doc; CLAUDE.md + PROJECT.md×2 agree; ADR exists |
| 2. Mega-file / context budget | Milestone-1: knowledge layout + trim root CLAUDE.md | Root < ~150 lines, topic files split, pointers not inlines |
| 3. Brittle path map | Milestone-1: `structure.md` | No non-load-bearing source paths in docs |
| 4. Scaffolding codified | Milestone-1: do-not boundaries | Explicit WIP/do-not rule present; no `/init`-generated content committed |
| 5. Ambiguous MUST/SHOULD + bare don'ts | Milestone-1: authoring standard (set in README phase, applied everywhere) | Rules tagged; prohibitions paired with dos |
| 6. Stale rules / no update protocol | Milestone-1: maintenance protocol (skeleton); ongoing | Protocol doc exists; volatile status not in base |
| 7. Unenforced rules | Milestone-1: enforcement layer (linters/hooks/CI) | Each mechanizable rule → tool; format/lint gated in CI |
| 8. Phantom future features | Milestone-1: scoping (README + architecture) | Docs describe only existing capabilities + stable glossary |
| Ownership/consistency races | Future epic: ownership/authorization | Aggregate invariant + concurrency control + tests |
| Bulk-op blast radius | Future epic: bulk operations | Batch limits, dry-run, idempotency, rollback path |
| Auto-repair feedback loop | Future epic: auto-repair | Rate-limit/breaker/max-attempts; flapping test |
| Event-before-commit / anemic domain / infra-in-domain | Future epics: per-service domain work | `PullEvents()` only after Save; behavior on aggregates; domain imports no infra/transport |

## Sources

- GitHub Blog — *How to write a great agents.md: lessons from over 2,500 repositories* (https://github.blog/ai-and-ml/github-copilot/how-to-write-a-great-agents-md-lessons-from-over-2500-repositories/) — directory listings don't help; tools mentioned get used 160x more — **HIGH**
- Augment Code — *A good AGENTS.md is a model upgrade. A bad one is worse than no docs at all.* (https://www.augmentcode.com/blog/how-to-write-good-agents-dot-md-files) — don'ts-without-dos degrade behavior; bloat hurts — **MEDIUM-HIGH**
- Anthropic — *Claude Code best practices / memory* (https://code.claude.com/docs/en/best-practices) — over-specified CLAUDE.md, per-call token cost, progressive disclosure — **HIGH**
- Bijit Ghosh — *The Complete Guide to CLAUDE.md* (https://medium.com/@bijit211987/the-complete-guide-to-claude-md-memory-rules-loading-and-cross-tool-compression-97cc12ed037b) — ~150 instruction slots, contradictory-instruction failure mode, pruning test, auto-gen risk — **MEDIUM**
- Cline Docs — *Memory Bank best practices* (https://docs.cline.bot/best-practices/memory-bank) — versioned topic files, keep lean, maintain — **HIGH**
- ADI Pod — *CLAUDE.md Best Practices: 7 to put in, 3 to leave out* (https://adipod.ai/blog/claude-md-best-practices/) — leave out code style (use linters) and directory listings — **MEDIUM**
- Project context: `.planning/PROJECT.md` (Constraints L63 EN vs requirement L28 RU), root `CLAUDE.md` (RU comments rule), `.planning/research/FEATURES.md`, `.planning/research/STACK.md` — **HIGH** (direct inspection)
- Platform/DDD + DCIM/HaaS hazards (anemic domain, infra-in-domain, aggregate boundaries, event-before-commit, ownership races, bulk blast radius, auto-repair loops) — domain knowledge + milestone context — **MEDIUM**

---
*Pitfalls research for: AI-conventions knowledge base (foundation, primary) + Go DDD/hexagonal HaaS platform — no CQRS bus (future epics)*
*Researched: 2026-06-17*
