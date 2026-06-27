---
phase: 05-dev
plan: 01
subsystem: infra
tags: [go, mongo-driver-v2, franz-go, kadm, kafka, mongodb, go-workspace, ginkgo]

# Dependency graph
requires:
  - phase: 05-dev (planning artefacts)
    provides: D-01..D-15 locked decisions, 05-RESEARCH verbatim API signatures, 05-PATTERNS analogs
provides:
  - "services/inventory мигрирован на mongo-driver/v2 v2.7.0 (v1.17.9 удалён); franz-go v1.21.4 + kadm v1.18.0 добавлены"
  - "internal/kafka/topology — single-source топологии: data-driven aggregates, имена топиков, cleanup-политики, Bootstrap(ctx, adm, partitions) на kadm"
  - "internal/repository/mongoconn — RS-aware Connect(ctx, uri) (фабрика + health-ping) на mongo-driver/v2, без UoW/repo"
affects: [05-02, 05-03, 05-04, 05-05, phase-06, phase-07, phase-08]

# Tech tracking
tech-stack:
  added:
    - "go.mongodb.org/mongo-driver/v2 v2.7.0"
    - "github.com/twmb/franz-go v1.21.4 (+ pkg/kadm v1.18.0)"
  patterns:
    - "Single-source топология топиков (D-06): имена/политики/партиции живут в одном пакете, зовут и CLI, и тесты"
    - "RS-aware connection-helper (D-14): тонкая client-фабрика + ping, без repository/UoW"
    - "Атомарный go.mod-свап под D-03: inventory всегда компилируется на каждом коммите"

key-files:
  created:
    - "services/inventory/internal/kafka/topology/topology.go"
    - "services/inventory/internal/kafka/topology/topology_test.go"
    - "services/inventory/internal/repository/mongoconn/conn.go"
  modified:
    - "services/inventory/go.mod"
    - "services/inventory/go.sum"
    - "go.work"
    - "go.work.sum"

key-decisions:
  - "uuid v1.6.0 удалён `go mod tidy` (нет импортёра в Phase 5); вернётся автоматически в Phase 6/7 при генерации внутреннего ID — go.mod держится tidy-консистентным (D-03)"
  - "go.work/go.mod подняты 1.24.6 → 1.25.0 установленным тулчейном (фактическая версия Go в среде — 1.25.0, не 1.24.6 как в PROJECT.md)"
  - "Task 2 закоммичен impl+test одним GREEN-коммитом, а не RED→GREEN: тест импортирует unexported helpers, RED-only коммит сломал бы компиляцию и нарушил D-03 (always-compiles invariant)"

patterns-established:
  - "Single-source топология (D-06): дублировать имена/cleanup-политики/партиции вне topology-пакета запрещено"
  - "Connection-helper-граница (D-14): только Connect+Ping; repository/UoW/Outbox — Phase 7"

requirements-completed: [SVC-05, SVC-07]

# Metrics
duration: 5min
completed: 2026-06-27
---

# Phase 5 Plan 01: Стек-фундамент inventory Summary

**Атомарная миграция inventory на mongo-driver/v2 (+ franz-go/kadm), single-source topology-пакет с Bootstrap на kadm и RS-aware Mongo connection-helper — workspace-build/vet/test зелёные.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-06-27T07:41:13Z
- **Completed:** 2026-06-27T07:45:59Z
- **Tasks:** 3
- **Files modified:** 7 (3 created, 4 modified)

## Accomplishments
- `services/inventory/go.mod` атомарно переведён с `mongo-driver v1.17.9` на `mongo-driver/v2 v2.7.0`; добавлены `franz-go v1.21.4` + `pkg/kadm v1.18.0`; `go mod tidy` пересобрал go.sum.
- Пакет `internal/kafka/topology` — единственный источник истины топологии (D-06): data-driven `aggregates=[host]`, деривация имён `inventory.host.events`/`.state`, cleanup-политики (`delete` / `compact`+24h `delete.retention.ms`), `Bootstrap(ctx, adm, partitions)` на `kadm.CreateTopics` (rf=1). Unit-тест констант зелёный без Docker.
- Пакет `internal/repository/mongoconn` — RS-aware `Connect(ctx, uri)`: `mongo.Connect`(v2, без ctx) + `writeconcern.Majority` + health-`Ping`, `Disconnect` при ошибке. Граница D-14 соблюдена (без UoW/repo).
- `go build ./...` + `go vet ./...` + `go test ./internal/kafka/topology/...` из модуля inventory — зелёные (SC1 + фундамент SC3).

## Task Commits

1. **Task 1: Атомарный свап go.mod на mongo-driver/v2 + franz-go/kadm** — `a422462` (chore)
2. **Task 2: Topology-пакет — константы агрегатов + Bootstrap на kadm** — `112803b` (feat, impl+test одним коммитом)
3. **Task 3: RS-aware Mongo connection-helper (v2)** — `8cdc6fd` (feat)

