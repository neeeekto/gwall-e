---
phase: 05-dev
verified: 2026-06-30T16:25:50Z
status: human_needed
score: 5/5 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Запустить make dev-up, дождаться healthy-статуса mongo, проверить rs.status().ok==1, убедиться что Kafka достижима на localhost:9092"
    expected: "docker compose ps показывает оба сервиса healthy; mongosh --eval 'rs.status().ok' возвращает 1; Kafka broker слушает на :9092 без fatal-ошибок в логах"
    why_human: "SC2 требует живого Docker engine + compose; sandbox не имеет Docker daemon; пользователь подтвердил smoke вручную (коммит 1cb00e9 — verified config), но это не автоматически проверяемое свидетельство"
---

# Phase 5: Dev-инфра и стек — Verification Report

**Phase Goal:** Готовое окружение и обновлённый стек, на котором можно писать репозитории и гонять интеграционные тесты — до первого доменного кода.
**Verified:** 2026-06-30T16:25:50Z
**Status:** human_needed
**Re-verification:** No — initial verification

---

## Goal Achievement

Цель фазы — рабочее dev-окружение + обновлённый стек перед первым доменным кодом. Все пять Success Criteria имеют кодовое воплощение; SC2 (docker compose up) задокументировано как manual-verified пользователем.

### Observable Truths

| #  | Truth (Success Criteria)                                                                              | Status     | Evidence                                                                                  |
|----|-------------------------------------------------------------------------------------------------------|------------|-------------------------------------------------------------------------------------------|
| 1  | `services/inventory` собирается с mongo-driver/v2; v1 удалён; go build/vet зелёные; inventory в go.work | VERIFIED   | `go build ./... && go vet ./...` → exit 0; go.mod содержит `mongo-driver/v2 v2.7.0`, нет v1.17.9; go.work включает `./services/inventory` |
| 2  | `docker compose up` поднимает Kafka KRaft + Mongo single-node RS (транзакции доступны)               | VERIFIED*  | docker-compose.yml существует, содержит confluent-local:7.5.0 (KRaft) + mongo:7 с idempotent rs.initiate healthcheck; SC2 подтверждён пользователем вручную (коммит 1cb00e9) |
| 3  | Bootstrap провижнит `inventory.*.events` (cleanup=delete) и `inventory.*.state` (cleanup=compact)    | VERIFIED   | topology.go — single-source (D-06); unit-тест констант зелёный; integration-тест ассертирует cleanup-политики через DescribeTopicConfigs; go vet -tags=integration exit 0 |
| 4  | Ginkgo v2 + Gomega + mockery подключены и проходят smoke-прогон; testcontainers-тест компилируется   | VERIFIED   | `go test -count=1 ./...` → exit 0 (example + topology unit-suite); topology_integration_test.go с `//go:build integration` компилируется (`go vet -tags=integration` exit 0); мок сгенерирован (102 строки) |
| 5  | build.md audit-рецепт (DOC-02) исправлен: `cd services/audit && go vet ./...` → exit 0              | VERIFIED   | `go vet ./...` из services/audit → exit 0 (проверено в ходе верификации); build.md содержит `go vet ./...`; нет GOWORK=off в канонах |

*SC2 — human_verified; автоматическая проверка невозможна без Docker daemon.

**Score:** 5/5 truths verified (SC2 требует human sign-off)

---

### Required Artifacts

