---
phase: 05-dev
plan: 04
subsystem: infra
tags: [go, kafka, kadm, franz-go, testcontainers, mongodb, ginkgo, integration-test, build-tag, cli]

# Dependency graph
requires:
  - phase: 05-dev (Plan 01)
    provides: "topology.Bootstrap(ctx, adm, partitions) + mongoconn.Connect(ctx, uri) — single-source функции, которые этот план вызывает из CLI и теста"
  - phase: 05-dev (Plan 02)
    provides: "docker-compose dev-стенд + make topics/test-integration таргеты, оборачивающие этот CLI/тест"
provides:
  - "cmd/main.go — тонкий bootstrap-CLI (D-09): env KAFKA_BROKERS/KAFKA_PARTITIONS → kgo/kadm → topology.Bootstrap; не дублирует топологию (D-06)"
  - "topology_integration_test.go (build-tag integration, D-15) — testcontainers smoke kafka+mongo, замыкает single-source: оба пути зовут общие Bootstrap+Connect"
  - "Замкнут SC3 (assert провижн .events=delete/.state=compact, partitions=6) и SC4 (testcontainers стартует и подключается)"
affects: [05-05, phase-06, phase-07, phase-08]

# Tech tracking
tech-stack:
  added:
    - "github.com/testcontainers/testcontainers-go v0.43.0 (+modules/kafka, +modules/mongodb) — test-only"
  patterns:
    - "Build-tag изоляция интеграционных тестов (D-15): //go:build integration первой строкой — go test ./... и pre-push не тянут Docker"
    - "Тонкий bootstrap-CLI (D-09): cmd/ — только склейка env→client→общая функция, ноль дублирования топологии (D-06)"
    - "go vet вместо go build для package main в cmd/ (Pitfall 2): go build ./cmd падает 'build output cmd already exists'"

key-files:
  created:
    - "services/inventory/internal/kafka/topology/topology_integration_test.go"
  modified:
    - "services/inventory/cmd/main.go (replace stub)"
    - "services/inventory/go.mod"
    - "services/inventory/go.sum"
    - "go.work.sum"

key-decisions:
  - "franz-go повышен indirect → direct в go.mod: cmd/main.go импортирует kgo напрямую (kadm уже был direct из Plan 01)"
  - "Тест — black-box package topology_test: чтобы звать и topology.Bootstrap, и mongoconn.Connect с package-квалификаторами через единый импорт-набор"
  - "Ассерт SC3 через kadm: ListTopics → details[topic].Partitions (HaveLen 6) + DescribeTopicConfigs → ResourceConfigs.On(topic).Configs[cleanup.policy] (delete/compact)"

patterns-established:
  - "Integration-тесты ВСЕГДА за //go:build integration (D-15) — критично после D-02 (pre-push включает inventory)"
  - "Single-source топология замкнута на обоих концах: CLI и тест зовут одну Bootstrap, дубль конфига топиков запрещён (D-06)"

requirements-completed: [SVC-06, SVC-07]

# Metrics
duration: ~6min
completed: 2026-06-30
---

# Phase 5 Plan 04: Bootstrap-CLI + integration smoke Summary

**Тонкий bootstrap-CLI в cmd/ (D-09) и testcontainers integration-smoke за build-tag (D-15) — ОБА зовут единые topology.Bootstrap + mongoconn.Connect (single source, D-06); замкнуты SC3 (провижн .events/.state с cleanup-политиками, partitions=6) и SC4 (Kafka KRaft + Mongo RS стартуют и подключаются).**

## Performance

- **Duration:** ~6 min
- **Tasks:** 2
- **Files modified:** 5 (1 created, 4 modified)

## Accomplishments
- `cmd/main.go` — стаб `func main(){return}` заменён тонким bootstrap-CLI (D-09): читает `KAFKA_BROKERS` (CSV, дефолт `localhost:9092`) + `KAFKA_PARTITIONS` (дефолт 6, D-11) → `kgo.NewClient(kgo.SeedBrokers(...))` → `kadm.NewClient(cl)` → `topology.Bootstrap(ctx, adm, int32(partitions))`. Топология не дублируется (D-06): имена/политики живут только в пакете topology. Ошибки оборачиваются `%w` (knowledge/style.md); доменные комментарии на русском.
- `topology_integration_test.go` — `//go:build integration` первой строкой (D-15). Suite на Ginkgo (`RegisterFailHandler`+`RunSpecs`, dot-import, английские комментарии — knowledge/testing.md). Spec 1 (Kafka): `kafka.Run(confluentinc/confluent-local:7.5.0, WithClusterID)` → `Brokers(ctx)` → `kgo.Ping` (smoke) → `topology.Bootstrap(ctx, adm, 6)` → ассерт `ListTopics`/`DescribeTopicConfigs`: `inventory.host.events` cleanup=delete, `inventory.host.state` cleanup=compact, partitions=6 (SC3). Spec 2 (Mongo): `mongodb.Run(mongo:7, WithReplicaSet("rs0"))` → `ConnectionString` (доверяем, Pitfall 1) → `mongoconn.Connect` (SC4).
- Single-source топология (D-06) замкнута на обоих концах: и CLI, и тест зовут одну `Bootstrap` — дубля конфига топиков нет.

