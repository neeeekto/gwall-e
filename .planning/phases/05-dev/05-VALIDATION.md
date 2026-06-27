---
phase: 5
slug: dev
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-27
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `05-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Ginkgo v2.23.4 + Gomega v1.38.0 (pinned в `pkg/go.mod`); mockery v3.7.1 |
| **Config file** | none — `RegisterFailHandler(Fail)` + `RunSpecs` в `TestXxx` (knowledge/testing.md); `.mockery.yaml` добавляется Wave 0 |
| **Quick run command** | `go build ./...` + `go vet ./...` (из корня workspace, D-01; без Docker) |
| **Full suite command** | `go test ./...` (unit) → `make test-integration` = `go test -tags=integration ./...` (testcontainers, D-15) |
| **Estimated runtime** | unit ~секунды; integration ~30–90 с (старт KRaft + Mongo RS контейнеров) |

---

## Sampling Rate

- **After every task commit:** Run `go vet ./...` + `go build ./...` (быстро, без Docker)
- **After every plan wave:** Run `go test ./...` (unit) + `make test-integration` (если Docker доступен)
- **Before `/gsd-verify-work`:** Full unit+integration зелёные; `make dev-up` ручной compose-smoke (SC2)
- **Max feedback latency:** unit < 30 s; integration < 120 s

---

## Per-Task Verification Map

> Planner fills one row per task (PLAN.md) with the concrete automated command.
> Mapping of Success Criteria → validation type below is authoritative.

| SC / Req | Behavior | Test Type | Automated Command / Observable | File Exists | Status |
|----------|----------|-----------|--------------------------------|-------------|--------|
| SC1 / SVC-05 | inventory собирается с mongo-driver/v2, v1 удалён | build (unit-gate) | `go build ./...` + `go vet ./...` зелёные; `go.mod` содержит только `mongo-driver/v2`, нет `v1.17.9` | ❌ W0 | ⬜ pending |
| SC2 / SVC-07 | `docker compose up` поднимает Kafka KRaft + Mongo RS | manual compose smoke | `make dev-up`; `rs.status().ok==1` (`?directConnection=true`); Kafka на :9092 | ❌ W0 | ⬜ pending |
| SC3 / SVC-07 | Bootstrap провижнит `*.events`(delete) + `*.state`(compact) | integration | `Bootstrap(ctx, adm, 6)` → `DescribeTopicConfigs` показывает cleanup.policy=delete/compact, partitions=6 | ❌ W0 | ⬜ pending |
| SC4 / SVC-06 | testcontainers (Kafka KRaft + Mongo RS) стартует; Ginkgo/Gomega/mockery smoke | integration + unit | `go test -tags=integration ./...` зелёный (kgo.Ping + mongo.Ping); `make generate-mocks` генерит мок; unit-spec с моком зелёный | ❌ W0 | ⬜ pending |
| SC5 / DOC-02 | build.md audit-рецепт проходит (exit 0) | manual/CI | новый рецепт (`go vet ./...`) → `echo $?` == 0; build.md обновлён | ❌ W0 | ⬜ pending |
| D-02 | pre-push включает inventory (unit only) | hook | `lefthook.yml` pre-push гоняет inventory unit-тесты; integration за тегом не запускается | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `services/inventory/go.mod` — свап на mongo-driver/v2 v2.7.0, `go mod tidy` (SC1)
- [ ] `internal/kafka/topology/` — const топологии + `Bootstrap` на `kadm` (SC3) + unit-тест констант
- [ ] connection-helper (Connect + Ping, RS-aware, writeconcern.Majority) в infra/repositories-слое (SC4)
- [ ] `*_integration_test.go` с `//go:build integration` — testcontainers KRaft+Mongo RS smoke (SC3/SC4)
- [ ] `.mockery.yaml` + throwaway-интерфейс + сгенерированный мок + unit-spec с моком (SC4)
- [ ] `docker-compose.yml` (confluent-local + mongo:7 RS healthcheck) (SC2)
- [ ] bootstrap CLI в `cmd/` → make-таргет (SC3 manual path)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `docker compose up` поднимает Kafka KRaft + Mongo RS, транзакции доступны | SVC-07 / SC2 | Требует Docker engine + compose-плагин (нет в sandbox); проверяется на dev-машине | `make dev-up`; затем `mongosh "mongodb://localhost:27017/?directConnection=true" --eval 'rs.status().ok'` → 1; Kafka reachable на :9092 |

---

## Validation Sign-Off

- [ ] All tasks have automated verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