| Артефакт                                                                                      | Ожидание (из PLAN)                                          | Уровень           | Статус     | Детали                                                                  |
|-----------------------------------------------------------------------------------------------|-------------------------------------------------------------|-------------------|------------|-------------------------------------------------------------------------|
| `services/inventory/go.mod`                                                                   | mongo-driver/v2 v2.7.0; нет v1                             | Exists+Substantive | VERIFIED   | indirect через pkg/mongoconn (ожидаемо после pkg-split); прямое использование в pkg/go.mod |
| `pkg/mongoconn/conn.go`                                                                       | RS-aware Connect+Ping на mongo-driver/v2 (D-14)            | Exists+Substantive+Wired | VERIFIED   | 43 строки; импортирует v2; используется integration-тестом и будущими репозиториями |
| `pkg/kafka/admin.go`                                                                          | NewAdminClient + EnsureTopics + StringPtr (извлечение механики) | Exists+Substantive+Wired | VERIFIED   | 50 строк; используется topology.go через `kafka.EnsureTopics`          |
| `services/inventory/internal/kafka/topology/topology.go`                                     | data-driven aggregates + Bootstrap на kadm (D-06)           | Exists+Substantive+Wired | VERIFIED   | 84 строки; Bootstrap делегирует pkg/kafka; нет дублирования конфигов   |
| `services/inventory/internal/kafka/topology/topology_test.go`                                | unit-тест имён/cleanup-политик (без Docker)                 | Exists+Substantive+Wired | VERIFIED   | 47 строк; Ginkgo suite; тестирует `eventsTopic`, `stateTopic`, `eventsConfig`, `stateConfig`, `aggregates` |
| `services/inventory/internal/kafka/topology/topology_integration_test.go`                    | `//go:build integration`; testcontainers kafka+mongo        | Exists+Substantive | VERIFIED   | 112 строк; `//go:build integration` первой строкой; ассертирует SC3/SC4 |
| `services/inventory/cmd/main.go`                                                              | тонкий bootstrap-CLI; зовёт topology.Bootstrap (D-09)      | Exists+Substantive+Wired | VERIFIED   | 93 строки; читает KAFKA_BROKERS/KAFKA_PARTITIONS; вызывает `kafka.NewAdminClient` + `topology.Bootstrap`; не дублирует топологию |
| `docker-compose.yml`                                                                          | confluent-local:7.5.0 + mongo:7 RS + healthcheck           | Exists+Substantive | VERIFIED*  | 39 строк; образы запиннены; idempotent rs.initiate; SC2 manual-verified |
| `Makefile`                                                                                    | MOCKERY_VERSION pin + dev-up/dev-down/topics/test-integration/generate-mocks | Exists+Substantive | VERIFIED   | Все таргеты присутствуют; MOCKERY_VERSION := v3.7.1; нет `go build ./cmd` |
| `.mockery.yaml`                                                                               | v3-конфиг (template testify); нацелен на example-пакет     | Exists+Substantive+Wired | VERIFIED   | v3-синтаксис; `template: testify`; packages → `internal/example`       |
| `services/inventory/internal/example/provisioner.go`                                        | throwaway ExampleProvisioner интерфейс                      | Exists+Substantive | VERIFIED   | 29 строк; типизированный ExampleID; sentinel ErrExampleProvisionFailed  |
| `services/inventory/internal/example/mocks/ExampleProvisioner.go`                           | сгенерированный mockery v3 мок                              | Exists+Substantive+Wired | VERIFIED   | 102 строки кодогена; NewMockExampleProvisioner; использован в provisioner_test.go |
| `services/inventory/internal/example/provisioner_test.go`                                   | Ginkgo unit-spec с mock (без build-tag)                     | Exists+Substantive+Wired | VERIFIED   | 59 строк; `go test` exit 0; тестирует success+failure через EXPECT().Return |
| `lefthook.yml`                                                                                | pre-push test-inventory (unit only); нет GOWORK=off        | Exists+Substantive | VERIFIED   | `cd services/inventory && go test ./...` присутствует; нет GOWORK=off; нет integration-тега |
| `knowledge/build.md`                                                                         | audit-рецепт go vet; нет GOWORK=off; inventory как член workspace | Exists+Substantive | VERIFIED   | go vet рецепт присутствует; нет GOWORK=off; раздел «inventory — член workspace» |
| `knowledge/structure.md`                                                                     | inventory = четыре члена go.work                            | Exists+Substantive | VERIFIED   | «Активные модули workspace — ровно четыре»; список включает `services/inventory` |
| `knowledge/boundaries.md`                                                                    | inventory убран из WIP-лесов; нет GOWORK=off               | Exists+Substantive | VERIFIED   | GOWORK=off не найден; inventory помечен как полноправный член go.work   |
| `.planning/ROADMAP.md`                                                                       | Phase 5 SC1 без «с GOWORK=off»                              | Exists+Substantive | VERIFIED   | GOWORK=off не найден в ROADMAP.md                                       |

---

### Key Link Verification

