# Phase 5: Dev-инфра и стек — Pattern Map

**Mapped:** 2026-06-27
**Files analyzed:** 13 (NEW + MODIFIED)
**Analogs found:** 7 структурных/конфиг-аналога / 13 файлов

> **Clean-slate-фаза.** `services/inventory/internal/` — пустой набор канон-директорий
> (`domain/usecase/query/repository/api/cron/app`), v1.0-код снесён. Большинство Go-файлов
> Phase 5 — **NEW без in-repo доменного аналога**; для них указан ближайший
> **структурный/конфиг**-аналог (паттерн Makefile-пиннинга, структура lefthook-хука,
> членство go.work, Ginkgo suite-каркас), а API-сигнатуры берутся verbatim из
> `05-RESEARCH.md` (§ Code Examples / § Standard Stack). Отсутствие аналога — само по
> себе планировочный сигнал, аналоги НЕ выдуманы.
>
> **Расхождение state vs CONTEXT (важно планнеру):** на момент маппинга
> `go.work` **уже** содержит `./services/inventory` (D-01 частично применён в state),
> а `lefthook.yml` **всё ещё** несёт исключение inventory (`GOWORK=off` lint + отсутствие
> pre-push команды). То есть D-01 для go.work — фактически done, а D-02/D-04 — нет.
> Каноны `build.md`/`structure.md`/`boundaries.md` всё ещё описывают inventory как
> «вне workspace, WIP» — это противоречит реальному go.work и подлежит правке (D-04).

## File Classification

| New/Modified файл | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `services/inventory/go.mod` (MODIFY) | config (deps) | — | `services/inventory/go.mod` (сам себя — свап require-блока) | exact (in-place) |
| `go.work` (MODIFY/verify) | config (workspace) | — | `go.work` (сам себя — inventory уже в `use`) | exact (no-op vs D-01) |
| `lefthook.yml` (MODIFY) | config (git-hooks) | event-driven | `lefthook.yml` § pre-push (его же `test-audit`/`test-pkg` команды) | exact (in-place) |
| `Makefile` (MODIFY) | config (tooling) | — | `Makefile` § `*_VERSION` + `tools` target | exact (паттерн пиннинга) |
| `services/inventory/internal/kafka/topology/*.go` (NEW) | utility (topology const + bootstrap) | batch / admin | НЕТ доменного аналога — структурно: `05-RESEARCH.md` § kadm bootstrap | no-analog (config-driven) |
| `services/inventory/cmd/` bootstrap CLI (NEW/replace stub) | cmd (composition/entrypoint) | request-response (one-shot) | `services/inventory/cmd/main.go` (stub `func main(){return}`) | role-match (stub only) |
| Mongo connection-helper (NEW, в `repository`/infra-слое) | repository-adjacent (connection factory) | request-response (connect+ping) | НЕТ — структурно: `05-RESEARCH.md` § Mongo helper v2 | no-analog (clean slate) |
| `*_integration_test.go` (`//go:build integration`) (NEW) | test (integration) | event-driven (containers) | `pkg/http/http_test.go` (suite-каркас) | role-match (suite only) |
| unit-spec с моком (NEW) | test (unit) | — | `pkg/http/middlewares_test.go` (Describe/Context/It + Gomega) | role-match |
| `.mockery.yaml` (NEW) | config (codegen) | — | НЕТ — структурно: `05-RESEARCH.md` § mockery v3 + `knowledge/testing.md` § мокинг | no-analog |
| throwaway example-интерфейс для mockery (NEW) | model (port/interface) | — | НЕТ доменных портов — иллюстрация в `knowledge/testing.md` (`OrderRepository`) | no-analog |
| `docker-compose.yml` (NEW) | config (infra) | — | НЕТ (нет авторитетного compose) — `05-RESEARCH.md` § docker-compose | no-analog |
| `knowledge/build.md` / `structure.md` / `boundaries.md` + `.planning/ROADMAP.md` (MODIFY) | docs (canon) | — | сами себя (правка существующих разделов про `GOWORK=off`) | exact (in-place) |

