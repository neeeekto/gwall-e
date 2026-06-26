# Phase 5: Dev-инфра и стек - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-27
**Phase:** 5-Dev-инфра и стек
**Areas discussed:** go.work и граница inventory, Стенд (compose ↔ testcontainers), Provisioning топиков, Миграция v2 + где тесты

---

## go.work и граница inventory

### Q1 — членство inventory в go.work

| Option | Description | Selected |
|--------|-------------|----------|
| Вынести из go.work (реком.) | GOWORK=off, совпадает с текущим каноном и SC1 | |
| Оставить в go.work | Пересмотреть канон, inventory в workspace | ✓ |
| Обсудить почему так | Разобрать причину/последствия перед решением | |

**User's choice:** Оставить в go.work.
**Notes:** Скаут нашёл расхождение: go.work уже включал inventory, тогда как build.md/structure.md и SC1 требовали GOWORK=off. Пользователь выбрал привести канон к реальности (workspace), а не наоборот.

### Q2 — как трактуем «inventory в go.work»

| Option | Description | Selected |
|--------|-------------|----------|
| Полноправный член workspace (реком. после выбора Q1) | Общий go build ./..., GOWORK=off-канон убираем, pre-push включает inventory | ✓ |
| В go.work, но WIP-изоляция | go.work ради IDE, но build/test/pre-push остаются GOWORK=off | |
| Обсудить ещё | Глубже trade-off | |

**User's choice:** Полноправный член workspace.
**Notes:** Осознанная цена — WIP-ошибка ломает весь workspace-build → инвариант «inventory всегда компилируется». Уместно, раз inventory становится эталонным сервисом v3.0. Требует переписать build.md/structure.md/boundaries.md + SC1.

---

## Стенд (compose ↔ testcontainers)

### Q3 — стратегия источника истины

| Option | Description | Selected |
|--------|-------------|----------|
| Два пути + общий провижн в Go (реком.) | compose=дев, TC=тесты; общая bootstrap-функция на kadm + константы | ✓ |
| Compose = источник, тесты через compose | testcontainers поднимает тот же compose-файл | |
| Два независимых пути | Полностью раздельно, версии в .env/Makefile, логика провижна дублируется | |

**User's choice:** Два пути + общий провижн в Go.
**Notes:** Дрейфующее знание (топики/политики/партиции/версии) централизуется в одном Go-модуле; compose и тесты зовут одну функцию.

### Q4 — Kafka-образ

| Option | Description | Selected |
|--------|-------------|----------|
| confluentinc/confluent-local (реком.) | Паритет с testcontainers 7.5.0, KRaft из коробки; Mongo mongo:7 RS | ✓ |
| apache/kafka (нативный KRaft) | Вендор-нейтральный, но образы дева/теста различаются | |
| Обсудить версии | Разобрать теги отдельно | |

**User's choice:** confluentinc/confluent-local + mongo:7 single-node RS.
**Notes:** Явная цель — не допустить «в тесте зелено, в compose красно».

---

## Provisioning топиков

### Q5 — запуск provision

| Option | Description | Selected |
|--------|-------------|----------|
| Go CLI в cmd + make-таргет (реком.) | Тонкий cmd зовёт общую функцию; make topics/dev-up; тесты — напрямую | ✓ |
| init-контейнер в compose | compose сам провижнит после healthcheck | |
| make + inline kafka-CLI | Без Go-бинаря, kafka-topics.sh; расходится с «общий провижн в Go» | |

**User's choice:** Go CLI в cmd + make-таргет.

### Q6 — какие топики в Phase 5

| Option | Description | Selected |
|--------|-------------|----------|
| Только host.* через data-driven список (реком.) | host.events+host.state; список агрегатов — данные, расширяется одной строкой | ✓ |
| host + project сразу | 4 топика, оба ядровых агрегата | |
| Все 4 placeholder'ами | host/project/module/location сразу — риск преждевременных имён | |

**User's choice:** Только host.* через data-driven список.

### Q7 — число партиций

| Option | Description | Selected |
|--------|-------------|----------|
| Configurable, дев-дефолт малый, prod отложить (реком.) | Параметр; дев=6; prod-число под 150k = ops, при первом консьюмере | ✓ |
| Прод-запас сразу как канон-дефолт | 24/50 уже сейчас (research «с запасом») | |
| 1 партиция везде | Просто, но не гоняет sticky-key на >1 партиции | |

**User's choice:** Configurable, дев-дефолт малый (6), prod отложить.
**Notes:** Дев=6 заодно проверяет порядок per-entity sticky-key partitioner'а.

---

## Миграция v2 + где тесты

### Q8 — глубина скаффолда

| Option | Description | Selected |
|--------|-------------|----------|
| Тонкий connection-helper + smoke (реком.) | Client-фабрика (RS-aware, writeconcern.Majority) + ping, без repo/UoW | ✓ |
| Только swap go.mod + чистый smoke | Тест открывает client инлайн, без хелпера | |
| Helper + тонкий repository-скелет | Заложить образец репозитория — риск пересечения с Phase 7 | |

**User's choice:** Тонкий connection-helper + smoke.
**Notes:** Граница: подключение = Phase 5; repository/UoW/Outbox = Phase 7.

### Q9 — изоляция интеграционных тестов

| Option | Description | Selected |
|--------|-------------|----------|
| build-tag `integration` + make-таргет (реком.) | go test ./... и pre-push — только unit; TC за тегом | ✓ |
| Отдельный пакет без тега | test/integration, но go test ./... всё равно их запустит | |
| Обсудить | — | |

**User's choice:** build-tag `integration` + make-таргет.
**Notes:** Критично после решения «inventory в pre-push» (Q2) — push не должен поднимать testcontainers.

---

## Claude's Discretion

- **DOC-02 (SC5):** замена audit-рецепта в build.md на команду с exit 0 (кандидат `go vet ./...`); финальная форма — эмпирическая верификация executor'ом.
- **mockery (SVC-06):** провод `.mockery.yaml` + `make generate-mocks`, smoke на одном throwaway-интерфейсе (реальных портов нет до Phase 6/7).
- Структура Makefile-таргетов, healthcheck'и compose, имена констант/переменных, лэйаут пакетов хелпера/теста.

## Deferred Ideas

- Prod-число партиций Kafka (под ~150k) → при первом консьюмере (отдельный milestone).
- Schema Registry → с первым консьюмером (BACKWARD_TRANSITIVE, SEED-002).
- `kprom`/Prometheus-метрики relay → до первой эксплуатации (Phase 8+).
- Redpanda testcontainers-модуль → только если понадобится SASL.
- Полноценный CI → вне Phase 5; решения заложены так, чтобы CI лёг без переделки.
