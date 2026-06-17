# gwall-e

## What This Is

gwall-e — платформа **Hardware-as-a-Service** для дата-центров. Она даёт овнерам, SRE и ITDC единый инструмент, чтобы инвентаризировать хосты и VM, видеть их состояние, выписывать SSH-права, выполнять действия над хостами (в т.ч. массовые) и автоматически их чинить — при этом сохраняя согласованность: никто не может «забрать» чужой хост в обход правил.

Технически это набор Go-микросервисов (бэкенд, DDD/гексагональная архитектура) и React/Nx фронтенд.

## Core Value

Безопасное и **согласованное** управление парком серверов как услугой: единый источник правды о хостах и контролируемый, неконфликтный доступ к действиям над ними между овнерами и SRE/ITDC.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — ship to validate)

### Active

<!-- Видение платформы. Это гипотезы до тех пор, пока не отгружены и не подтверждены.
     Конкретный текущий milestone — "Фундамент / memory-bank" — см. REQUIREMENTS.md. -->

**Текущий milestone — Фундамент:**

- [ ] Зафиксированы правила для ИИ/команды в `memory-bank/`: язык (русские комментарии и доменная терминология), тестирование (Ginkgo + Gomega), стиль кода и структура
- [ ] Описаны конвенции архитектуры (DDD + гексагон, без CQRS-шины) и раскладки репозитория как воспроизводимые правила
- [ ] Описаны процессы работы: сборка/тесты (в т.ч. `GOWORK=off` для модулей вне workspace), git-конвенции

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
| Первый milestone — только memory-bank (правила/конвенции), без кода скелета | Большой проект с множеством эпиков; нужен прочный фундамент правил для ИИ/команды до фич | — Pending |
| DDD + гексагональная архитектура как стандарт (БЕЗ CQRS-шины) | Чёткое разделение слоёв и портов; CQRS-диспетчер (mediatr) удалён как лишняя сложность | — Pending |
| Write-side: 1 use case = 1 struct + `Execute` (паттерн interactor) | Идиоматично для Go, изолированно, дружелюбно к ИИ; без диспетчера | — Pending |
| Read-side: query-сервисы читают Mongo напрямую в DTO («CQRS-lite») | Разделение reads/writes без инфраструктуры CQRS-шины | — Pending |
| Транзакции через порт `UnitOfWork` (Mongo-транзакция) | `TxManager` удалён; UoW оборачивает запись, репозитории берут транзакцию из ctx | — Pending |
| Доменные события — transactional outbox внутри UoW + relay | Нет dual-write, at-least-once; поддерживает core value «согласованность» | — Pending |
| Язык кода — русские комментарии/доменная терминология (имена — англ.) | Снять противоречие EN/RU; единый ubiquitous language | — Pending |
| `inventory` вне `go.work` (сборка через `GOWORK=off`) | Изолировать незавершённый эталонный сервис от общего build | — Pending |
| Фундамент проектируется заново (снесён `internal/` inventory, `pkg/mediatr`, `tx.go`) | Старый код — леса; архитектура переосмыслена без CQRS/TxManager | — Pending |

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
*Last updated: 2026-06-17 after Phase 3 (доки конвенций и целевой архитектуры: `style.md`, `testing.md`, `architecture.md` DDD+гексагон БЕЗ CQRS, копируемый `patterns.md`). DOC-03/DOC-04/DOC-05/PAT-01 ✓. Ранее: Phase 1 (раскладка `knowledge/` + точки входа, KB-01..04 ✓), Phase 2 (foundation-доки). Дальше — Phase 4: enforcement-слой (тулинг/CI).*
