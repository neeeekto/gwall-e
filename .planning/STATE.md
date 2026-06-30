---
gsd_state_version: 1.0
milestone: v3.0
milestone_name: — Inventory + Event-backbone
status: completed
stopped_at: Phase 6 context gathered
last_updated: "2026-06-30T18:12:25.440Z"
last_activity: "2026-06-30 -- Phase 05 verified 5/5; post-exec рефактор: generic Kafka/Mongo → pkg/; строки ошибок → EN"
progress:
  total_phases: 6
  completed_phases: 1
  total_plans: 5
  completed_plans: 5
  percent: 17
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-06-26)

**Core value:** Безопасное и согласованное управление парком серверов как услугой — единый источник правды о хостах.
**Current focus:** Phase 05 — dev

## Current Position

Phase: 05 (dev) — COMPLETE (verified 5/5)
Plan: 5 of 5
Status: Phase complete — все must-have критерии подтверждены; SC2 dev-стенд smoke approved пользователем (rs.status().ok==1, Kafka :9092)
Last activity: 2026-06-30 -- Phase 05 verified 5/5; post-exec рефактор: generic Kafka/Mongo → pkg/; строки ошибок → EN

Progress: [██████████] 100%

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
| Phase 05 P03 | ~2min | 2 tasks | 7 files |
| Phase 05 P04 | ~6min | 2 tasks | 5 files |
| Phase 05 P05 | ~8min | 3 tasks | 7 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table. Recent decisions affecting current work:

- **Build order v3.0 (жёсткий):** инфра/стек → доменная модель+glossary → эталон записи (UoW+Outbox+gRPC) → схемы+relay→Kafka → топология → верификация. Forward-compat envelope (`eventId`/`version`/`actor`, SEED-002/003) в доменных событиях (Phase 6) + protobuf (Phase 8) **до** relay-кода.
- **Идентичность = внутренний постоянный `ID`**; re-add = новый ID без авто-мерджа; FQDN уникален среди `active` (partial index); `decommissioned` ≠ `deleted` (история на событиях, не soft-delete-флаг).
- **Event-backbone = dual-topic per aggregate:** `*.events` (cleanup=delete, история фактов) + `*.state` (cleanup=compact by entityID, снапшот/backfill). Kafka key = внутренний ID; tombstone только на терминальный `deleted`.
- **Стек v3.0:** franz-go (idempotent producer by default), mongo-driver/**v2** (v1 deprecated — миграция до репозиториев), testcontainers (KRaft + Mongo single-node RS). Schema registry не вводится (продюсер-only).
- Канон v1.0 не пересматривается: DDD+гексагон без CQRS-шины, UoW (Mongo-txn), transactional outbox в той же txn, relay — отдельный async-процесс (нет dual-write).
- [Phase ?]: Dev-стенд (Plan 05-02): mongo запускается явным mongod, иначе --replSet не доходит до демона; SC2 smoke пройден вручную (rs.status().ok==1, Kafka :9092)
- [Phase 5]: mockery v3 smoke (Plan 05-03): .mockery.yaml v3-синтаксис (template testify, .SrcPackageName, моки в {{.InterfaceDir}}/mocks); throwaway example-пакет доказал кодоген до реальных портов (Phase 6/7); testify втянут тест-кодом (T-05-07 accept)
- [Phase 5]: bootstrap-CLI + integration smoke (Plan 05-04): тонкий cmd/main.go (env→kgo/kadm→topology.Bootstrap, D-09, без дубля топологии D-06); integration-тест за //go:build integration (D-15 — go test ./... и pre-push без Docker); single-source замкнут (CLI и тест зовут одну Bootstrap); SC3/SC4 ассерт-код готов, фактический прогон требует Docker (make test-integration); go build ./cmd падает (Pitfall 2) — валидация через go vet
- [Phase 5][post-exec]: язык кода — комментарии RU (canon-MUST style.md), но строки ошибок/логов → EN (established-конвенция pkg/http/errors.go); поправлено в main.go/topology.go/conn.go (commit e4b7ddb)
- [Phase 5][post-exec]: **правило** — все общие/generic элементы живут в `pkg/`, не дублируются между сервисами. Generic Kafka-механика → `pkg/kafka` (NewAdminClient/EnsureTopics/TopicSpec/StringPtr), RS-aware Mongo connect → `pkg/mongoconn`. Доменная топология топиков остаётся в inventory; D-06 single-source сохранён (topology.Bootstrap делегирует в pkg/kafka). pkg go-директива → 1.25.0 (mongo-driver/v2); inventory: require+replace на локальный pkg (commits 8b2025f/4e293b0/e7423b6)
- [Phase 5]: lefthook de-exclusion + каноны (Plan 05-05): pre-push гоняет inventory unit (go test ./..., без integration build tag, D-15); GOWORK=off снят везде (lefthook/build/structure/boundaries + drift в testing.md/README.md, T-05-12); DOC-02 закрыт — audit-рецепт go vet ./... exit 0 (go build ./... падает build output cmd already exists); каноны: inventory = четвёртый полноправный член go.work, всегда компилируется (D-01/D-03/D-04); pre-push smoke зелёный без Docker; live-firing требует разового lefthook install

### Pending Todos

None yet.

### Blockers/Concerns

- **Pitfall-гейты (ресёрч, HIGH):** (1) compacted ≠ история — нужен отдельный `*.events`; (2) tombstone только на delete; (3) Kafka key = ID, не FQDN/INV/MAC; (4) relay `ORDER BY sequence`, не `created_at`; (5) outbox-append в той же UoW-txn (тест atomicity). Все верифицируются в Phase 8/10 чеклистом «Looks Done But Isn't».
- **LC1/LC2 граница с Health:** `failed`/`maintenance` как факт-статус vs health-flag — решить в DOC-07 glossary (Phase 6) до статусной машины.

## Deferred Items

Items carried forward from v1.0 milestone close (2026-06-17); адресуются в v3.0 где указано:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| requirement | DOC-02 — build.md audit-рецепт падает (exit 1); рабочие формы `go build ./cmd` / `go vet ./...` | resolved (Plan 05-05: go vet ./... exit 0) | 2026-06-17 |
| requirement | DOC-07 glossary (ubiquitous language) — активируется в domain-milestone | deferred → Phase 6 | 2026-06-17 |
| verification | Phase 02 — 02-VERIFICATION.md (DOC-02 build claim) | gaps_found | 2026-06-17 |
| verification | Phase 04 — 04-VERIFICATION.md (live hook firing needs one-time bootstrap) | human_needed | 2026-06-17 |
| uat | Phase 04 — 04-UAT.md: 7 pending live-firing scenarios (`make tools` + `lefthook install`) | testing | 2026-06-17 |
| nyquist | Phases 1–2 no VALIDATION.md; Phases 3–4 nyquist_compliant=false (sign-off pending) | partial | 2026-06-17 |

## Session Continuity

Last session: 2026-06-30T18:12:25.430Z
Stopped at: Phase 6 context gathered
Resume file: .planning/phases/06-inventory/06-CONTEXT.md

## Operator Next Steps

- Phase 05 завершена и верифицирована (5/5). Следующий шаг: `/gsd-discuss-phase 6` → `/gsd-plan-phase 6` (Доменная модель Inventory: агрегаты Project/Host, glossary DOC-07).
- Опционально: разовый `make tools` + `lefthook install` на dev-машине для live-firing хуков; `make test-integration` для фактического прогона testcontainers-smoke (требует Docker).