| От                                              | До                              | Через                                     | Статус   | Детали                                                                   |
|-------------------------------------------------|---------------------------------|-------------------------------------------|----------|--------------------------------------------------------------------------|
| `topology.go`                                   | `pkg/kafka.EnsureTopics`        | `kafka.EnsureTopics(ctx, adm, partitions, replicationFactor, specs()...)` | WIRED    | Строка 82 topology.go; делегирование механики в pkg/kafka подтверждено  |
| `topology.go`                                   | `pkg/kafka.StringPtr`           | `kafka.StringPtr("delete")` и `kafka.StringPtr("compact")` | WIRED    | Строки 51, 58, 59 topology.go                                            |
| `cmd/main.go`                                   | `topology.Bootstrap`            | `topology.Bootstrap(ctx, adm, int32(partitions))` | WIRED    | Строка 56 cmd/main.go; grep подтверждён                                  |
| `cmd/main.go`                                   | `pkg/kafka.NewAdminClient`      | `kafka.NewAdminClient(brokers)`            | WIRED    | Строка 49 cmd/main.go; CLI использует pkg/kafka (не kadm напрямую)       |
| `topology_integration_test.go`                  | `topology.Bootstrap`            | `topology.Bootstrap(ctx, adm, partitions)` | WIRED    | Строка 62 integration-теста; single-source (D-06) замкнут               |
| `topology_integration_test.go`                  | `pkg/mongoconn.Connect`         | `mongoconn.Connect(ctx, uri)`              | WIRED    | Строки 24+95; импортирует `github.com/gwall-e/pkg/mongoconn`             |
| `docker-compose.yml mongo healthcheck`          | `rs.initiate`                   | idempotent try/catch healthcheck           | WIRED    | `rs.initiate({_id:'rs0',members:[{_id:0,host:'127.0.0.1:27017'}]})` присутствует |
| `Makefile generate-mocks`                       | `mockery`                       | `MOCKERY_VERSION := v3.7.1`                | WIRED    | `go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)` в tools    |
| `.mockery.yaml`                                 | `services/inventory/internal/example` | `interfaces: { ExampleProvisioner: {} }` | WIRED    | packages → `github.com/gwall-e/services/inventory/internal/example`     |
| `provisioner_test.go`                           | `mocks.NewMockExampleProvisioner` | `GinkgoT()` + `EXPECT().Return`           | WIRED    | Строки 32, 37; auto-Cleanup сверяет ожидания                             |
| `lefthook.yml pre-push test-inventory`         | `cd services/inventory && go test ./...` | unit only (D-15)                    | WIRED    | Присутствует; нет `-tags=integration`                                    |

---

### Post-Execution Changes (pkg split) — D-06 Single-Source Intact

После выполнения планов была произведена легитимная реорганизация: `internal/repository/mongoconn` перемещён в `pkg/mongoconn`, а generic Kafka-механика вынесена в `pkg/kafka`. Верификация D-06 по-прежнему проходит:

- **Bootstrap-сигнатура** `topology.Bootstrap(ctx context.Context, adm *kadm.Client, partitions int32) error` **сохранена** — CLI и integration-тест зовут её без изменений.
- **Имена/политики топиков НЕ дублированы**: topology.go — единственный источник доменной топологии; `pkg/kafka` содержит только механику (`EnsureTopics`), не знает доменных имён.
- **Инвариант D-14 соблюдён**: `pkg/mongoconn.Connect` — только client-фабрика + ping, без UoW/repo.
- **go.mod inventory** содержит `mongo-driver/v2 v2.7.0 // indirect` — это корректно: прямое использование перешло в `pkg/mongoconn`; inventory зависит от него через `github.com/gwall-e/pkg` (direct require + replace directive).

---

### Data-Flow Trace (Level 4)

Применимо только к артефактам, рендерящим динамические данные. Все Phase 5 артефакты — инфра-плумбинг (не data-rendering компоненты). Level 4 SKIPPED: нет компонентов с состоянием UI/data-binding.

---

### Behavioral Spot-Checks

| Поведение                                           | Команда                                                             | Результат                                                  | Статус  |
|-----------------------------------------------------|---------------------------------------------------------------------|------------------------------------------------------------|---------|
| pkg module build/vet/test (mongo-driver/v2)         | `cd pkg && go build ./... && go vet ./... && go test ./...`         | exit 0; ok pkg/http; kafka/mongoconn [no test files]       | PASS    |
| inventory unit tests (без Docker)                   | `cd services/inventory && go build ./... && go vet ./... && go test -count=1 ./...` | exit 0; ok example + topology                  | PASS    |
| integration-тест компилируется (не запускается без Docker) | `cd services/inventory && go vet -tags=integration ./internal/kafka/topology/...` | exit 0                                    | PASS    |
| integration-тест изолирован тегом                   | `go test ./...` (без тега) — integration-spec НЕ запускается       | ok (unit only)                                             | PASS    |
| audit go vet (DOC-02)                               | `cd services/audit && go vet ./...`                                 | exit 0                                                     | PASS    |
| `//go:build integration` первой строкой              | `head -1 topology_integration_test.go`                              | `//go:build integration`                                   | PASS    |
| topology.Bootstrap вызывается из cmd/main.go         | `grep 'topology.Bootstrap' cmd/main.go`                             | строка 56 — вызов подтверждён                               | PASS    |
| GOWORK=off удалён из всех канонов                   | `grep -r GOWORK=off knowledge/ lefthook.yml .planning/ROADMAP.md`  | empty output (not found)                                   | PASS    |
| inventory в go.work                                 | `grep inventory go.work`                                            | `./services/inventory`                                     | PASS    |

