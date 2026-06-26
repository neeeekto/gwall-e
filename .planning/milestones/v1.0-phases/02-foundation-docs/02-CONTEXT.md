# Phase 2: Стабильные доки-основы - Context

**Gathered:** 2026-06-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Фаза наполняет контентом **инфра/процесс-доки** базы знаний — доки навигации и процессов
для ИИ/команды, **без описания доменной модели системы**. В скоупе четыре дока:
`structure.md` (раскладка репо/workspace), `build.md` (команды сборки/тестов),
`git.md` (git-конвенции), `boundaries.md` (границы «не трогать» + анти-дублирование).

**Формат уже зафиксирован Phase 1** (не переобсуждается): плоская раскладка `knowledge/*.md`,
kebab-case английские имена файлов, русское содержимое (тех-термины/примеры — английские),
authoring-стандарт MUST/SHOULD/WON'T + парность «запрет → do» ([knowledge/authoring.md](../../../knowledge/authoring.md)),
индекс в [knowledge/README.md](../../../knowledge/README.md). Наполнение снимает статус
«запланировано» и добавляет ссылки в индекс.

**Скоуп-изменение в этой сессии:** `glossary.md` (**DOC-07**) **выведен из этого milestone**
и отложен в будущий domain-milestone (см. `<decisions>` D-01 и `<deferred>`). Доменный язык
(host/VM/owner/SRE/ITDC и их отношения) фиксируется тогда, когда проектируется сама система,
а не на этапе фундамента правил для ИИ. ROADMAP.md и REQUIREMENTS.md обновлены соответственно.

Требования (после декскоупа): **DOC-01** (`structure.md`), **DOC-02** (`build.md`),
**DOC-06** (`git.md`), **DOC-08** (`boundaries.md`). **DOC-07** (`glossary.md`) — отложено.

</domain>

<decisions>
## Implementation Decisions

### Скоуп milestone (мета-решение)
- **D-01:** **`glossary.md` (DOC-07) отложен** в будущий domain-milestone. Доменная модель ещё
  не спроектирована — фиксировать её сейчас = риск расхождения с будущей моделью и противоречит
  цели milestone'а («фундамент правил для ИИ/команды», не «описание системы»). Phase 2 = чистая
  инфра/процесс-база. `architecture.md` (Phase 3) опирается на DDD/гексагон-конвенции (слои, порты),
  а не на доменный глоссарий, — проживёт без него. **Action:** обновить ROADMAP (убрать glossary
  из success criteria Phase 2 и из зависимостей Phase 3) и REQUIREMENTS (DOC-07 → v2/deferred).

### structure.md (DOC-01)
- **D-02:** Описывать репозиторий **на уровне модулей/workspace**, без целевой слоистой раскладки
  сервиса (domain/usecases/query/...). Внутренняя слоистая раскладка — это доменно-архитектурный
  док, отложен (architecture.md Phase 3 / domain-milestone). Здесь: что в `go.work` (`./pkg`,
  `./services/analytics`, `./services/audit`), что вне (`inventory` — `GOWORK=off`), мульти-модульный
  workspace, Go 1.24.6, фронтенд `web/` (Nx) как отдельная часть.
- **D-03:** Пустые stale-каталоги в `services/inventory/internal/` (`api/app/cron/domain/repository/usecase`
  — singular, без `query`) **НЕ фиксировать как авторитетную структуру**. Явно: `inventory`
  пересобирается, его текущая внутренняя раскладка невалидна и не эталон; ссылка на `boundaries.md`
  (не «чинить» леса). Это пример phantom-наоборот: не документировать снесённое как структуру.
- **D-04:** Сервисы `gateway` и `outgate` (только `README` + `go.mod`, кода нет, вне `go.work`)
  **не упоминать** в structure.md. Документируем только активные модули (`pkg`, `analytics`,
  `audit`, `inventory`). Стабы добавим, когда появится код.

