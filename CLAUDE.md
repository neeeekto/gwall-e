# gwall-e

Gwall-E — платформа управления серверами в дата-центрах. Позволяет: видеть состояние хостов, инвентаризировать их, выписывать на них права, выполнять операции над хостами (в т.ч. массовые действия) и автолечение.

Технологии: Go-микросервисы (бэкенд) + React/Nx (фронтенд). Архитектурный подход: DDD + clear архитектура.

<!-- GSD:project-start source:PROJECT.md -->

## Project

**gwall-e**

gwall-e — платформа **Hardware-as-a-Service** для дата-центров. Она даёт овнерам, SRE и ITDC единый инструмент, чтобы инвентаризировать хосты и VM, видеть их состояние, выписывать SSH-права, выполнять действия над хостами (в т.ч. массовые) и автоматически их чинить — при этом сохраняя согласованность: никто не может «забрать» чужой хост в обход правил.

Технически это набор Go-микросервисов (бэкенд, DDD/гексагональная архитектура) и React/Nx фронтенд.

**Core Value:** Безопасное и **согласованное** управление парком серверов как услугой: единый источник правды о хостах и контролируемый, неконфликтный доступ к действиям над ними между овнерами и SRE/ITDC.

### Constraints

- **Tech stack (backend):** Go 1.24.6, мульти-модульный workspace; DDD + гексагональная архитектура; MongoDB (outbound), GRPC.
- **Tech stack (frontend):** React + Nx.
- **Тесты:** Ginkgo + Gomega; комментарии в тестах на английском, доменные комментарии в коде — на русском.
- **Язык:** комментарии и доменная терминология в Go-коде — на русском (имена идентификаторов — английские); сохранять этот стиль.
- **Сборка:** `inventory` собирать/тестировать с `GOWORK=off` из каталога модуля; при добавлении сервиса в общий build — добавлять модуль в `go.work`.
- **Git:** remote `origin` → `github.com/neeeekto/gwall-e`; основная ветка — `main`.

<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->

## Technology Stack

## Recommended Stack

### Core Technologies (platform — what the conventions standardize)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go (toolchain) | 1.24.6 | Language/runtime | Already mandated. `go.work` multi-module workspace is the right call for a many-service monorepo: per-service `go.mod` keeps dependency graphs isolated, workspace gives one-command local builds. Keep the deliberate "`inventory` excluded, build with `GOWORK=off`" pattern documented as a rule. |
| google.golang.org/grpc | v1.81.x (1.81.1 latest) | Inbound/outbound RPC transport | Mandated by PROJECT.md. The canonical, highest-throughput Go gRPC implementation; broad interop. Use it directly for the inbound adapter in the hexagon. Support policy: only the two latest major Go releases — fine at 1.24. |
| google.golang.org/protobuf | v1.36.x | Protobuf runtime (APIv2) | Required by grpc-go and the buf toolchain. Standardize the APIv2 (`google.golang.org/protobuf`) import — never the deprecated `github.com/golang/protobuf`. |
| go.mongodb.org/mongo-driver/v2 | v2.7.x (v2 GA Jan 2025) | Persistence (outbound adapter) | MongoDB mandated. **Migrate off v1** — the repo currently pins v1 `mongo-driver v1.17.9`, which is now deprecated. v2 has new import path `.../v2/mongo`, merged GridFS, global `omitempty`, and `errors.Is/As` support on public APIs. Lock this as a convention so no new service starts on v1. |
| buf (CLI) | v1.71.x (Jun 2026) | Protobuf build/lint/breaking-change/codegen | The de-facto standard for managing `.proto` in 2025/26. `buf generate` (with `buf.gen.yaml`) replaces hand-rolled `protoc` invocations; `buf lint` + `buf breaking` enforce schema hygiene in CI. Critical for a multi-service contract-first platform. |
| log/slog (stdlib) | Go 1.24 stdlib | Structured logging | Use the standard library `log/slog` as the logging *interface* across all services. Avoids a forced dependency choice and is natively supported by grpc-middleware's slog provider. Pick a handler (JSON in prod) at the composition root only. |

