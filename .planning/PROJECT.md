# gwall-e

## What This Is

gwall-e — платформа **Hardware-as-a-Service** для дата-центров. Она даёт овнерам, SRE и ITDC единый инструмент, чтобы инвентаризировать хосты и VM, видеть их состояние, выписывать SSH-права, выполнять действия над хостами (в т.ч. массовые) и автоматически их чинить — при этом сохраняя согласованность: никто не может «забрать» чужой хост в обход правил.

Технически это набор Go-микросервисов (бэкенд, DDD/гексагональная архитектура) и React/Nx фронтенд.

## Core Value

Безопасное и **согласованное** управление парком серверов как услугой: единый источник правды о хостах и контролируемый, неконфликтный доступ к действиям над ними между овнерами и SRE/ITDC.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

**v1.0 — Фундамент (knowledge base + enforcement), shipped 2026-06-17:**

- ✓ Зафиксированы правила для ИИ/команды в `knowledge/` (10 доков, ~1101 строк): язык (русские комментарии/доменная терминология), тестирование (Ginkgo v2 + Gomega), стиль кода, структура — v1.0
- ✓ Описаны конвенции архитектуры (DDD + гексагон, БЕЗ CQRS-шины) и раскладки репозитория (`go.work`) как воспроизводимые правила (`architecture.md`, `structure.md`, `patterns.md`) — v1.0
- ✓ Описаны процессы работы: сборка/тесты (вкл. `GOWORK=off` для `inventory`), git-конвенции (`build.md`, `git.md`) — v1.0
- ✓ Тонкие точки входа: `AGENTS.md` (источник истины) + урезанный `CLAUDE.md` (51 строка) + `knowledge/README.md` индекс + authoring-стандарт MUST/SHOULD/WON'T — v1.0
- ✓ Enforcement-слой: `.golangci.yml` v2, `lefthook.yml`, commitlint, buf-скелет + статус enforcement на каждом механизируемом правиле — v1.0 (live hook firing → bootstrap UAT)

> **Известный gap (v1.0):** DOC-02 — `build.md` audit-рецепт `cd services/audit && go build ./...` падает (exit 1); рабочие формы — `go build ./cmd` / `go vet ./...`. Принят как tech debt, см. MILESTONES.md и STATE.md Deferred Items.

### Active

<!-- Видение платформы. Это гипотезы до тех пор, пока не отгружены и не подтверждены. -->

**Следующий milestone (кандидаты):**

- [ ] DOC-07: `knowledge/glossary.md` — ubiquitous language доменных терминов (отложено из v1, активируется в domain-milestone когда проектируется доменная модель)
- [ ] Починить DOC-02 build-рецепт + закрыть Nyquist sign-off и live-firing UAT (tech debt из v1.0)