## Pattern Assignments

### `services/inventory/go.mod` (config, MODIFY — D-13/SVC-05)

**Analog:** сам файл — точечный свап require-блока.

**Текущее состояние** (`services/inventory/go.mod` строки 5-8) — то, что меняется:
```go
require (
	github.com/google/uuid v1.6.0
	go.mongodb.org/mongo-driver v1.17.9
)
```
**Целевой паттерн** (RESEARCH § Standard Stack / Installation):
- удалить `go.mongodb.org/mongo-driver v1.17.9`;
- добавить `go.mongodb.org/mongo-driver/v2 v2.7.0`, `github.com/twmb/franz-go v1.21.4`,
  `github.com/twmb/franz-go/pkg/kadm v1.18.0` (+ test-deps testcontainers v0.43.0 — попадут
  при первом импорте), `go mod tidy`;
- `github.com/google/uuid v1.6.0` — **сохранить** (генерация внутреннего `ID`).
- indirect v1-deps (`xdg-go/*`, `montanaflynn/stats`, `golang/snappy`, `youmark/pkcs8`,
  строки 11-21) — уйдут после `go mod tidy` (импортов v1 нет, `internal/` пуст → правок кода нет).

**Инвариант (D-03/Pitfall 3):** inventory теперь в go.work → свап делать **атомарно**
(go.mod + tidy + `go build ./...` зелёный в одном шаге), иначе ломается весь workspace-build.

---

### `go.work` (config, MODIFY/verify — D-01)

**Analog:** сам файл.

**Текущее состояние** (`go.work` целиком) — D-01 **уже отражён**:
```go
go 1.24.6

use (
	./pkg
	./services/analytics
	./services/audit
	./services/inventory
)
```
**Действие планнера:** verify-only (inventory уже в `use`-блоке). Реальная работа D-01 —
не в go.work (он готов), а в **отмене `GOWORK=off`** в lefthook.yml + каноны build.md/
structure.md/boundaries.md, которые всё ещё описывают inventory как «вне workspace».

---

### `lefthook.yml` (config, event-driven, MODIFY — D-02)

**Analog:** сам файл — снять исключение inventory, добавить inventory unit-команду.

**Pre-commit `lint-inventory`** (строки 33-39) — содержит `GOWORK=off`, подлежит изменению:
```yaml
    lint-inventory:
      glob: "services/inventory/**/*.go"
      run: cd services/inventory && GOWORK=off golangci-lint run ./...
      stage_fixed: true
```
→ под D-01 либо слить inventory в `lint-workspace` цикл (строки 24-32), либо убрать `GOWORK=off`.

**Pre-push** (строки 49-57) — паттерн для копирования новой inventory-команды:
```yaml
pre-push:
  parallel: false
  commands:
    test-pkg:
      run: cd pkg && go test ./...
    test-audit:
      run: cd services/audit && go test ./...
    test-analytics:
      run: cd services/analytics && go test ./...   # 0 packages → no-op, NOT an error
```
**Что добавить (D-02/D-15):** `test-inventory: run: cd services/inventory && go test ./...`
— **БЕЗ** `-tags=integration` (pre-push гоняет только unit, без Docker; integration за тегом).
Комментарий-блок строк 44-47 («inventory INTENTIONALLY NOT tested») — **удалить** (он
обосновывал отменённое исключение). Источник истины «без `go test ./cmd`-коллизии» — см.
DOC-02 ниже (тот же pitfall package-main).

---

### `Makefile` (config, MODIFY — D-09/D-15/SVC-06)

**Analog:** сам файл — паттерн `*_VERSION` пиннинга + target.