### git.md (DOC-06)
- **D-05:** **Общие конвенции + краткое упоминание GSD**. Канон: ветки (`dev`/`main`,
  remote `origin` → `github.com/neeeekto/gwall-e`), Conventional Commits, нормы PR, «когда коммитить».
  Плюс краткий блок GSD-практик (коммиты в `.planning/`, phase-scoped scope вида `docs(NN):`,
  Co-Authored-By trailer, фильтрация `.planning` из PR через `gsd-pr-branch`) — со ссылкой, что это
  GSD-тулинг, а не общий стандарт. Полный GSD git-workflow дублировать — **WON'T** (живёт в GSD-доках).

### boundaries.md (DOC-08)
- **D-06:** Правила «не трогать» — **на уровне возможностей + примеры, без хрупких path-карт**:
  «WIP-сервисы вне `go.work` не чинить/не расширять», «стале `README`/`Makefile`/`docker-compose.yml`
  не авторитетны», «не документировать несуществующие фичи (phantom rules)». `inventory`/`gateway`/`outgate`
  — как **примеры**, не как жёсткий список путей (соответствует authoring.md: без хрупких путей).
- **D-07:** Перенести два готовых посева из Phase 1:
  (а) «что НЕ кладём в `knowledge/`» — черновики/спайки/планы (`.planning/`) не канон;
  (б) риск ре-раздувания корневого `CLAUDE.md` через `generate-claude-md` — зафиксировать как
  правило (не запускать авто-генерацию тяжёлых секций / держать тонкий гибрид D-06 Phase 1).
- **D-08:** **Явная карта владения фактами** (pointer-over-copy): короткая таблица «какой факт где
  канон» — язык кода → `style.md` (Phase 3); WIP-статус `inventory` → `structure.md` (boundaries
  ссылается); команды сборки/тестов → `build.md`. Один факт = один канон, остальные доки ссылаются,
  не копируют. (Расширяет существующий блок «Что где живёт» в `knowledge/README.md`.)

### Claude's Discretion
- **build.md (DOC-02)** — не выбирался для детального обсуждения, отдан на усмотрение планировщика
  в рамках принципов сессии: документировать **реально работающие** команды (`cd pkg && go test`,
  сборка/тесты `analytics`/`audit` в workspace), `GOWORK=off` для `inventory` с WIP-пометкой
  (модуль сейчас не гарантированно компилируется — пометить, не выдавать за рабочее), фронтенд
  `npx nx` как инфра-справку. Без phantom-команд (не документировать то, что не запускается).
  Команды — на уровне рецептов, не хрупких длинных листингов (authoring.md).
- Точное расположение «карты владения» (D-08): в `boundaries.md` или расширением `README.md` —
  на усмотрение планировщика; обязателен сам факт наличия карты.
- Конкретная разбивка доков на под-файлы при превышении ~150–200 строк (authoring.md) — по факту.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Источник истины по скоупу и решениям
- `.planning/PROJECT.md` — Key Decisions (DDD+гексагон БЕЗ CQRS, язык кода, `inventory` вне `go.work`,
  снос `internal/`/`pkg/mediatr`/`tx.go`), Context (целевая архитектура сервиса, workspace), Constraints.
- `.planning/REQUIREMENTS.md` — DOC-01/02/06/08 (скоуп Phase 2 после декскоупа), DOC-07 (отложено),
  карта v1/v2 (что НЕ документировать сейчас).
- `.planning/ROADMAP.md` — Phase 2 goal и success criteria (обновлены: без glossary).

### Стандарт и индекс базы знаний (созданы в Phase 1 — следовать им)
- `knowledge/authoring.md` — authoring-стандарт: MUST/SHOULD/WON'T, парность «запрет → do»,
  pointer-over-copy, размер/дробление, no-phantom, статус enforcement. **Все доки Phase 2 MUST следовать.**
- `knowledge/README.md` — индекс базы: таблица доков (со статусом «запланировано» для Phase 2/3),
  порядок чтения, блок «Что где живёт» (knowledge/ = канон, .planning/ = процесс). Обновляется при наполнении.
- `AGENTS.md` / `CLAUDE.md` (корень) — тонкие точки входа; ссылаются на `knowledge/`.