### Supporting Libraries (platform)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/grpc-ecosystem/go-grpc-middleware/v2 | v2.3.3 | Interceptor chaining: logging, recovery, auth, validator | Standard cross-cutting concerns for every gRPC server. Mandate chain order in conventions: logging first, recovery **last** (so logging sees recovered state). Use `interceptors/logging` with the `providers/slog` adapter. |
| github.com/bufbuild/protovalidate-go | v1.x (protovalidate GA Sep 2025) | Declarative request validation from `.proto` rules | Successor to deprecated `protoc-gen-validate`. Define validation once in schema (`buf.validate`), enforce via the grpc-middleware protovalidate interceptor. Keeps validation out of domain code. |
| github.com/google/uuid | v1.6.0 | ID generation for typed IDs | Already in repo. Pairs with the project's typed-ID convention (`type HostID string`). Keep. |
| github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery | (within v2.3.3) | Panic → `codes.Internal` | Always installed; never let a panic cross the RPC boundary. |
| go.uber.org/mock (gomock) | v0.5.x | Generated mocks for ports/interfaces | For mocking outbound ports (repositories, bot client) in Ginkgo specs. Prefer over hand-written mocks for large interfaces. Optional — counterfeiter is an alternative if you prefer interface fakes. |
| github.com/onsi/ginkgo/v2 | v2.23.4 | BDD test framework | Mandated. Already pinned in `pkg/go.mod`. |
| github.com/onsi/gomega | v1.38.0 | Matcher/assertion library | Mandated. Already pinned. Bump root module's stale `gomega v1.37.0` to match. |
| github.com/sony/gobreaker | v1.0.0 | Circuit breaker for outbound calls | Already in `pkg/`. Useful for the bot-service HTTP client and other flaky external dependencies. |
| github.com/hashicorp/go-retryablehttp | v0.7.8 | Retrying HTTP client | Already in `pkg/`. For outbound HTTP adapters (external bot/DC integrations). Bump root module's stale v0.7.7. |

### Development Tools (the milestone-1 enforcement layer)

