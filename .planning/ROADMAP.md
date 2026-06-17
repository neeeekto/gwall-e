# Roadmap: gwall-e — Фундамент (knowledge base + enforcement)

## Overview

Первый milestone закладывает фундамент конвенций для ИИ/команды: версионируемую базу знаний `knowledge/` плюс проводку enforcement-тулинга. Путь идёт строго по зависимостям: сначала **раскладка и тонкие точки входа** + authoring-стандарт (MUST/SHOULD/WON'T, запрет всегда с альтернативой), затем **стабильные доки-основы** (структура репо, сборка, git, границы), потом **доки конвенций и архитектуры** (стиль/язык, тесты, DDD+гексагон БЕЗ CQRS-шины, каталог паттернов) и в конце **enforcement-слой** (golangci-lint v2, gofumpt/gci, lefthook, commitlint, buf + тегирование статуса enforcement). Бизнес-фичи, фронтенд и **доменные доки (глоссарий ubiquitous language)** — вне скоупа: глоссарий отложен в будущий domain-milestone, т.к. доменная модель ещё не спроектирована, а этот milestone закладывает правила для ИИ/команды, а не описание системы.

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Раскладка базы знаний и точки входа** - `knowledge/` + тонкий `CLAUDE.md`/`AGENTS.md` + authoring-стандарт (completed 2026-06-17)
- [x] **Phase 2: Стабильные доки-основы** - структура, сборка, git, границы (глоссарий отложен в domain-milestone) (completed 2026-06-17)
- [x] **Phase 3: Доки конвенций и архитектуры** - стиль/язык, тесты, DDD+гексагон (без CQRS), паттерны (completed 2026-06-17)
- [ ] **Phase 4: Enforcement-слой (тулинг)** - golangci-lint v2, lefthook, commitlint, buf + статус enforcement

## Phase Details

### Phase 1: Раскладка базы знаний и точки входа

**Goal**: Структура базы знаний и тонкие точки входа зафиксированы; единый authoring-стандарт задан так, что все последующие доки ему следуют
**Depends on**: Nothing (first phase)
**Requirements**: KB-01, KB-02, KB-03, KB-04
**Success Criteria** (what must be TRUE):

  1. Существует каталог `knowledge/` с `README.md`, который перечисляет порядок чтения и 1-строчное назначение каждого дока
  2. Корневой `CLAUDE.md` урезан до тонкого индекса (~<150 строк), ссылается на `knowledge/*.md` (progressive disclosure) и не дублирует детали
  3. Существует `AGENTS.md` как тонкий кросс-тульный указатель на `CLAUDE.md`/`knowledge/` без дублирования контента
  4. Зафиксирован authoring-стандарт: каждое нормативное правило помечается MUST/SHOULD/WON'T, каждый запрет сопровождается предписанной альтернативой («do»)

**Plans:** 2/2 plans complete
**Wave 1**

- [x] 01-01-PLAN.md — authoring-стандарт (`knowledge/authoring.md`) + индекс (`knowledge/README.md`)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — `AGENTS.md` (источник истины) + урезание `CLAUDE.md` до тонкого гибрида

### Phase 2: Стабильные доки-основы

**Goal**: Команда/ИИ имеют стабильные инфра/процесс-доки навигации — раскладка репозитория, команды сборки/тестов, git-конвенции и границы «не трогать» (без описания доменной модели)
**Depends on**: Phase 1
**Requirements**: DOC-01, DOC-02, DOC-06, DOC-08

> **Скоуп-изменение (2026-06-17):** `glossary.md` (DOC-07) выведен из Phase 2 и отложен в будущий domain-milestone — доменная модель ещё не спроектирована, фундамент правил для ИИ её не описывает. Phase 3 `architecture.md` опирается на DDD/гексагон-конвенции (слои/порты), а не на доменный глоссарий, и проживёт без него.

**Success Criteria** (what must be TRUE):

  1. `knowledge/structure.md` описывает раскладку `go.work` и какие модули в/вне workspace (статус `inventory` как WIP) на уровне возможностей, без хрупкой карты путей
  2. `knowledge/build.md` даёт команды сборки/запуска/тестов, включая `GOWORK=off` для `inventory`, `cd pkg && go test`, фронтенд `npx nx`
  3. `knowledge/git.md` фиксирует git-конвенции: ветки, Conventional Commits, нормы PR, когда коммитить
  4. `knowledge/boundaries.md` содержит правила «do-not»: не чинить/не расширять WIP-леса; стале `README`/`Makefile`/`docker-compose.yml` не авторитетны; не документировать несуществующие фичи

**Plans:** 3/3 plans complete

**Wave 1**

- [x] 02-01-PLAN.md — `structure.md` (раскладка go.work) + `build.md` (команды сборки/тестов) + индекс

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 02-02-PLAN.md — `git.md` (git-конвенции + краткий GSD-блок) + индекс

**Wave 3** *(blocked on Wave 1+2 completion)*

- [x] 02-03-PLAN.md — `boundaries.md` (do-not правила + карта владения фактами) + обратная ссылка structure.md + индекс

### Phase 3: Доки конвенций и архитектуры

**Goal**: Канонические правила стиля/языка, тестирования и целевой архитектуры (DDD + гексагон БЕЗ CQRS-шины) зафиксированы вместе с копируемым каталогом паттернов
**Depends on**: Phase 2
**Requirements**: DOC-04, DOC-03, DOC-05, PAT-01
**Success Criteria** (what must be TRUE):

  1. `knowledge/style.md` даёт канонический MUST по языку (русские комментарии/доменная терминология; имена — английские), типизированные ID, sentinel vs обёрнутые ошибки, маппинг DTO→домен внутри хендлера — и является единственным местом правила про язык комментариев
  2. `knowledge/testing.md` фиксирует конвенции тестов: Ginkgo v2 + Gomega, комментарии в тестах на английском, структуру спеков
  3. `knowledge/architecture.md` существует и явно заявляет DDD + гексагон БЕЗ CQRS-шины: правила слоёв/импортов, usecases-interactor (`Execute`), query-lite (read-side в DTO), порт `UnitOfWork`, transactional outbox + relay, фабрики агрегатов и `PullEvents`; содержит явный MUST NOT возрождать CQRS-диспетчер/`TxManager`
  4. `knowledge/patterns.md` даёт копируемые пошаговые рецепты «как добавить use case / query / aggregate / repository», согласованные с `architecture.md`

**Plans:** 5/5 plans complete

**Wave 1**

- [x] 03-01-PLAN.md — `style.md` (язык кода, typed IDs, sentinel/`%w`, DTO→домен) [DOC-04]

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 03-02-PLAN.md — `testing.md` (Ginkgo v2 + Gomega, suite-бутстрап, mockery-конвенция) [DOC-03]
- [x] 03-03-PLAN.md — `architecture.md` (DDD+гексагон БЕЗ CQRS, слои/импорты, инварианты, MUST NOT) [DOC-05]

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 03-04-PLAN.md — `patterns.md` (4 рецепта вертикальным срезом, ссылки на architecture.md) [PAT-01]

**Wave 4** *(blocked on Waves 1–3 completion)*

- [x] 03-05-PLAN.md — интеграция: индекс README.md + карта владения boundaries.md + статусы AGENTS.md (no broken links) [DOC-04, DOC-03, DOC-05, PAT-01]

### Phase 4: Enforcement-слой (тулинг)

**Goal**: Механизируемые правила базы знаний подкреплены тулингом, а каждое правило помечено статусом enforcement — база перестаёт быть только декларативной
**Depends on**: Phase 3
**Requirements**: ENF-01, ENF-02, ENF-03, ENF-04, ENF-05
**Success Criteria** (what must be TRUE):

  1. Существует `.golangci.yml` (golangci-lint v2; gofumpt как форматтер; gci для порядка импортов), консистентный с workspace
  2. Существует `lefthook.yml` с хуками: pre-commit (lint + format), pre-push (тесты), commit-msg (commitlint)
  3. Существует конфиг commitlint (Conventional Commits), подключённый к commit-msg хуку
  4. Существует скелет `buf.yaml` + `buf.gen.yaml` для proto (lint / breaking / codegen)
  5. Каждое механизируемое правило в `knowledge/*.md` помечено статусом enforcement (CI-gated / hook / convention-only)

**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Раскладка базы знаний и точки входа | 2/2 | Complete    | 2026-06-17 |
| 2. Стабильные доки-основы | 3/3 | Complete   | 2026-06-17 |
| 3. Доки конвенций и архитектуры | 5/5 | Complete    | 2026-06-17 |
| 4. Enforcement-слой (тулинг) | 0/TBD | Not started | - |