**Паттерн пиннинга версий** (строки 8-16):
```makefile
GOLANGCI_VERSION := v2.12.2
LEFTHOOK_VERSION := v2.1.9
BUF_VERSION      := v1.71.0

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
```
**Что добавить (по этому же паттерну):**
- `MOCKERY_VERSION := v3.7.1` + `go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)`
  (в `tools` или отдельный target). **НЕ** добавлять `tool`-блок в корневой go.mod — он
  rotten (комментарий строк 5-6, go 1.23.6 вне go.work).
- `.PHONY`-таргеты (точные имена — на усмотрение планнера в рамках D-09/D-15):
  `dev-up` (`docker compose up -d` — v2-плагин, см. RESEARCH § Open Question 3),
  `topics` (запуск CLI провижна),
  `test-integration` (`go test -tags=integration ./...` из inventory),
  `generate-mocks` (`mockery`).
- **Стиль (важно):** комментарии в этом файле — **на английском** (как существующие
  строки 1-6), в отличие от доменного кода. Соблюсти.

---

### `services/inventory/internal/kafka/topology/*.go` (utility, batch/admin, NEW — D-06/D-10/D-11/D-12, SVC-07)

**Analog:** НЕТ in-repo аналога (нет kafka-кода). Структурный источник — `05-RESEARCH.md`
§ Code Examples / kadm bootstrap (строки 405-438) с verbatim-сигнатурой
`kadm.CreateTopics(ctx, partitions int32, rf int16, configs map[string]*string, topics...)`.

**Ключевые инварианты для копирования из RESEARCH (НЕ переоткрывать):**
- `var aggregates = []string{"host"}` — data-driven (D-10); добавление агрегата = одна строка.
- `*.events` → `cleanup.policy=delete`; `*.state` → `cleanup.policy=compact` +
  `delete.retention.ms=86400000` (≥24h) (D-12).
- партиции — **параметр** функции, дев-дефолт `6` (D-11); rf `int16(1)` (single broker, D-07).
- ошибки — `fmt.Errorf("... %w", err)` (sentinel+wrap, `errorlint`-hook).
- **Эта функция — single source** (D-06): её зовут И CLI, И integration-тест. Дублировать
  конфиг топиков где-либо ещё — anti-pattern.
- **Язык:** доменные комментарии в коде — **на русском** (`knowledge/style.md`);
  идентификаторы — английские; типизированные ID если фигурируют.

---

### `services/inventory/cmd/` bootstrap CLI (cmd, NEW/replace — D-09, SVC-07)

**Analog:** `services/inventory/cmd/main.go` — текущий стаб (целиком):
```go
package main

func main() {
	return
}
```
**Целевой паттерн:** тонкий CLI = parse env/flags (`KAFKA_BROKERS`, partitions) →
`kgo.NewClient(kgo.SeedBrokers(...))` → `kadm.NewClient(cl)` → `topology.Bootstrap(ctx, adm, parts)`.
**НЕ** дублировать топологию (зовёт функцию из topology-пакета). Точное размещение
(`cmd/topics/` подпакет vs флаг в `cmd/main.go`) — на усмотрение планнера (RESEARCH § Recommended
Structure отмечает оба варианта). **Pitfall (DOC-02/Pitfall 2):** если в `cmd/` лежит
`package main`, `go build ./...` упадёт `build output "cmd" already exists` — рецепты сборки
использовать `go vet ./...` / `go build -o /dev/null ./...`, **не** `go build ./cmd`.

---

### Mongo connection-helper (repository-adjacent, request-response, NEW — D-14, SVC-05)

**Analog:** НЕТ in-repo аналога (нет mongo-кода; `internal/repository/` пуст). Структурный
источник — `05-RESEARCH.md` § Code Examples / Mongo connection-helper v2 (строки 440-467).