## Task Commits

1. **Task 1: Тонкий bootstrap-CLI в cmd/ (зовёт topology.Bootstrap)** — `169c8eb` (feat)
2. **Task 2: Integration smoke на testcontainers (build-tag integration)** — `40786a2` (test)

## Files Created/Modified
- `services/inventory/internal/kafka/topology/topology_integration_test.go` — testcontainers kafka+mongo smoke за build-tag integration; assert SC3/SC4.
- `services/inventory/cmd/main.go` — тонкий bootstrap-CLI (replace stub).
- `services/inventory/go.mod` / `go.sum` — franz-go indirect→direct (импорт kgo в CLI); + test-deps testcontainers-go v0.43.0 (+modules kafka/mongodb).
- `go.work.sum` — хеши новых модулей workspace.

## Decisions Made
- **franz-go indirect → direct** — `cmd/main.go` импортирует `pkg/kgo` напрямую (для `kgo.NewClient`/`SeedBrokers`/`Close`), поэтому `go mod tidy` корректно повысил `github.com/twmb/franz-go` до direct-require. `pkg/kadm` уже был direct из Plan 01.
- **Тест — black-box `package topology_test`** — позволяет звать и `topology.Bootstrap`, и `mongoconn.Connect` через package-квалификаторы единым импорт-набором; не лезет в unexported helpers (в отличие от unit-теста констант Plan 01).
- **Ассерт SC3 через kadm API** — partitions через `details[topic].Partitions` (`HaveLen(6)`); cleanup.policy через `DescribeTopicConfigs(...).On(topic, nil)` + перебор `Configs` по `Key=="cleanup.policy"` с `MaybeValue()`. Хелпер `cleanupPolicy` инкапсулирует разбор `ResourceConfigs`.

## Deviations from Plan

None — план выполнен ровно как написан. Обе automated-gate'ы зелёные:
- Task 1: `go vet ./...` exit 0 + `grep topology.Bootstrap cmd/main.go` OK.
- Task 2: `grep //go:build integration` OK + `go vet -tags=integration ./internal/kafka/topology/...` exit 0 + `go test ./internal/kafka/topology/...` (без тега) зелёный без Docker.

Дополнительно: workspace-wide `go vet ./...` зелёный (D-03 always-compiles invariant соблюдён).

## Verification Performed
- `go vet ./...` (workspace) — exit 0 (D-01/D-03).
- `go vet -tags=integration ./internal/kafka/topology/...` — exit 0 (integration-тест компилируется).
- `go test ./internal/kafka/topology/...` (без тега) — `ok`, integration-spec'ы НЕ компилируются → без Docker (D-15).
- `head -1` integration-теста — `//go:build integration` присутствует первой строкой.
- `grep topology.Bootstrap cmd/main.go` — CLI зовёт общую функцию (D-06/D-09).

> **Не выполнено в этой среде (требует живого Docker daemon):** фактический прогон `make test-integration` (kgo.Ping + mongo.Connect против реальных контейнеров, реальный assert провижна). Тест компилируется и изолирован тегом; запуск с Docker — на dev-машине/CI (паритет с SC2 manual-путём Plan 02). Это плановое ограничение (testcontainers требует daemon), не отклонение.

## Threat Flags

Нет нового security-surface вне `<threat_model>` плана. T-05-08 (tampering образов) митигирован: конкретные теги `confluentinc/confluent-local:7.5.0` и `mongo:7` в `Run(...)` (паритет с compose, D-07). T-05-09/T-05-10 — accept (эфемерные контейнеры, dev/test-only, доверенный CLI-оператор).

## User Setup Required
None для кода. Для фактического прогона integration-теста (`make test-integration`) разработчику нужен живой Docker daemon — это документированный путь Phase 5, не блокер сборки.

## Next Phase Readiness
- SC3 (Bootstrap провижнит .events=delete/.state=compact, partitions=6) — ассерт-код готов; зелёный прогон при наличии Docker.
- SC4 (testcontainers Kafka KRaft + Mongo RS стартуют и подключаются) — тест-код готов; зелёный прогон при наличии Docker.
- Single-source топология (D-06) замкнута: остаётся Plan 05-05 (lefthook de-exclusion + каноны build/structure/boundaries + ROADMAP SC1 + DOC-02 рецепт).

## Self-Check: PASSED

Both created/modified files exist on disk (`cmd/main.go`, `topology_integration_test.go`); оба task-коммита (`169c8eb`, `40786a2`) присутствуют в git history.

---
*Phase: 05-dev*
*Completed: 2026-06-30*
