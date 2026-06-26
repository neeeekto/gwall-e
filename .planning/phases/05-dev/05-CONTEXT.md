# Phase 5: Dev-инфра и стек - Context

**Gathered:** 2026-06-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Готовое локальное окружение + обновлённый стек, на котором можно писать репозитории и
гонять интеграционные тесты — **до** первого доменного кода. Нулевой шаг v3.0, продолжает
v1.0-фундамент.

**В скоупе (Success Criteria из ROADMAP):**
1. `services/inventory` собирается с mongo-driver/**v2** (v1 удалён из `go.mod`), `go build`/`go vet` зелёные
2. `docker compose up` поднимает Kafka (KRaft) + MongoDB single-node replica set (транзакции доступны)
3. Bootstrap провижнит топики `inventory.*.events` (`cleanup=delete`) и `inventory.*.state` (`cleanup=compact`)
4. Интеграционный тест на testcontainers (Kafka KRaft + Mongo RS) стартует и подключается; Ginkgo v2 + Gomega + mockery — smoke-прогон
5. DOC-02: `build.md` audit-рецепт реально проходит (exit 0)

**НЕ в скоупе (другие фазы):** доменная модель/агрегаты (Phase 6), UnitOfWork/Outbox/repositories/gRPC (Phase 7), protobuf-схемы событий + relay→Kafka (Phase 8), топология connections (Phase 9), test-consumer/верификация backbone (Phase 10). Schema registry, консьюмеры, inbox/dedup — вне v3.0.

</domain>

<decisions>
## Implementation Decisions

### Граница inventory / go.work
- **D-01:** inventory становится **полноправным членом `go.work`** (остаётся в `use`-блоке). Канон `GOWORK=off` **отменяется**; сборка/вет — общий `go build ./...` / `go vet ./...` из корня workspace.
- **D-02:** pre-push **больше не исключает** inventory (снять намеренное исключение из `lefthook.yml`).
- **D-03:** Осознанная цена решения D-01/D-02: WIP-ошибка компиляции в inventory ломает сборку **всего** workspace → действует инвариант «inventory всегда компилируется на каждом коммите» (уместно, раз он теперь эталонный сервис v3.0).
- **D-04:** Переписать под новое решение каноны: `knowledge/build.md` (раздел `GOWORK=off` + строка про pre-push), `knowledge/structure.md` (раздел inventory вне workspace), `knowledge/boundaries.md`, **и формулировку Success Criterion 1 в ROADMAP** (убрать «с `GOWORK=off`»). Это in-scope Phase 5, не deferred.

### Локальный стенд (compose ↔ testcontainers)
- **D-05:** Два целевых пути: **docker-compose** = ручной дев-стенд, **testcontainers** = изолированные эфемерные интеграционные тесты (идиоматично для Go). Не объединяем тесты через compose-модуль.
- **D-06:** Дрейфующее знание (имена + cleanup-политики топиков, число партиций, версии образов) живёт в **одном Go-модуле**: общая bootstrap-функция на `kadm` + константы. И compose-провижн, и тесты зовут **эту же** функцию — единственный источник истины для топологии топиков.
- **D-07:** Kafka-образ — `confluentinc/confluent-local` (паритет дев≈тест: testcontainers kafka-модуль по умолчанию `confluentinc/confluent-local:7.5.0`), KRaft из коробки, один брокер.
- **D-08:** Mongo — `mongo:7` как **single-node replica set** (`--replSet` + `rs.initiate()`); `mongo-driver/v2` v2.7 требует MongoDB ≥ 4.4, транзакции требуют RS.

### Provisioning топиков
- **D-09:** Запуск провижна — тонкий **Go CLI в `cmd/`**, который зовёт общую bootstrap-функцию (D-06); дёргается `make`-таргетом (`make topics` / `make dev-up`) после подъёма compose; тесты зовут ту же функцию напрямую.
- **D-10:** В Phase 5 заводим **только** `inventory.host.events` + `inventory.host.state`. Список агрегатов — **data-driven** (константа/конфиг): добавление `project`/`module`/`location` в Phase 6+ = одна строка, без преждевременной фиксации имён до проектирования домена.
- **D-11:** Число партиций — **параметр** bootstrap'а; дев/тест-дефолт малый (**6** — заодно гоняет sticky-key partitioner на >1 партиции, проверяя порядок per-entity). Prod-число под целевой ~150k парк — **отложено** (ops-решение, фиксируется при первом консьюмере; не блокирует инфру).
- **D-12:** Политики топиков (залочено ресёрчем/каноном, не пересматривается): `*.events` → `cleanup.policy=delete` (длинный retention, immutable-история фактов); `*.state` → `cleanup.policy=compact` + `delete.retention.ms ≥ 24h`; Kafka message key = внутренний `ID` (НЕ FQDN/INV/MAC).

### Миграция mongo-driver/v2 + интеграционные тесты
- **D-13:** `internal/` пуст, кода на v1 нет → «миграция» = свап `go.mod` (`mongo-driver v1.17.9` → `mongo-driver/v2 v2.7.0`, удалить v1) + доказать v2 рабочим подключением к Mongo RS интеграционным тестом.
- **D-14:** Глубина скаффолда — **тонкий connection-helper**: Mongo client-фабрика (RS-aware, `writeconcern.Majority`) + health-ping. Эталон «как подключаемся» для Phase 6/7. **Без** repository/UoW/Outbox — это Phase 7 (не залезаем).
- **D-15:** Интеграционные тесты изолированы **build-tag `integration`** + `make test-integration`. `go test ./...` и pre-push гоняют **только unit** (быстро, без контейнеров); testcontainers-тесты — за тегом / в CI. Критично, т.к. после D-02 pre-push включает inventory.

### Claude's Discretion
- **DOC-02 (SC5):** заменить audit-рецепт в `knowledge/build.md` (сейчас `cd services/audit && go build ./...` падает: `build output "cmd" already exists and is a directory`) на команду с exit 0. Кандидаты: `go vet ./...` (предпочтительно — без бинаря, валидирует) / `go build ./cmd` / `go build -o /dev/null ./...`. **Финальную форму executor верифицирует эмпирически** (SC5 требует реального exit 0); привести рецепты к новому workspace-канону (D-01).
- **mockery (SVC-06):** провод `.mockery.yaml` + `make generate-mocks`, доказать smoke одним throwaway/example-интерфейсом (реальных доменных портов ещё нет — появятся в Phase 6/7).
- Конкретная структура Makefile-таргетов, healthcheck'и compose, имена переменных/констант, лэйаут пакетов хелпера и теста — на усмотрение планнера/executor'а в рамках решений выше.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Стек и ресёрч (версии, обоснования, pitfalls)
- `.planning/research/STACK.md` — точные версии и обоснование: franz-go `v1.21.4` (`pkg/kgo`+`pkg/kadm`), mongo-driver/v2 `v2.7.0`, testcontainers-go `v0.43.0` (+ `modules/kafka` KRaft, `modules/mongodb` RS), команды `go get`; что НЕ использовать (schema registry, confluent-kafka-go/CGO, segmentio, EOS); MongoDB v2 специфика UoW (callback на `context.Context`, RS обязателен).
- `.planning/research/SUMMARY.md` — executive summary v3.0, Phase 1 (=наша Phase 5) rationale, dual-topic, gaps (число партиций, `delete.retention.ms`).
- `.planning/research/PITFALLS.md` — 9 pitfalls; для Phase 5 актуальны pitfall #3 (Kafka key = ID, не FQDN/INV/MAC) и dual-topic (история vs compacted-снапшот физически несовместимы в одном топике).

### Каноны репозитория (обновляются в этой фазе — D-04)
- `knowledge/build.md` — рецепты сборки/тестов; разделы про `GOWORK=off` (§ inventory, строки ~68-80), pre-push исключение (~44-45), audit-рецепт DOC-02 (~61). **МЕНЯЮТСЯ** под D-01/D-02/DOC-02.
- `knowledge/structure.md` — членство go.work, WIP-статус inventory (§ ~36). **МЕНЯЕТСЯ** под D-01.
- `knowledge/boundaries.md` — границы модулей/исключений. Свериться/обновить под D-01/D-02.
- `knowledge/testing.md` — Ginkgo v2 + Gomega + mockery конвенции (англ. комментарии в тестах). Источник для smoke-структуры (SVC-06).
- `knowledge/architecture.md` — канон слоёв `domain/usecases/query/repositories/api/cron` + `app`/`cmd`, UoW/outbox/relay (контекст для connection-helper, чтобы не залезть в Phase 7).
- `Makefile` — пиннинг версий инструментов (golangci-lint v2, lefthook, commitlint, buf); сюда добавляются новые dev-таргеты (D-09/D-15).
- `lefthook.yml` — pre-push hook (снять исключение inventory — D-02).
- `go.work` — членство workspace (inventory остаётся — D-01).
- `services/inventory/go.mod` — текущий `mongo-driver v1.17.9` → свап на `/v2` (D-13).

### Требования и роадмап
- `.planning/REQUIREMENTS.md` — SVC-05 (mongo-driver/v2), SVC-06 (Ginkgo/Gomega/mockery + testcontainers), SVC-07 (docker-compose + bootstrap топиков), DOC-02.
- `.planning/ROADMAP.md` § Phase 5 — Goal + 5 Success Criteria (SC1 формулировка корректируется — D-04).
- `.planning/seeds/SEED-002-audit-logging.md`, `SEED-003-authorization-on-all-actions.md` — forward-compat контекст (почему envelope `actor`/`eventId` важны позже; не реализуется в Phase 5, но определяет, почему топики/ключи делаем стабильными сейчас).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `services/inventory/go.mod`: уже есть `github.com/google/uuid v1.6.0` (генерация внутреннего `ID`); mongo-driver v1.17.9 — подлежит замене на v2 (D-13).
- `services/inventory/cmd/main.go`: стаб `func main(){ return }` — точка, куда ляжет CLI/composition (новый bootstrap-CLI — D-09).
- `buf.yaml` / `buf.gen.yaml`: buf-скелет на месте — НЕ активируется в Phase 5 (proto событий = Phase 8); упоминается, чтобы не путать.
- `Makefile`: уже пиннит версии инструментов — паттерн для добавления dev/test-таргетов и пиннинга версий образов.

### Established Patterns
- `services/inventory/internal/` существует как **пустой** набор директорий (`domain/usecase/query/repository/api/cron/app`) — `internal/` снесён в v1.0 (чистый лист). Connection-helper (D-14) ляжет в канон-слой `repositories` (или подходящий infra-пакет), но БЕЗ repository/UoW-реализаций.
- Тесты: Ginkgo v2 + Gomega (в `pkg/go.mod` уже v2.23.4/v1.38.0); комментарии в тестах — на английском, доменные комментарии в коде — на русском (`knowledge/style.md`).
- inventory исторически собирался `GOWORK=off` — этот паттерн **отменяется** (D-01); все рецепты приводятся к workspace-сборке.

### Integration Points
- `go.work` ↔ pre-push (`lefthook.yml`) ↔ build.md — связка, которую правит D-01/D-02/D-04 согласованно.
- bootstrap-функция топиков ↔ {docker-compose провижн, testcontainers-тесты} — единая точка (D-06).
- connection-helper ↔ будущие repositories/UoW (Phase 7) — helper задаёт сигнатуру подключения, не реализацию транзакций.

</code_context>

<specifics>
## Specific Ideas

- Дев-дефолт партиций = **6** выбран не «потому что число», а чтобы стенд реально прогонял sticky-key partitioner на >1 партиции (валидация порядка per-entity), оставаясь дешёвым на single-broker.
- Паритет образов дев≈тест (`confluentinc/confluent-local`) — явная цель: не допустить «в тесте зелено, в `docker compose` красно».
- DOC-02: предпочтение к `go vet ./...` как канонической проверке (без артефакта-бинаря), но окончательно — по факту exit 0.

</specifics>

<deferred>
## Deferred Ideas

- **Prod-число партиций Kafka** (под ~150k парк) — ops-решение; фиксируется при появлении первого консьюмера (Search/Analytics/Audit, отдельный milestone). В Phase 5 — только configurable-параметр с дев-дефолтом.
- **Schema Registry** — вводится с первым консьюмером (режим BACKWARD_TRANSITIVE для protobuf), SEED-002. Вне v3.0.
- **`kprom` / Prometheus-метрики продюсера** — дешёвый плагин franz-go, но наблюдаемость relay откладываем до первой эксплуатации (Phase 8+).
- **Redpanda-модуль testcontainers** — fallback, только если тестам понадобится SASL/SCRAM (у Kafka-модуля известны проблемы с SASL). Для plaintext-дева не нужен.
- **CI** (полноценный pipeline) — за рамками Phase 5; сейчас enforcement через локальный lefthook. Решения D-01/D-02/D-15 заложены так, чтобы будущий CI лёг без переделки (workspace-build + тег-изоляция интеграционных тестов).

</deferred>

---

*Phase: 5-Dev-инфра и стек*
*Context gathered: 2026-06-27*