**Ключевые инварианты (verbatim из RESEARCH, не переоткрывать):**
- v2-сигнатура: `mongo.Connect(opts ...*options.ClientOptions)` — **БЕЗ `ctx`** (отличие от v1).
- `options.Client().ApplyURI(uri).SetWriteConcern(writeconcern.Majority())`.
- health-`Ping(ctx, nil)`; при ошибке — `Disconnect(ctx)` + wrap-ошибка.
- URI локального RS: `...?replicaSet=rs0&directConnection=true` (Pitfall 1 — обход discovery).
- **Граница (D-14):** только client-фабрика + ping. **БЕЗ** repository/UoW/Outbox/транзакций —
  это Phase 7 (architecture.md: MUST NOT возрождать `TxManager`/`tx.go`).
- Размещение — `repositories`/infra-пакет (architecture.md), точный путь — планнер.

---

### `*_integration_test.go` (`//go:build integration`) (test, event-driven, NEW — D-15, SVC-06)

**Analog (suite-каркас):** `pkg/http/http_test.go` (целиком — реальный, компилируется):
```go
package http

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIdmServicesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Package Suite")
}
```
**Целевой паттерн (RESEARCH § testcontainers smoke, строки 469-529):**
- первая строка файла — `//go:build integration` (D-15; `go test ./...` и pre-push НЕ
  компилируют → нет Docker в pre-push, критично после D-02).
- тот же suite-каркас (`RegisterFailHandler(Fail)` + `RunSpecs`), dot-imports ginkgo/gomega.
- `kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID(...))` →
  `kc.Brokers(ctx)` → `kgo.NewClient(kgo.SeedBrokers(...))` → `cl.Ping(ctx)` (+ `topology.Bootstrap`).
- `mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))` → `mc.ConnectionString(ctx)` →
  `Connect(ctx, uri)` (helper выше). **Доверять** `ConnectionString` — не хардкодить хост (Pitfall 1).
- `DeferCleanup(func(){ _ = testcontainers.TerminateContainer(c) })`.
- комментарии в тесте — **на английском** (testing.md § язык).

---

### unit-spec с моком (test, NEW — SVC-06)

**Analog:** `pkg/http/middlewares_test.go` — структура спека (см. `knowledge/testing.md`
строки 67-100, реальный компилируемый пример). Паттерн `Describe` → `Context` → `BeforeEach`
→ `It` + Gomega-ассерты (`Expect(...).To(...)`, `ToNot(HaveOccurred())`, `MatchError`).

**Целевой паттерн мокинга** (`knowledge/testing.md` строки 127-143):
- `repo := NewMockExampleProvisioner(GinkgoT())` (НЕ `t`; авто-Cleanup сверит ожидания).
- `repo.EXPECT().Method(mock.Anything).Return(nil)`.
- `import "github.com/stretchr/testify/mock"` — **обычный** qualified-импорт (НЕ dot, в
  отличие от ginkgo/gomega).