### Research (раскладка базы знаний и грабли)
- `.planning/research/PITFALLS.md` — Pitfall 2 (мега-файл/контекст-бюджет), Pitfall 8 (phantom-правила),
  Pitfall 5 (MUST/SHOULD/WON'T + «do»).
- `.planning/research/FEATURES.md` — table-stakes доки, anti-features (что НЕ класть), порядок зависимостей.
- `.planning/research/ARCHITECTURE.md` — структура базы знаний (тонкие входы + progressive disclosure).

### Прецедент (как писались доки в этой базе)
- `.planning/phases/01-knowledge-base-layout/01-CONTEXT.md` — решения D-01..D-08 Phase 1
  (раскладка, имена, язык, источник истины AGENTS.md, риск регенерации CLAUDE.md).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `knowledge/README.md`, `knowledge/authoring.md` — существуют и задают формат/индекс. Наполнение
  Phase 2 встраивается в этот индекс (снять статус «запланировано», добавить ссылки).

### Established Patterns
- Progressive disclosure: тонкие входы (`AGENTS.md`/`CLAUDE.md`) → topic-доки `knowledge/*.md`.
- Authoring-стандарт (MUST/SHOULD/WON'T + «do», pointer-over-copy) — обязателен для нового контента.

### Integration Points
- `knowledge/README.md` индекс и порядок чтения — обновить при появлении каждого нового дока
  (без битых ссылок: ссылка появляется вместе с файлом).
- `boundaries.md` ↔ `structure.md` — взаимные ссылки по WIP-статусу `inventory` (D-03, D-08).

### Risks / Constraints for planner
- ⚠️ **go.work / реальность репо:** в `go.work` только `pkg`, `analytics`, `audit`. `inventory`,
  `gateway`, `outgate` — вне workspace. `inventory/internal/*` — пустые stale-каталоги (D-03):
  НЕ документировать как структуру.
- ⚠️ **Phantom-команды (build.md):** `inventory` может не компилироваться — команды для него
  помечать WIP, не выдавать за проверенно рабочие.
- ⚠️ **Регенерация CLAUDE.md:** `generate-claude-md` может ре-раздуть корневой `CLAUDE.md` —
  зафиксировать запрет/осторожность в `boundaries.md` (D-07), перенос риска из Phase 1.
- ⚠️ **Кросс-док дубли:** язык кода → `style.md` (Phase 3), не дублировать в Phase 2 доках;
  владение фактами — по карте D-08.

</code_context>

<specifics>
## Specific Ideas

- Пользователь явно переосмыслил скоуп в ходе сессии: «сейчас цель — подготовить ИИ-базу, инфра-базу,
  без доменных доков; систему описываем в отдельном Milestone». Это привело к декскоупу glossary (D-01)
  и к выбору модульно-уровневого structure.md (D-02). Downstream: держаться инфра/процесс-уровня,
  не уходить в описание доменной модели.
- git.md: учесть реальные конвенции репо, уже видимые в истории — Conventional Commits с phase-scope
  (`docs(01):`, `docs(phase-1):`), ветки `dev`/`main`, Co-Authored-By trailer.

</specifics>

<deferred>
## Deferred Ideas

- **`glossary.md` (DOC-07) → domain-milestone** (D-01): доменный ubiquitous language (host, VM, owner,
  SRE, ITDC, namespace, project + роли/отношения доступа, «согласованность»/«нельзя забрать чужой хост»)
  фиксируется, когда проектируется система. Не в фундаменте правил для ИИ.
- Целевая слоистая раскладка сервиса (domain/usecases/query/repositories/api/cron/app/cmd) —
  `architecture.md` (Phase 3), на уровне конвенций; конкретная доменная раскладка — domain-milestone.
- ADR-доки, `anti-patterns.md`, `libraries.md`, onboarding, maintenance-протокол — v2 (REQUIREMENTS.md).
- Полный GSD git-workflow — не дублируем в `git.md` (D-05); живёт в GSD-доках.

</deferred>

---

*Phase: 2-Стабильные доки-основы*
*Context gathered: 2026-06-17*