---

### Probe Execution

Пробов (probe-*.sh) в этой фазе нет. Step 7c: SKIPPED.

---

### Requirements Coverage

| Требование | Источник плана  | Описание                                                          | Статус    | Свидетельство                                                                    |
|------------|-----------------|-------------------------------------------------------------------|-----------|----------------------------------------------------------------------------------|
| SVC-05     | 05-01, 05-05    | Персистентность через mongo-driver/v2 (миграция с v1)             | SATISFIED | go.mod содержит v2.7.0; v1.17.9 удалён; pkg/mongoconn.Connect на v2; build зелёный |
| SVC-06     | 05-02, 05-03, 05-04 | Ginkgo v2 + Gomega + mockery; интеграционные через testcontainers | SATISFIED | unit-спеки зелёные; mockery-мок сгенерирован; integration-тест компилируется (go vet exit 0); Docker-прогон — manual/CI |
| SVC-07     | 05-01, 05-02, 05-04 | docker-compose (Kafka KRaft + Mongo RS) + bootstrap провижна топиков | SATISFIED | docker-compose.yml с идемпотентным RS healthcheck; SC2 manual-verified; Bootstrap unit-тест + integration-тест с cleanup-assert |
| DOC-02     | 05-05           | Починить build.md audit-рецепт (exit 0)                           | SATISFIED | `cd services/audit && go vet ./...` → exit 0; build.md обновлён; нет GOWORK=off |

---

### Anti-Patterns Found

Сканирование файлов, изменённых в фазе:

| Файл                                       | Строка | Паттерн | Серьёзность | Комментарий                                                       |
|--------------------------------------------|--------|---------|-------------|-------------------------------------------------------------------|
| (нет совпадений)                           | —      | TBD/FIXME/XXX | — | Ни в одном из проверенных файлов не найдено unreferenced debt-маркеров |

Единственный потенциальный «стаб» — `services/inventory/internal/example/` (throwaway пакет) — является намеренным smoke-артефактом, задокументирован как таковой (комментарии на русском, план явно называет его throwaway). Не является дефектом.

---

### Human Verification Required

#### 1. SC2: docker compose up smoke

**Test:** Из корня репозитория выполнить `make dev-up` (или `docker-compose up -d`), дождаться healthy-статуса (~30 с).

**Expected:**
1. `docker compose ps` — оба сервиса в статусе healthy (kafka + mongo)
2. `mongosh "mongodb://localhost:27018/?replicaSet=rs0&directConnection=true" --quiet --eval 'rs.status().ok'` → `1`
3. Kafka broker доступен на `localhost:9092` (нет fatal-ошибок в `docker compose logs kafka`)
4. `docker-compose down` — нормальное завершение

**Why human:** SC2 требует живого Docker engine + compose-бинаря (V1 в операторской конфигурации). Sandbox не имеет Docker daemon. Пользователь уже проводил эту проверку вручную и закоммитил verified-конфигурацию (коммит 1cb00e9), однако формальный sign-off должен быть зафиксирован.

**Note:** Mongo host-порт в операторской конфигурации — `27018` (локальная mongo занимает `27017`). Внутренний порт контейнера по-прежнему `27017`.

---

### Gaps Summary

Разрывов не обнаружено. Все пять Success Criteria кодово реализованы и верифицированы на уровне компиляции и unit-тестов. Единственный pending-элемент — формальный human sign-off по SC2 (docker compose smoke), который пользователь уже выполнял в ходе планирования, но который должен быть подтверждён явно как часть верификационного процесса.

---

### Примечания по post-execution изменениям

Изменения, внесённые после выполнения планов (pkg-split), **не нарушают** цели фазы:

1. **pkg/kafka/admin.go** — generic Kafka-механика перенесена из сервисного кода в общую библиотеку. D-06 (single source топологии) остаётся целым: topology.go по-прежнему единственный источник доменных имён/политик, механика делегируется в pkg/kafka.

2. **pkg/mongoconn/conn.go** — RS-aware Connect перенесён из `services/inventory/internal/repository/mongoconn` в `pkg/mongoconn`. D-14 (граница helper) соблюдена: только Connect+Ping, без UoW/repo. SVC-05 удовлетворён: mongo-driver/v2 используется в `pkg/mongoconn/conn.go` напрямую.

3. **Локализация log-сообщений** (коммит e4b7ddb, строки на английском) — соответствует style.md канону (комментарии кода на русском, error/log strings на английском).

---

_Verified: 2026-06-30T16:25:50Z_
_Verifier: Claude (gsd-verifier)_
