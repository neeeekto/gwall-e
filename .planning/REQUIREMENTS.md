# Requirements: gwall-e — Фундамент (knowledge base + enforcement)

**Defined:** 2026-06-17
**Core Value:** Безопасное и согласованное управление парком серверов как услугой; этот milestone закладывает фундамент конвенций для ИИ/команды.

> **Scope этого milestone:** база знаний `knowledge/` (правила для ИИ/команды) + проводка enforcement-тулинга. Бизнес-фичи платформы (инвентаризация, SSH-права, действия, массовые работы, автопочинка, мониторинг) и фронтенд — вне скоупа, отдельными эпиками.

## v1 Requirements

### Knowledge Base — структура и точки входа (KB)

- [ ] **KB-01**: Корневой `CLAUDE.md` урезан до тонкого индекса (~<150 строк), который ссылается на `knowledge/*.md` (progressive disclosure), без дублирования деталей
- [ ] **KB-02**: Есть `knowledge/README.md` — индекс базы знаний с порядком чтения и 1-строчным назначением каждого дока
- [ ] **KB-03**: Есть `AGENTS.md` как тонкий кросс-тульный указатель на `CLAUDE.md`/`knowledge/` (без дублирования контента)
- [ ] **KB-04**: Зафиксирован authoring-стандарт: каждое нормативное правило помечается MUST/SHOULD/WON'T, каждый запрет сопровождается предписанной альтернативой («do»)

### Доки конвенций (DOC)

- [ ] **DOC-01**: `knowledge/structure.md` — раскладка `go.work`, какие модули в/вне workspace, статус `inventory` как WIP (на уровне возможностей, без хрупкой карты путей)
- [ ] **DOC-02**: `knowledge/build.md` — команды сборки/запуска/тестов, включая `GOWORK=off` для `inventory`, `cd pkg && go test`, фронтенд `npx nx`
- [ ] **DOC-03**: `knowledge/testing.md` — конвенции тестов: Ginkgo v2 + Gomega, комментарии в тестах на английском, структура спеков
- [ ] **DOC-04**: `knowledge/style.md` — канонический MUST по языку (русские комментарии/доменная терминология; имена — английские); типизированные ID; sentinel vs обёрнутые ошибки; маппинг DTO→домен внутри хендлера
- [ ] **DOC-05**: `knowledge/architecture.md` — DDD + гексагон (БЕЗ CQRS-шины): правила слоёв/импортов, usecases-interactor (`Execute`), query-lite (read-side в DTO), порт `UnitOfWork`, transactional outbox + relay, фабрики агрегатов и `PullEvents`; явный MUST NOT возрождать CQRS-диспетчер/`TxManager`
- [ ] **DOC-06**: `knowledge/git.md` — git-конвенции: ветки, Conventional Commits, нормы PR, когда коммитить
- [ ] **DOC-07**: `knowledge/glossary.md` — ubiquitous language: доменные термины (host, VM, owner, SRE, ITDC, namespace, project и т.д.) с маппингом EN/RU
- [ ] **DOC-08**: `knowledge/boundaries.md` — правила «do-not»: не чинить/не расширять WIP-леса; стале `README`/`Makefile`/`docker-compose.yml` не авторитетны; не документировать несуществующие фичи (phantom rules)

### Pattern catalog (PAT)

- [ ] **PAT-01**: `knowledge/patterns.md` — рецепты «как добавить use case / query / aggregate / repository» как копируемые пошаговые процедуры, согласованные с `architecture.md`

### Enforcement-тулинг (ENF)

- [ ] **ENF-01**: `.golangci.yml` (golangci-lint v2; gofumpt как форматтер; gci для порядка импортов), консистентный с workspace
- [ ] **ENF-02**: `lefthook.yml` — хуки: pre-commit (lint + format), pre-push (тесты), commit-msg (commitlint)
- [ ] **ENF-03**: Конфиг commitlint (Conventional Commits), подключённый к commit-msg хуку
- [ ] **ENF-04**: Скелет `buf.yaml` + `buf.gen.yaml` для proto (lint / breaking / codegen)
- [ ] **ENF-05**: Каждое механизируемое правило в `knowledge/*.md` помечено статусом enforcement (CI-gated / hook / convention-only)

## v2 Requirements

Acknowledged, но вне текущего milestone.

### Knowledge Base (расширения)

- **ADR-01**: `decisions/ADR-*.md` — записи решений (без CQRS, UnitOfWork+outbox, RU-комментарии, inventory вне go.work), посеянные из PROJECT.md Key Decisions
- **DOC-09**: `knowledge/anti-patterns.md` — каталог анти-паттернов
- **DOC-10**: `knowledge/libraries.md` — справочник общих пакетов `pkg/`
- **DOC-11**: Onboarding-гайд (человеко-ориентированный quickstart)
- **DOC-12**: Maintenance/self-update протокол базы знаний

## Out of Scope

| Feature | Reason |
|---------|--------|
| Все бизнес-фичи (инвентаризация, SSH-права, действия над хостами, массовые работы, автопочинка, мониторинг) | Фундамент закладывается отдельно; фичи — будущие эпики/milestone'ы |
| Фронтенд (React/Nx dashboard) | Сначала правила и бэкенд-фундамент |
| Реальная интеграция с внешними системами (bot-сервис, железо ДЦ) | На этапе фундамента — заглушки/моки |
| Reference-service walkthrough (`reference-service.md`) | `inventory` сейчас не собирается; разблокируется, когда сервис компилируется |
| Восстановление кода `inventory`/общего application-слоя | Это отдельный эпик реализации, не milestone правил |
| Полный CI-pipeline (выбор раннера, матрицы) | Этот milestone делает локальную проводку (lefthook/линтеры); CI — позже |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| KB-01 | Phase 1 | Pending |
| KB-02 | Phase 1 | Pending |
| KB-03 | Phase 1 | Pending |
| KB-04 | Phase 1 | Pending |
| DOC-07 | Phase 2 | Pending |
| DOC-01 | Phase 2 | Pending |
| DOC-02 | Phase 2 | Pending |
| DOC-06 | Phase 2 | Pending |
| DOC-08 | Phase 2 | Pending |
| DOC-04 | Phase 3 | Pending |
| DOC-03 | Phase 3 | Pending |
| DOC-05 | Phase 3 | Pending |
| PAT-01 | Phase 3 | Pending |
| ENF-01 | Phase 4 | Pending |
| ENF-02 | Phase 4 | Pending |
| ENF-03 | Phase 4 | Pending |
| ENF-04 | Phase 4 | Pending |
| ENF-05 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 18 total
- Mapped to phases: 18 ✓
- Unmapped: 0

---
*Requirements defined: 2026-06-17*
*Last updated: 2026-06-17 after roadmap creation (traceability filled)*
