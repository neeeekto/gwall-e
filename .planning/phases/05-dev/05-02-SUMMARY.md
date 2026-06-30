---
phase: 05-dev
plan: 02
subsystem: infra
tags: [docker-compose, kafka, confluent-local, kraft, mongodb, replica-set, mockery, makefile]

# Dependency graph
requires:
  - phase: 05-dev
    provides: "стек-фундамент inventory (franz-go, mongo-driver/v2, testcontainers) — Plan 01"
provides:
  - "Ручной дев-стенд docker-compose.yml: Kafka (confluent-local KRaft, :9092) + Mongo (mongo:7 single-node RS rs0) с транзакциями"
  - "Идемпотентный rs.initiate-healthcheck (try rs.status / catch rs.initiate, host 127.0.0.1)"
  - "Makefile: MOCKERY_VERSION pin + таргеты dev-up/dev-down/topics/test-integration/generate-mocks"
  - "Клиентский URI стенда: mongodb://localhost:<host-port>/?replicaSet=rs0&directConnection=true"
affects: [topology, bootstrap-cli, integration-tests, mocks]

# Tech tracking
tech-stack:
  added: [confluentinc/confluent-local:7.5.0, mongo:7, mockery/v3@v3.7.1]
  patterns:
    - "Запиннутые конкретные image-теги (НЕ floating) для детерминизма / supply-chain (T-05-03)"
    - "Идемпотентный single-node RS healthcheck (try/catch rs.initiate) — Pitfall 1"
    - "*_VERSION pin + go install ...@$(VERSION) в tools-таргете (D-11), без tool-блока в rotten go.mod"

key-files:
  created: [docker-compose.yml]
  modified: [Makefile]

key-decisions:
  - "Дев-стенд в корне репозитория (D-05); plaintext без auth — явно dev-only (T-05-04 accepted)"
  - "mongo запускается явным `mongod` в command, иначе --replSet не доходит до демона (эмпирический фикс)"
  - "Операторская verified-конфигурация: host-порт mongo 27018 (локальная mongo на 27017), docker-compose V1"

patterns-established:
  - "Idempotent RS init: try{rs.status()}catch{rs.initiate({_id:'rs0',members:[{host:'127.0.0.1:27017'}]})}"
  - "Mockery pinning по тому же *_VERSION-паттерну, что golangci/lefthook/buf"

requirements-completed: [SVC-07, SVC-06]

# Metrics
duration: ~3 дня (с учётом двух human-verify итераций smoke-проверки)
completed: 2026-06-30
---

# Phase 05 Plan 02: Dev-стенд (docker-compose) + Makefile dev-таргеты Summary

**Ручной дев-стенд docker-compose.yml — Kafka confluent-local (KRaft, :9092) + Mongo mongo:7 single-node RS rs0 с идемпотентным rs.initiate-healthcheck — плюс Makefile с пиннутым mockery v3.7.1 и таргетами dev-up/dev-down/topics/test-integration/generate-mocks; SC2 smoke подтверждён пользователем вручную (rs.status().ok==1, Kafka :9092).**

## Performance

- **Duration:** ~3 дня (две итерации human-verify smoke; чистое время реализации — минуты)
- **Started:** 2026-06-27
- **Completed:** 2026-06-30
- **Tasks:** 3 (2 авто + 1 human-verify чекпойнт, пройден)
- **Files modified:** 2 (docker-compose.yml создан, Makefile изменён)

## Accomplishments
- Создан корневой `docker-compose.yml`: сервис `kafka` (confluentinc/confluent-local:7.5.0, KRaft, single broker, :9092) и сервис `mongo` (mongo:7, single-node replica set rs0, транзакции доступны).
- Идемпотентный RS-healthcheck (`try{rs.status()}catch{rs.initiate(...)}`, member-host `127.0.0.1`) — RS поднимается сам при первом `docker compose up`, повторные пробы не валятся.
- Расширен `Makefile`: `MOCKERY_VERSION := v3.7.1` + `go install mockery/v3` в `tools`; таргеты `dev-up`, `dev-down`, `topics`, `test-integration`, `generate-mocks`.
- SC2 (`docker compose up` поднимает стенд, транзакции доступны) подтверждён пользователем вручную: `rs.status().ok == 1`, Kafka слушает на `:9092`.

## Task Commits

1. **Task 1: docker-compose.yml — confluent-local + mongo:7 RS** - `ba47910` (feat)
2. **Task 1 fixup: явный `mongod` чтобы --replSet дошёл до демона** - `c034050` (fix, deviation Rule 1)
3. **Task 2: Makefile — MOCKERY_VERSION pin + dev-таргеты** - `2853674` (feat)
4. **Task 3 (human-verify): verified-конфигурация пользователя (порт 27018, compose V1, dev-down)** - `1cb00e9` (fix)