## Files Created/Modified
- `services/inventory/internal/kafka/topology/topology.go` — single-source топология: aggregates, имена топиков, cleanup-конфиги, `Bootstrap`.
- `services/inventory/internal/kafka/topology/topology_test.go` — unit-тест имён/политик (Ginkgo suite, без Docker).
- `services/inventory/internal/repository/mongoconn/conn.go` — RS-aware Mongo `Connect`+ping (v2).
- `services/inventory/go.mod` / `go.sum` — свап на mongo-driver/v2 + franz-go/kadm.
- `go.work` / `go.work.sum` — подняты до go 1.25.0 + sums новых модулей (тулчейн).

## Decisions Made
- **uuid v1.6.0 удалён `go mod tidy`** — нет ни одного импортёра в Phase 5 (`internal/` начинался пустым, новый код uuid не использует). Принудительно держать неиспользуемый direct-require нельзя: следующий `go mod tidy`/pre-push снова его уберёт, что нарушит D-03 (атомарный зелёный build). Зависимость вернётся естественно в Phase 6/7 при генерации внутреннего `ID`.
- **go 1.24.6 → 1.25.0** — фактический установленный тулчейн = Go 1.25.0 (env-метка «1.24.6» из PROJECT.md устарела); `go.work` уже был на 1.25.0 до плана, `go get`/`tidy` синхронизировали go-директиву go.mod. Это не дрейф, а выравнивание по реальной среде.
- **Task 2 — один GREEN-коммит вместо RED→GREEN** — `topology_test.go` тестирует unexported helpers (`eventsTopic`/`stateConfig`/…) из того же пакета; коммит «только тест» не скомпилировался бы и нарушил локированный D-03 («inventory всегда компилируется на каждом коммите»). D-03 как архитектурный инвариант приоритетнее `tdd="true"`-флага (config `tdd_mode: false`).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] uuid удалён `go mod tidy` вопреки `<done>` «uuid сохранён»**
- **Found during:** Task 1 (go mod tidy)
- **Issue:** План требовал сохранить `github.com/google/uuid v1.6.0`, но `go mod tidy` корректно удалил его как unused direct-require (нет импортёра в Phase 5). Оставлять его силой создаёт tidy-несогласованный go.mod, который pre-push/следующий tidy снова почистит — конфликт с D-03 (атомарный зелёный build).
- **Fix:** Оставил go.mod в tidy-консистентном состоянии без uuid. Зависимость вернётся автоматически в Phase 6/7 при появлении кода генерации `ID`.
- **Files modified:** services/inventory/go.mod, services/inventory/go.sum
- **Verification:** `go build ./...` + `go vet ./...` зелёные; Task 1 automated-gate (v2 present / v1 absent) проходит.
- **Committed in:** a422462 (Task 1 commit)

**2. [Rule 3 - Blocking] go-директива/go.work подняты до 1.25.0 тулчейном**
- **Found during:** Task 1 (go get)
- **Issue:** Установленный Go = 1.25.0; `go get` поднял `go.mod`/`go.work` go-директиву с 1.24.6 до 1.25.0. PROJECT.md упоминал 1.24.6.
- **Fix:** Принял выравнивание по реальному тулчейну (go.work уже был 1.25.0 до плана). Откат сломал бы build под установленным Go.
- **Files modified:** services/inventory/go.mod, go.work, go.work.sum
- **Verification:** workspace build/vet зелёные.
- **Committed in:** a422462 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (1 tidy-correctness bug, 1 blocking toolchain-version)
**Impact on plan:** Обе правки необходимы для tidy-консистентного зелёного build (D-03). Scope creep отсутствует — все артефакты плана созданы, automated-gates всех трёх задач зелёные.

## Issues Encountered
- `go build ./...` из **корня репозитория** падает `directory prefix . does not contain modules listed in go.work` — это корневой «rotten leftover» go.mod (`github.com/gwall-e`, go 1.23.6), вне go.work и явно неавторитетный (CLAUDE.md/boundaries.md). Per-module и per-package build/vet inventory зелёные. Out-of-scope, не трогался.

## Threat Flags

Нет нового security-surface вне `<threat_model>` плана. T-05-01/T-05-SC (supply-chain) митигированы: версии точные (не `@latest`), верифицированы по go-proxy в RESEARCH, go.sum фиксирует хеши.

## User Setup Required
None — внешняя конфигурация сервисов не требуется (топики только создаются Bootstrap'ом в последующих планах; этот план — только код).

## Next Phase Readiness
- SC1 (inventory собирается с mongo-driver/v2, v1 удалён, build/vet зелёные) — выполнен.
- Фундамент для Plan 05-02..05-05: `topology.Bootstrap` готов к вызову из CLI (05-02) и integration-тестов (05-03); `mongoconn.Connect` готов к smoke-тесту против testcontainers Mongo RS (05-03).
- Открытый downstream-долг (по плану, не этого плана): docker-compose, CLI провижна, mockery-обвязка, integration-тесты, lefthook D-02, каноны D-04 — планы 05-02/05-03.

## Self-Check: PASSED

All 4 created files exist on disk; all 3 task commits (a422462, 112803b, 8cdc6fd) present in git history.

---
*Phase: 05-dev*
*Completed: 2026-06-27*