| Tool | Purpose | Notes |
|------|---------|-------|
| golangci-lint | v2.12.x (v2.12.2, May 2026) | **Primary convention enforcer.** Use v2 config schema (`linters.default: standard`, then opt-in). Commit a single root `.golangci.yml`. Note: v2 dropped `enable-all/disable-all` → `linters.default`; run `golangci-lint migrate` only if you have a v1 config (you don't — start clean on v2). Run with `GOWORK=off` semantics where needed (lint per module). |
| gofumpt | stricter `gofmt` superset | Standardize formatting beyond `gofmt`. Wire as a golangci-lint formatter (`formatters` block in v2) so format == lint == CI are identical. |
| goimports / gci | import grouping | Enforce deterministic import ordering (stdlib / external / `github.com/gwall-e/...`). Configure via golangci-lint v2 `formatters`. |
| buf lint + buf breaking | proto schema governance | Run in CI on every PR touching `.proto`. `buf breaking` against the `main` branch prevents accidental wire-incompatible changes across services. |
| Lefthook | v1.x | Git hooks manager (pre-commit, commit-msg, pre-push) | **Recommended over Husky** for this Go repo: single dependency-free Go binary, one `lefthook.yml`, parallel execution, no `node_modules`. `pre-commit` → `golangci-lint run` + `gofumpt`; `pre-push` → `go test`/`ginkgo`; `commit-msg` → commitlint. Keep hooks fast; CI is the source of truth. Requires `lefthook install` per clone (document this). |
| commitlint + config-conventional | `@commitlint/cli` + `@commitlint/config-conventional` (latest) | Enforce **Conventional Commits** in the `commit-msg` hook (`commitlint --edit {1}`). Only Node dependency in the repo; isolate it (devDependency, documented). Conventional Commits also unlocks future automated changelog/versioning. |
| go vet / go test -race | stdlib | Baseline correctness + data-race detection in CI. Document `cd services/inventory && GOWORK=off go vet ./...` form for out-of-workspace modules. |

## Installation

# --- Go module dependencies (per service module, run inside the module dir) ---

# Test deps

# --- CLI / dev tooling (host install) ---

# --- commitlint (the only Node tooling; pin exact) ---

# --- activate hooks (once per clone) ---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| grpc-go (direct) | ConnectRPC (connect-go) | Choose Connect if you later need browser/gRPC-Web clients without a proxy, HTTP/1.1, or curl-able endpoints. It's gRPC-wire-compatible and migration is incremental. For a pure backend-to-backend platform mandated to use gRPC, stay on grpc-go (higher throughput, canonical). |
| Manual DI (composition root in `main.go`) | google/wire | The project already uses hand-wired DI in `main.go` — keep it as the convention while services are few. Adopt **google/wire** (compile-time, codegen, zero runtime cost, fast startup) if/when manual wiring in `main.go` becomes error-prone. |
| Manual DI | uber/fx | Only if many services start sharing large infra modules (metrics/logging/health/config) and you want lifecycle hooks (graceful start/stop). Heavier, runtime-reflection-based, steeper learning curve — overkill now. |
| log/slog | zap / zerolog | Use zap/zerolog only if you measure slog's allocation/throughput as a real bottleneck. grpc-middleware ships providers for all three, so swapping is localized. Default to slog for zero-dependency portability. |
| protovalidate | hand-written validation in handlers | Hand validation is fine for trivial cases, but schema-driven protovalidate keeps rules in one place and out of domain code. |
| Lefthook | Husky / pre-commit (Python) | Husky only if the repo gains a real JS/Node app needing its ecosystem. `pre-commit` (Python) if the team already standardizes on it elsewhere. Otherwise Lefthook (no extra runtime). |
| gomock | counterfeiter | counterfeiter generates interface fakes (no call-expectation DSL) — preferred by some for hexagonal port doubles. Either is fine; pick one and standardize. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `go.mongodb.org/mongo-driver` v1 (`v1.17.x`) | Officially deprecated since v2 GA (Jan 2025); no new features. Repo currently pins it. | `go.mongodb.org/mongo-driver/v2` |
| `github.com/golang/protobuf` (APIv1) | Legacy protobuf API; superseded. | `google.golang.org/protobuf` (APIv2) |
| `github.com/grpc-ecosystem/go-grpc-middleware` v1 | v1 superseded by v2 (different interceptor model). | `.../go-grpc-middleware/v2` |
| `protoc-gen-validate` (PGV) | Deprecated predecessor of protovalidate; aging architecture. | `protovalidate` / `protovalidate-go` |
| `github.com/onsi/ginkgo` v1 (`v1.16.5`, still in root `go.mod`) | Superseded by Ginkgo v2; mixing v1/v2 causes confusion. | `github.com/onsi/ginkgo/v2` only |
| `github.com/go-chi/chi` as primary transport | Repo's `inventory` still wires chi HTTP, but PROJECT.md mandates **gRPC** as the transport. | gRPC inbound adapter; expose chi/`net/http` only for health/metrics/REST-gateway if needed. |
| Raw `protoc` + shell scripts for codegen | Fragile, machine-specific plugin versions, no breaking-change detection. | `buf generate` + `buf.gen.yaml` (pins plugin versions). |
| golangci-lint v1 config schema | v1 EOL'd at v1.64.8; new projects should not adopt the old `enable-all/disable-all` schema. | golangci-lint v2 with `linters.default`. |
| Heavy integration tests in pre-commit hooks | Makes every commit painful; devs will `--no-verify`. | Keep hooks fast (lint/format/unit); run full suites in CI. |
| Husky in a Go-only repo | Drags ~1,500 npm deps + Node runtime into a non-JS project. | Lefthook (single Go binary). |

## AI-Agent Conventions Knowledge Base — Format & Layout (milestone 1 core)

- **`AGENTS.md` is the cross-tool open standard** (released Aug 2025; donated to the Linux Foundation's Agentic AI Foundation Dec 2025; 60k+ repos). Read natively by Codex, Cursor, Copilot, Gemini CLI, Aider, Windsurf, Zed. It is plain, schema-free Markdown — "a README for agents."
- **`CLAUDE.md` (Anthropic) is richer for Claude Code** (3-layer memory model + import mechanism). The repo already has a `CLAUDE.md`. Recommended pattern: **keep `CLAUDE.md` as the Claude-specific entrypoint and add an `AGENTS.md`** (or symlink/point one at the other) so non-Claude agents are covered too. Don't duplicate prose — have one point to the canonical `memory-bank/`.
- **`memory-bank/` (versioned rule files) is the right home for detailed rules.** Keep `CLAUDE.md`/`AGENTS.md` *lean and pointer-based*; push the substance (testing, style, structure, libraries, architecture, processes) into topic files under `memory-bank/`. The repo's CLAUDE.md already references `memory-bank/testing.md`, `structure.md`, `libraries.md`, `agents.md` — formalize this set.

| Convention | Recommendation | Why |
|------------|----------------|-----|
| Entry files | `CLAUDE.md` + `AGENTS.md` at repo root, both thin, both pointing to `memory-bank/` | Cover Claude + every other agent without duplicating content. |
| Rule home | `memory-bank/` with one Markdown file per topic (`testing.md`, `style.md`, `structure.md`, `libraries.md`, `architecture.md`, `processes.md`, `git.md`) | Topic files stay small and avoid the "monolith dilutes priority" problem. |
| Requirement tiers | Phrase rules as **MUST / SHOULD / WON'T** | Yields consistent, unambiguous agent behavior. |
| Pointers over copies | Reference paths/commands; avoid pasting code that rots | Standard best practice — snippets go stale and mislead agents. |
| Scoped rules (optional) | `.claude/rules/*.md` with `paths:` frontmatter for path-specific guidance | Keeps the root file light; activates rules only for relevant files. |
| Language rule | Codify: Go domain comments/terminology in the chosen project language; test comments in English (per memory-bank) | Already a project decision — make it an explicit MUST. |

## Stack Patterns by Variant

- Add a gRPC-Gateway or adopt ConnectRPC for that service's inbound adapter.
- Because direct grpc-go has no browser transport; don't bolt a hand-rolled proxy on.
- Introduce `google/wire` (compile-time codegen) before reaching for runtime DI.
- Because it preserves Go's compile-time safety and zero startup overhead, matching the project's preference for explicit wiring.
- Keep the current in-process `AggregateRoot.PullEvents()` pattern for intra-service consistency; add an outbound message-broker adapter (e.g. NATS/Kafka) only at the integration boundary.
- Because the hexagon already isolates the `EventPublisher` port — swap the adapter, not the domain.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| grpc-go v1.81.x | Go 1.24.6 | Support policy = two latest Go majors; 1.24 is fine. |
| grpc-go v1.74.2 | go-grpc-middleware/v2 v2.3.3 | v2.3.3 was built against grpc 1.74.2; newer grpc patch releases are compatible. |
| mongo-driver/v2 v2.7.x | MongoDB server ≥ 4.4 | v2.7 requires server 4.4+; v3.6 is EOL in the driver. |
| golangci-lint v2.12.x | Go 1.24 | v2 schema only; do not mix with v1 `.golangci.yml`. |
| buf v1.71.x | protobuf APIv2 (`google.golang.org/protobuf`) | Pin codegen plugin versions in `buf.gen.yaml`. |
| ginkgo/v2 v2.23.4 | gomega v1.38.0 | Bump root module's stale gomega v1.37.0 / ginkgo v1 to match `pkg/`. |

## Sources

- pkg.go.dev / GitHub releases — grpc-go v1.81.1 (latest stable), Go support policy. **HIGH** (official, cross-checked)
- GitHub mongodb/mongo-go-driver releases + MongoDB docs — v2 GA Jan 2025, v2.7 requirements, v1 deprecation. **HIGH**
- github.com/golangci/golangci-lint/releases + golangci-lint.run changelog — v2.12.2 (2026-05-06), v2 schema changes. **HIGH** (official, cross-checked)
- github.com/grpc-ecosystem/go-grpc-middleware/releases — v2.3.3 (2024-11-04), built vs grpc 1.74.2; interceptor chain ordering (logging first, recovery last). **HIGH** (official release page fetched)
- github.com/bufbuild/buf/releases — buf CLI v1.71.0 (2026-06-16). **HIGH** (official release page fetched)
- buf.build/blog — protovalidate v1.0 GA (Sep 2025), PGV deprecation. **HIGH**
- connectrpc.com / Buf blog / benchmark posts — connect-go vs grpc-go tradeoffs. **MEDIUM** (vendor + community benchmarks)
- Medium/Leapcell/freeCodeCamp — Wire vs Fx vs manual DI guidance. **MEDIUM** (community consensus)
- openai.com / linuxfoundation.org / agentsmd.io / community — AGENTS.md standard, AAIF donation (Dec 2025), CLAUDE.md comparison, MUST/SHOULD/WON'T + pointer-based best practices. **MEDIUM-HIGH** (primary announcements + community synthesis)
- evilmartians/lefthook + comparison posts — Lefthook vs Husky for Go, commitlint commit-msg pattern. **MEDIUM** (official repo + community)
- Repo inspection (`go.work`, all `go.mod`, `CLAUDE.md`) — current pinned versions, chi/mongo-v1 drift to correct. **HIGH** (direct observation)

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:

- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