- ассерт результата — через Gomega. **Этот spec — БЕЗ** build-tag (unit; гоняется pre-push'ем).

---

### `.mockery.yaml` (config, NEW — SVC-06)

**Analog:** НЕТ in-repo аналога. Источник — `05-RESEARCH.md` § mockery v3 (строки 555-570).

**v3-инварианты (Pitfall 5 — не путать с v2):**
- `template: testify`, `formatter: goimports`, `all: false`.
- `.SrcPackageName`/`.SrcPackagePath` (НЕ `.PackageName`/`.PackagePath` — v2-ключи).
- моки рядом с кодом / `{{.InterfaceDir}}/mocks`, один файл на пакет.
- `packages:` указывает на throwaway example-пакет с `ExampleProvisioner` (реальных доменных
  портов нет — Phase 6/7). Smoke: `make generate-mocks` генерит мок → unit-spec выше зелёный.

---

### `throwaway example-интерфейс` (model/port, NEW — SVC-06)

**Analog:** доменных портов нет. Иллюстративный шаблон — `knowledge/testing.md`
(`OrderRepository.Save`, строки 130-137). Завести **один** example-интерфейс (напр.
`ExampleProvisioner`) только чтобы mockery доказал smoke; пометить как throwaway/пример
(не доменный порт). Типизированные ID / sentinel-ошибки если фигурируют (style.md).

---

### `docker-compose.yml` (config, NEW — D-05/D-07/D-08, SVC-07)

**Analog:** НЕТ авторитетного compose в репозитории (`docker-compose.yml` отсутствует;
`boundaries.md` § «Стале-файлы» упоминает гипотетический stale-compose — его НЕТ физически).
Источник — `05-RESEARCH.md` § docker-compose (строки 531-553).

**Целевой паттерн:**
- `kafka: confluentinc/confluent-local:7.5.0` (KRaft, 1 broker, `:9092`) — паритет с
  testcontainers-модулем (D-07). Точные `KAFKA_ADVERTISED_LISTENERS` — **executor верифицирует
  эмпирически** при `docker compose up` (RESEARCH A1/A2, SC2 требует реального подъёма).
- `mongo: mongo:7`, `command: --replSet rs0 --bind_ip_all`, healthcheck идемпотентный
  `try{rs.status()}catch{rs.initiate({...host:'127.0.0.1:27017'})}` (Pitfall 1).
- клиент к стенду: `mongodb://localhost:27017/?replicaSet=rs0&directConnection=true`.
- **Размещение** (корень vs `infra/`/`deploy/`) — планнер (RESEARCH Open Question 2). Если
  кладётся в корень — отметить как **актуальный** (убрать из «неавторитетных» в boundaries.md).
- образы — **конкретные теги** (НЕ `:latest`) — security § V14 / D-07/D-08.

---

### Canon docs: `build.md` / `structure.md` / `boundaries.md` / `ROADMAP.md` (docs, MODIFY — D-04)

**Analog:** сами файлы — правка разделов про отменённый `GOWORK=off`.

| Файл | Раздел/строки | Что меняется (D-04) |
|------|---------------|---------------------|
| `knowledge/build.md` | § «inventory — WIP, `GOWORK=off`» (стр. 68-81) | Удалить/переписать: inventory теперь член workspace, собирается общим `go build ./...` без `GOWORK=off`. |
| `knowledge/build.md` | § pre-push исключение (стр. 44-46) | Убрать «inventory исключён намеренно»; теперь pre-push включает inventory unit. |
| `knowledge/build.md` | § audit-рецепт (стр. 61) | DOC-02: `cd services/audit && go build ./...` → `go vet ./...` (exit 0, verified). |
| `knowledge/structure.md` | § «inventory — вне workspace, WIP» (стр. 36-48) | Переписать: inventory — полноправный член go.work (D-01); снять WIP-«вне workspace». |
| `knowledge/structure.md` | список workspace (стр. 19-24) | «ровно три» → четыре (+ inventory). |
| `knowledge/boundaries.md` | карта владения / pre-push исключение (стр. 72) | Свериться/обновить: убрать факт «исключение inventory из pre-push». |
| `knowledge/boundaries.md` | § «Не трогать WIP-леса» (стр. 18-21) | inventory как пример WIP-лесов — пересмотреть (теперь эталонный, компилируемый). |
| `.planning/ROADMAP.md` | § Phase 5 SC1 | Убрать формулировку «с `GOWORK=off`» (D-04). |

**Стиль каноны:** MUST/SHOULD/WON'T + парность «запрет→do» + pointer-over-copy
(`knowledge/authoring.md`); один факт = один канон — не дублировать команды между build.md
и structure.md.

## Shared Patterns

### Tooling-пиннинг версий
**Source:** `Makefile` строки 8-16 (`*_VERSION` + `go install ...@$(VERSION)`).
**Apply to:** mockery (`MOCKERY_VERSION := v3.7.1`). НЕ в корневой go.mod `tool`-блок (rotten).
```makefile
GOLANGCI_VERSION := v2.12.2
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
```

### Ginkgo suite-каркас
**Source:** `pkg/http/http_test.go` (целиком).
**Apply to:** integration-suite И любой новый unit-suite в inventory.
```go
func TestXxxSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "<Suite Name>")
}
```
dot-imports `. "github.com/onsi/ginkgo/v2"` / `. "github.com/onsi/gomega"`; `*_test.go` рядом с кодом.

### Build-tag изоляция интеграционных тестов
**Source:** `05-RESEARCH.md` § Pattern 2 (строки 306-311) + D-15.
**Apply to:** все файлы, поднимающие testcontainers.
```go
//go:build integration
```
`go test ./...` (+ pre-push) НЕ компилируют → unit-прогон без Docker. `make test-integration` =
`go test -tags=integration ./...`. **Критично после D-02** (pre-push теперь включает inventory).

### Sentinel-ошибки + `%w`-wrap
**Source:** `knowledge/style.md` (errorlint — hook).
**Apply to:** topology bootstrap, connection-helper, CLI.
```go
return fmt.Errorf("создать топик %s.events: %w", agg, err)
```

### Язык кода vs тестов
**Source:** `knowledge/style.md` / `knowledge/testing.md` § язык.
**Apply to:** весь новый Go-код.
- Доменные комментарии в **коде** (topology, helper, CLI) — **на русском**; идентификаторы — английские.
- Комментарии **в тестах** — **на английском**.
- Makefile-комментарии — **на английском** (по образцу существующих строк 1-6).

### Single-source топология (anti-дубль)
**Source:** D-06 + `05-RESEARCH.md` § Pattern 1.
**Apply to:** CLI + integration-test — оба зовут `topology.Bootstrap(...)`; конфиг топиков
(имена/политики/партиции) живёт **только** в topology-пакете. Дублировать в CLI/тесте — anti-pattern.

### DOC-02 / package-main build-коллизия
**Source:** `05-RESEARCH.md` § DOC-02 (строки 666-680, эмпирически verified) / Pitfall 2.
**Apply to:** все build-рецепты (build.md, Makefile-таргеты, lefthook).
- Валидация сборки — `go vet ./...` (exit 0, без артефакта) или `go build -o /dev/null ./...`.
- **НЕ** `go build ./cmd` / `go build ./...` для модуля с `package main` в `cmd/` (падает
  `build output "cmd" already exists`).

## No Analog Found

Файлы без in-repo аналога (планнер берёт паттерн из `05-RESEARCH.md` § Code Examples):

| Файл | Role | Data Flow | Reason | RESEARCH-источник |
|------|------|-----------|--------|-------------------|
| `internal/kafka/topology/*.go` | utility | batch/admin | Kafka-кода в репо нет (clean slate) | § kadm bootstrap (стр. 405-438) |
| Mongo connection-helper | repository-adjacent | request-response | mongo-кода нет; `internal/repository/` пуст | § Mongo helper v2 (стр. 440-467) |
| `.mockery.yaml` | config | — | mockery в репо отсутствует (testing.md: «сейчас нет») | § mockery v3 (стр. 555-570) |
| throwaway example-интерфейс | model/port | — | доменных портов ещё нет (Phase 6/7) | testing.md illustration (стр. 130-137) |
| `docker-compose.yml` | config | — | авторитетного compose нет физически | § docker-compose (стр. 531-553) |

## Metadata

**Analog search scope:** `Makefile`, `lefthook.yml`, `go.work`, `services/inventory/{go.mod,cmd,internal/*}`,
`pkg/**` (go + test файлы), `knowledge/{build,structure,boundaries,testing}.md`, корневой compose/mockery (отсутствуют).
**Files scanned:** ~18 (5 конфигов + 1 stub + 5 pkg-Go + 4 canon-doc + 2 verify-отсутствия + go.mod-ы).
**Pattern extraction date:** 2026-06-27