**Видение платформы (будущие эпики/milestone'ы):**

- [ ] Инвентаризация хостов в ДЦ и VM (единый источник правды)
- [ ] Выдача и управление SSH-правами на доступ к хостам
- [ ] Выполнение действий над хостами (ребут, переналивка и т.п.)
- [ ] Массовые работы над группами хостов
- [ ] Автопочинка хостов
- [ ] Мониторинг состояния хостов (здоровье, проверки, CPU и т.д.)
- [ ] Согласованность работ/действий между овнерами и SRE/ITDC (никто не может забрать хосты просто так)

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Любые бизнес-фичи в первом milestone (SSH-права, инвентаризация, действия, массовые работы, автопочинка, мониторинг) — фундамент закладывается отдельно, фичи идут следующими эпиками
- Фронтенд (React/Nx dashboard) в первом milestone — сначала правила и бэкенд-фундамент
- Реальная интеграция с внешними системами (bot-сервис, железо ДЦ) на этапе фундамента — пока заглушки/моки

## Context

- **Стадия:** ранний WIP / скелет. Часть кода (`pkg/mediatr`, `services/inventory`) сейчас в процессе сноса/пересборки — фундамент закладывается заново.
- **Чистый лист:** `services/inventory/internal/` снесён целиком (все файлы удалены), `pkg/mediatr` (бывшая CQRS-шина) удалён, `tx.go`/`TxManager` удалён. Старый код из git HEAD невалиден — фундамент проектируется заново.
- **Целевая архитектура сервиса** (эталон, зафиксировано): слои `domain` (агрегаты, VO, доменные события, порты) / `usecases` (write-side, 1 use case = 1 struct + `Execute`) / `query` (read-side, query-сервисы читают Mongo напрямую в DTO) / `repositories` (Mongo-реализации портов + UnitOfWork) / `api` (gRPC-адаптеры) / `cron` (джобы); `app` — composition root (ручной DI); `cmd` — `main`. Без CQRS-шины/диспетчера: gRPC-хендлеры зовут use case'ы напрямую. Транзакции — через порт `UnitOfWork` (Mongo-транзакция). Доменные события — `PullEvents`, публикация через transactional outbox внутри UnitOfWork-транзакции + отдельный relay.
- **Workspace:** мульти-модульный Go workspace (`go.work`, Go 1.24.6); каждый сервис/пакет — отдельный модуль `github.com/gwall-e/...`. `go.work` включает `./pkg`, `./services/analytics`, `./services/audit`; `inventory` намеренно НЕ в workspace (собирается с `GOWORK=off`).
- **Memory Bank:** в корне есть `knowledge/` — версионируемые правила проекта (тесты, стиль и т.п.), это и есть предмет первого milestone.
- **Shipped v1.0 (2026-06-17):** база знаний `knowledge/` (10 доков, ~1101 строк) + тонкие точки входа (`AGENTS.md`/`CLAUDE.md`) + enforcement-тулинг (`.golangci.yml` v2, `lefthook.yml`, commitlint, buf-скелет, `Makefile` с пиннингом версий). Доки заземлены на реальный репозиторий (no-phantom), каждое механизируемое правило помечено статусом enforcement.


## Constraints

- **Tech stack (backend):** Go 1.24.6, мульти-модульный workspace; DDD + гексагональная архитектура; MongoDB (outbound), GRPC.
- **Tech stack (frontend):** React + Nx.
- **Тесты:** Ginkgo + Gomega; комментарии в тестах на английском, доменные комментарии в коде — на русском.
- **Язык:** комментарии и доменная терминология в Go-коде — на русском (имена идентификаторов — английские); сохранять этот стиль.
- **Сборка:** `inventory` собирать/тестировать с `GOWORK=off` из каталога модуля; при добавлении сервиса в общий build — добавлять модуль в `go.work`.
- **Git:** remote `origin` → `github.com/neeeekto/gwall-e`; основная ветка — `main`.

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Первый milestone — только memory-bank (правила/конвенции), без кода скелета | Большой проект с множеством эпиков; нужен прочный фундамент правил для ИИ/команды до фич | ✓ Good — v1.0 shipped: `knowledge/` + enforcement |
| DDD + гексагональная архитектура как стандарт (БЕЗ CQRS-шины) | Чёткое разделение слоёв и портов; CQRS-диспетчер (mediatr) удалён как лишняя сложность | ✓ Канонизировано в `architecture.md` (v1.0); валидация в коде — pending |
| Write-side: 1 use case = 1 struct + `Execute` (паттерн interactor) | Идиоматично для Go, изолированно, дружелюбно к ИИ; без диспетчера | ✓ В `patterns.md`/`architecture.md` (v1.0); валидация в коде — pending |
| Read-side: query-сервисы читают Mongo напрямую в DTO («CQRS-lite») | Разделение reads/writes без инфраструктуры CQRS-шины | ✓ В `architecture.md` (v1.0); валидация в коде — pending |
| Транзакции через порт `UnitOfWork` (Mongo-транзакция) | `TxManager` удалён; UoW оборачивает запись, репозитории берут транзакцию из ctx | ✓ В `architecture.md`/`patterns.md` (v1.0); валидация в коде — pending |
| Доменные события — transactional outbox внутри UoW + relay | Нет dual-write, at-least-once; поддерживает core value «согласованность» | ✓ В `architecture.md` (v1.0); валидация в коде — pending |
| Язык кода — русские комментарии/доменная терминология (имена — англ.) | Снять противоречие EN/RU; единый ubiquitous language | ✓ Good — единственный канон в `style.md`, enforcement-метки (v1.0) |
| `inventory` вне `go.work` (сборка через `GOWORK=off`) | Изолировать незавершённый эталонный сервис от общего build | ✓ Good — задокументировано в `structure.md`/`build.md` (v1.0) |
| Фундамент проектируется заново (снесён `internal/` inventory, `pkg/mediatr`, `tx.go`) | Старый код — леса; архитектура переосмыслена без CQRS/TxManager | ✓ Good — MUST NOT возрождать CQRS зафиксирован + depguard ban на `pkg/mediatr` (v1.0) |
| Enforcement: golangci-lint v2 + lefthook + commitlint + buf, версии пиннятся в `Makefile` | Механизировать правила; локальная проводка хуков до полного CI | ⚠️ Revisit — live firing требует bootstrap (`make tools`+`lefthook install`); CI ещё нет |
| DOC-07 (glossary) отложен из v1 в domain-milestone | Доменная модель ещё не спроектирована; риск расхождения | — Pending — активируется в domain-milestone |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-06-17 after v1.0 milestone — Фундамент (knowledge base + enforcement). 4 фазы / 14 планов / 17 v1-требований ✓ (DOC-07 отложен в domain-milestone). База знаний `knowledge/` (10 доков), тонкие точки входа (`AGENTS.md`/`CLAUDE.md`), enforcement-тулинг (golangci-lint v2 / lefthook / commitlint / buf-скелет). Известный gap: DOC-02 (build.md audit-рецепт). Дальше — `/gsd-new-milestone`.*
