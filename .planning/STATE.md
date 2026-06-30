---
gsd_state_version: 1.0
milestone: v3.0
milestone_name: — Inventory + Event-backbone
status: executing
stopped_at: Phase 5 context gathered
last_updated: "2026-06-30T15:46:17.262Z"
last_activity: 2026-06-27 -- Phase 05 execution started
progress:
  total_phases: 6
  completed_phases: 0
  total_plans: 5
  completed_plans: 2
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-06-26)

**Core value:** Безопасное и согласованное управление парком серверов как услугой — единый источник правды о хостах.
**Current focus:** Phase 05 — dev

## Current Position

Phase: 05 (dev) — EXECUTING
Plan: 3 of 5
Status: Ready to execute
Last activity: 2026-06-27 -- Phase 05 execution started

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**

- Total plans completed: 0 (в v3.0; v1.0 — 14 планов shipped)
- Average duration: —
- Total execution time: —

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 5 | 0 | - | - |

**Recent Trend:**

- Last 5 plans: —
- Trend: —

*Updated after each plan completion*
| Phase 05 P01 | 5min | 3 tasks | 7 files |
| Phase 05 P02 | 3 days | 3 tasks | 2 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table. Recent decisions affecting current work:

- **Build order v3.0 (жёсткий):** инфра/стек → доменная модель+glossary → эталон записи (UoW+Outbox+gRPC) → схемы+relay→Kafka → топология → верификация. Forward-compat envelope (`eventId`/`version`/`actor`, SEED-002/003) в доменных событиях (Phase 6) + protobuf (Phase 8) **до** relay-кода.
- **Идентичность = внутренний постоянный `ID`**; re-add = новый ID без авто-мерджа; FQDN уникален среди `active` (partial index); `decommissioned` ≠ `deleted` (история на событиях, не soft-delete-флаг).
- **Event-backbone = dual-topic per aggregate:** `*.events` (cleanup=delete, история фактов) + `*.state` (cleanup=compact by entityID, снапшот/backfill). Kafka key = внутренний ID; tombstone только на терминальный `deleted`.
- **Стек v3.0:** franz-go (idempotent producer by default), mongo-driver/**v2** (v1 deprecated — миграция до репозиториев), testcontainers (KRaft + Mongo single-node RS). Schema registry не вводится (продюсер-only).
- Канон v1.0 не пересматривается: DDD+гексагон без CQRS-шины, UoW (Mongo-txn), transactional outbox в той же txn, relay — отдельный async-процесс (нет dual-write).
- [Phase ?]: Dev-стенд (Plan 05-02): mongo запускается явным mongod, иначе --replSet не доходит до демона; SC2 smoke пройден вручную (rs.status().ok==1, Kafka :9092)

### Pending Todos

None yet.

### Blockers/Concerns

- **Pitfall-гейты (ресёрч, HIGH):** (1) compacted ≠ история — нужен отдельный `*.events`; (2) tombstone только на delete; (3) Kafka key = ID, не FQDN/INV/MAC; (4) relay `ORDER BY sequence`, не `created_at`; (5) outbox-append в той же UoW-txn (тест atomicity). Все верифицируются в Phase 8/10 чеклистом «Looks Done But Isn't».
- **LC1/LC2 граница с Health:** `failed`/`maintenance` как факт-статус vs health-flag — решить в DOC-07 glossary (Phase 6) до статусной машины.

## Deferred Items

Items carried forward from v1.0 milestone close (2026-06-17); адресуются в v3.0 где указано:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| requirement | DOC-02 — build.md audit-рецепт падает (exit 1); рабочие формы `go build ./cmd` / `go vet ./...` | gaps_found → Phase 5 | 2026-06-17 |
| requirement | DOC-07 glossary (ubiquitous language) — активируется в domain-milestone | deferred → Phase 6 | 2026-06-17 |
| verification | Phase 02 — 02-VERIFICATION.md (DOC-02 build claim) | gaps_found | 2026-06-17 |
| verification | Phase 04 — 04-VERIFICATION.md (live hook firing needs one-time bootstrap) | human_needed | 2026-06-17 |
| uat | Phase 04 — 04-UAT.md: 7 pending live-firing scenarios (`make tools` + `lefthook install`) | testing | 2026-06-17 |
| nyquist | Phases 1–2 no VALIDATION.md; Phases 3–4 nyquist_compliant=false (sign-off pending) | partial | 2026-06-17 |

## Session Continuity

Last session: 2026-06-30T15:45:49.972Z
Stopped at: Phase 5 context gathered
Resume file: .planning/phases/05-dev/05-CONTEXT.md

## Operator Next Steps

- Plan first phase with `/gsd-plan-phase 5`