**Plan metadata:** см. финальный `docs(05-02)`-коммит.

## Files Created/Modified
- `docker-compose.yml` — дев-стенд: Kafka confluent-local (KRaft, :9092) + Mongo mongo:7 single-node RS rs0 (host-порт 27018→контейнер 27017) с идемпотентным rs.initiate-healthcheck.
- `Makefile` — MOCKERY_VERSION pin + go install mockery/v3 в tools; .PHONY-таргеты dev-up/dev-down/topics/test-integration/generate-mocks (docker-compose V1).

## Decisions Made
- **Явный `mongod` в command** (`["mongod","--replSet","rs0",...]`): без имени бинаря флаг `--replSet` не доходил до демона (`not running with --replSet`), RS не инициализировался.
- **Операторская verified-конфигурация закоммичена как есть** (по явному решению пользователя): host-порт mongo `27018:27017` (у пользователя локальная mongo на 27017; внутренний порт контейнера не менялся), `docker compose` → `docker-compose` (V1) + новый таргет `dev-down`.
- **Образы запиннены конкретными тегами** (confluent-local:7.5.0, mongo:7), НЕ floating — детерминизм/supply-chain (T-05-03).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] mongo не запускался с --replSet (bare command не доходил до mongod)**
- **Found during:** Task 3 (human-verify smoke, 1-я итерация)
- **Issue:** `command: ["--replSet","rs0",...]` без явного `mongod` → сервер поднимался как `not running with --replSet`, идемпотентный healthcheck не мог инициализировать RS, single-node RS (D-05) не поднимался.
- **Fix:** prepend `mongod` к command → `["mongod","--replSet","rs0","--bind_ip_all","--port","27017"]`; имя RS `rs0` совпадает с `rs.initiate` в healthcheck.
- **Files modified:** docker-compose.yml
- **Verification:** повторный smoke пользователя — `rs.status().ok == 1`.
- **Committed in:** `c034050`

**2. [Rule 1 - Consistency] Добавлен `.PHONY: dev-down`; устаревший комментарий обновлён**
- **Found during:** Task 3 (фиксация verified-конфигурации пользователя)
- **Issue:** новый пользовательский таргет `dev-down` был добавлен без `.PHONY`-объявления (все остальные таргеты файла — phony); комментарий всё ещё ссылался на «docker compose v2 plugin», хотя код перешёл на `docker-compose` V1.
- **Fix:** добавлен `.PHONY: dev-down`; комментарий приведён в соответствие коду (V1 binary, operator's verified setup).
- **Files modified:** Makefile
- **Verification:** оба verify-гейта (Task 1 + Task 2) проходят после правок.
- **Committed in:** `1cb00e9`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 consistency). Плюс одно операторское решение по конфигурации (порт 27018 + compose V1), закоммичено по явному выбору пользователя.
**Impact on plan:** Фикс №1 необходим для работоспособности стенда (SC2). Фикс №2 — консистентность Makefile. Scope без расширения.

## Issues Encountered
- 1-я итерация smoke: Kafka OK, Mongo падал `not running with --replSet` — устранено фиксом `c034050` (явный `mongod`).
- 2-я итерация smoke: пройдена (`rs.status().ok == 1`, Kafka :9092 OK).
- Замечание по verify-гейтам плана: литеральный `grep 'mockery/v3@$(MOCKERY_VERSION)'` падает из-за BRE-трактовки `$(` (анкор + скобка); содержимое файла корректно (подтверждено `grep -F`). Это квирк строки гейта, не дефект артефакта.

## User Setup Required
None — внешних сервисов конфигурировать не требуется. Дев-стенд поднимается локально через `make dev-up`. Клиентский URI: `mongodb://localhost:27018/?replicaSet=rs0&directConnection=true` (host-порт 27018 в operator-конфигурации; в дефолтной конфигурации — 27017).

## Next Phase Readiness
- Инфра-обвязка готова: `make topics` ссылается на bootstrap-CLI, который создаёт Plan 04; `make test-integration` / `make generate-mocks` готовы к коду Phase 6/7.
- Дев-стенд даёт реальный Kafka + Mongo RS для локальной разработки и ручных проверок.

## Self-Check: PASSED

- docker-compose.yml — FOUND
- Makefile — FOUND
- 05-02-SUMMARY.md — FOUND
- Commits ba47910, c034050, 2853674, 1cb00e9 — all FOUND

---
*Phase: 05-dev*
*Completed: 2026-06-30*
