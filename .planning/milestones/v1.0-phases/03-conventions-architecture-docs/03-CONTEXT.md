# Phase 3: Доки конвенций и архитектуры - Context

**Gathered:** 2026-06-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Фаза наполняет контентом **доки конвенций и целевой архитектуры** базы знаний — четыре
канонических дока для ИИ/команды. В скоупе:
`style.md` (язык/стиль кода), `testing.md` (конвенции тестов), `architecture.md`
(DDD + гексагон **БЕЗ CQRS-шины**: слои/импорты/порты/события), `patterns.md`
(копируемые пошаговые рецепты). Это последний «контентный» milestone-слой перед
enforcement-тулингом (Phase 4).

**Формат уже зафиксирован Phase 1–2** (не переобсуждается): плоская раскладка
`knowledge/*.md`, kebab-case английские имена файлов, **русское содержимое**
(тех-термины/примеры кода — английские), authoring-стандарт MUST/SHOULD/WON'T +
парность «запрет → do» ([knowledge/authoring.md](../../../knowledge/authoring.md)),
pointer-over-copy и карта владения фактами ([knowledge/boundaries.md](../../../knowledge/boundaries.md)),
индекс/порядок чтения в [knowledge/README.md](../../../knowledge/README.md). Наполнение
снимает статус «запланировано» и добавляет ссылки в индекс — **без битых ссылок**
(ссылка появляется вместе с файлом).

**Доменная модель остаётся вне скоупа** (D-01 Phase 2): `glossary.md` (ubiquitous
language host/VM/owner/SRE/ITDC) отложен в domain-milestone. `architecture.md` опирается
на DDD/гексагон-**конвенции** (слои, порты), а не на доменный глоссарий, и примеры
строятся на нейтральном плейсхолдере (D-05 ниже).

Требования: **DOC-04** (`style.md`), **DOC-03** (`testing.md`), **DOC-05**
(`architecture.md`), **PAT-01** (`patterns.md`).

</domain>

<decisions>
## Implementation Decisions

### patterns.md — глубина рецептов (PAT-01)
- **D-01:** Рецепты используют **иллюстративные Go-сниппеты**, реальные и идиоматичные, но
  **явно помеченные** как «целевой вид / иллюстрация» — НЕ из компилируемого файла. Это даёт
  «копируемость» PAT-01, не нарушая no-phantom (boundaries.md): доки не утверждают, что такой
  файл/сервис уже существует. **Контекст:** эталонный сервис сейчас не компилируется
  (`inventory/internal/` снесён), reference-service walkthrough отложен в Out of Scope до
  момента, когда сервис начнёт собираться.
- **D-02:** Каждый рецепт — **вертикальный срез до wiring включительно**: создать
  `struct + Execute` → объявить/использовать порты и репозиторий → зарегистрировать в
  composition root (`app`, ручной DI) → выставить gRPC-адаптер в `api`. Не ограничиваться
  механикой одного слоя — показать полный путь «как подключить».
- **D-03:** Каталог рецептов покрывает (success criteria PAT-01): «добавить use case»,
  «добавить query (read-side, DTO)», «добавить aggregate (+ фабрика, `PullEvents`)»,
  «добавить repository (Mongo-реализация порта + участие в UnitOfWork)». Рецепты согласованы
  с `architecture.md` и ссылаются на него за правилами (не повторяют их — pointer-over-copy).

### architecture.md ↔ patterns.md — граница (DOC-05 / PAT-01)
- **D-04:** **«Правила vs рецепты»** (pointer-over-copy): `architecture.md` = инварианты/правила/
  *why* (слои, направление импортов, порт `UnitOfWork`, transactional outbox + relay, фабрики
  агрегатов + `PullEvents`, query-lite read-side в DTO, usecases-interactor `Execute`, явный
  **MUST NOT** возрождать CQRS-диспетчер/`pkg/mediatr`/`TxManager`). `patterns.md` = пошаговые
  how-to со сниппетами, ссылается на `architecture.md`, правила не дублирует.
- **D-05:** Примеры в обоих доках строятся на **едином нейтральном плейсхолдере-агрегате**
  (напр. `Order`/`Widget`), сквозном между `patterns.md` и `architecture.md`. **НЕ** использовать
  доменные термины gwall-e (host/VM/owner) — чтобы не предрешать доменную модель до
  domain-milestone (согласовано с D-01 Phase 2). Конкретное имя плейсхолдера — на усмотрение
  планировщика, но одно и то же во всех доках.
- **D-06:** `architecture.md` содержит **текстовую диаграмму/таблицу направления импортов**
  между слоями (напр. `domain ← usecases ← api`/`repositories`; `domain` не импортирует ничего
  наружу) — фиксирует главный инвариант гексагона наглядно (ASCII или таблица «кто кого может
  импортировать»).

### style.md — охват и формат (DOC-04)
- **D-07:** Охват — **только проект-специфика** gwall-e: язык (русские комментарии/доменная
  терминология; имена идентификаторов — английские; комментарии в тестах — английские),
  типизированные ID, sentinel vs обёрнутые ошибки (`%w`), маппинг DTO→домен **внутри хендлера**.
  Общий Go-стиль (формат/нейминг) **не дублировать** — ссылка на gofumpt (Phase 4) + Effective Go.
  `style.md` — **единственное место** канона про язык комментариев (карта владения, D-08 Phase 2).
- **D-08:** Формат каждого правила — **«правило + плохо/хорошо»**: короткий MUST/SHOULD + мини-пример
  «плохо → хорошо» (согласовано с парностью «запрет → do» authoring.md). Особенно для typed IDs,
  sentinel-vs-wrapped errors, DTO→домен.

### testing.md — конвенции (DOC-03)
- **D-09:** Степень мандата: **MUST на каркас + SHOULD на структуру**. MUST: suite-бутстрап
  (`RegisterFailHandler(Fail)` + `RunSpecs(t, "...")` — как в реальном `pkg/http`), комментарии
  в тестах на английском, тесты лежат рядом с кодом (`*_test.go` в том же пакете). SHOULD:
  `Describe`(юнит) / `Context`(сценарий) / `It`(утверждение), `DescribeTable` для табличных кейсов.
- **D-10:** Тестирование use case'ов, зависящих от портов (`UnitOfWork`, repositories) — через
  **генерируемые моки (кодоген)**. Канонический инструмент — **mockery** (`vektra/mockery`).
  `testing.md` показывает, как мокать порты mockery + проверять через Gomega. Обвязка
  (`go:generate`/`.mockery.yaml`/установка инструмента) — деталь enforcement-тулинга Phase 4
  (помечается planned, см. D-11); сам выбор инструмента закреплён здесь.

### Сквозное — enforcement-статус (готовит ENF-05 Phase 4)
- **D-11:** Механизируемые правила во **всех** доках Phase 3 получают **пред-пометку
  enforcement-статуса** уже сейчас (напр. `convention-only` / `planned: CI-gated Phase 4` /
  `planned: hook Phase 4`), согласно authoring-стандарту. Phase 4 (ENF-05) затем только **меняет
  статус** на фактический, а не ретрофитит пометки с нуля. Примеры будущего enforcement: формат →
  gofumpt/gci; направление импортов → линтер (depguard и т.п.); генерация моков → `go:generate`.

### Claude's Discretion
- Конкретное имя нейтрального плейсхолдера-агрегата (D-05) — на усмотрение планировщика
  (важно лишь: вне домена gwall-e и одинаковое во всех доках).
- Точная форма диаграммы импортов (D-06): ASCII-схема или таблица — на усмотрение планировщика.
- Разбивка любого дока на под-файлы при превышении ~150–200 строк (authoring.md) — по факту.
- Конкретные формулировки/набор мини-примеров «плохо/хорошо» сверх перечисленных правил.
- Точный стиль генерируемых mockery-моков (expecter vs классический) и как они стыкуются с
  Gomega-ассертами — на усмотрение планировщика/ресёрча в рамках D-10.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Источник истины по скоупу и решениям
- `.planning/PROJECT.md` — Key Decisions (DDD+гексагон БЕЗ CQRS; write-side `1 use case = 1 struct + Execute`;
  read-side query-lite в DTO; порт `UnitOfWork`; transactional outbox + relay; язык кода RU-комментарии/EN-имена;
  `inventory` вне `go.work`; снос `internal/`/`pkg/mediatr`/`tx.go`), Context (целевая раскладка слоёв
  `domain/usecases/query/repositories/api/cron`, `app` = composition root, `cmd` = main), Constraints.
- `.planning/REQUIREMENTS.md` — DOC-03/DOC-04/DOC-05/PAT-01 (скоуп Phase 3); DOC-07 (glossary) отложено в v2;
  карта v1/v2 (что НЕ документировать сейчас); Out of Scope (reference-service walkthrough — пока `inventory` не собирается).
- `.planning/ROADMAP.md` — Phase 3 goal и success criteria (style/testing/architecture/patterns).

### Стандарт и индекс базы знаний (созданы Phase 1–2 — СЛЕДОВАТЬ им)
- `knowledge/authoring.md` — authoring-стандарт: MUST/SHOULD/WON'T, парность «запрет → do»,
  pointer-over-copy, размер/дробление (~150–200 строк), no-phantom, **статус enforcement**.
  Все доки Phase 3 MUST следовать.
- `knowledge/boundaries.md` — do-not правила (no-phantom: не документировать несуществующее как
  работающее; WIP-сервисы не эталон) + **карта владения фактами** (язык кода → `style.md`).
  Phase 3 доки регистрируются в этой карте.
- `knowledge/README.md` — индекс/порядок чтения + блок «что где живёт». Обновить при наполнении
  каждого дока (без битых ссылок).
- `knowledge/structure.md` — раскладка `go.work` (модули в/вне workspace, `inventory` WIP вне workspace).
- `knowledge/build.md` — команды сборки/тестов (`cd pkg && go test`, `GOWORK=off` для `inventory`).
- `AGENTS.md` / `CLAUDE.md` (корень) — тонкие точки входа; ссылаются на `knowledge/`.

### Прецедент (как писались доки в этой базе)
- `.planning/phases/01-knowledge-base-layout/01-CONTEXT.md` — D-01..D-08 Phase 1 (раскладка, имена, язык,
  источник истины AGENTS.md, риск регенерации CLAUDE.md).
- `.planning/phases/02-foundation-docs/02-CONTEXT.md` — D-01..D-08 Phase 2 (декскоуп glossary,
  модульный structure.md, карта владения фактами, инфра/процесс-уровень).

### Research
- `.planning/research/ARCHITECTURE.md`, `.planning/research/FEATURES.md`, `.planning/research/PITFALLS.md`,
  `.planning/research/STACK.md`, `.planning/research/SUMMARY.md` — структура базы знаний, table-stakes доки,
  грабли (Pitfall 2 мега-файл/контекст-бюджет, Pitfall 5 MUST/SHOULD/WON'T+«do», Pitfall 8 phantom-правила),
  стек.

### Эталон кода (реальный, для testing.md)
- `pkg/http/*_test.go` — **реальные** Ginkgo v2 + Gomega тесты: `RegisterFailHandler(Fail)` +
  `RunSpecs(t, "HTTP Package Suite")`, suite-бутстрап. Единственный компилируемый тест-эталон в репо.
- `pkg/go.mod` — фиксирует `github.com/onsi/ginkgo/v2 v2.23.4`, `github.com/onsi/gomega v1.38.0`.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `knowledge/README.md`, `knowledge/authoring.md`, `knowledge/boundaries.md` — существуют, задают
  формат/индекс/карту владения. Наполнение Phase 3 встраивается в индекс (снять «запланировано»,
  добавить ссылки) и регистрируется в карте владения фактами.
- `pkg/http/*_test.go` — живой Ginkgo+Gomega эталон для `testing.md` (suite-бутстрап, импорт
  dot-import `. "github.com/onsi/ginkgo/v2"` / `. "...gomega"`).

### Established Patterns
- Progressive disclosure: тонкие входы (`AGENTS.md`/`CLAUDE.md`) → topic-доки `knowledge/*.md`.
- Authoring-стандарт (MUST/SHOULD/WON'T + «do», pointer-over-copy, статус enforcement) — обязателен.
- Pointer-over-copy / карта владения: один факт = один канон; язык кода канонизируется в `style.md`.

### Integration Points
- `knowledge/README.md` индекс + порядок чтения — обновить при появлении каждого дока.
- `boundaries.md` карта владения — добавить строки: язык/стиль → `style.md`, тесты → `testing.md`,
  архитектура слоёв → `architecture.md`, рецепты → `patterns.md`.
- `architecture.md` ↔ `patterns.md` — взаимные ссылки (правила ↔ рецепты, D-04).
- `style.md`/`testing.md`/`architecture.md` enforcement-пометки (D-11) → вход для Phase 4 (ENF-05).

### Risks / Constraints for planner
- ⚠️ **Нет компилируемого эталонного сервиса:** слои `domain/usecases/query/repositories/api/cron`
  не существуют как код (`inventory/internal/` — пустые stale-каталоги). Примеры в
  `architecture.md`/`patterns.md` — иллюстративные, помеченные (D-01), на нейтральном
  плейсхолдере (D-05). НЕ выдавать за существующий код (no-phantom, boundaries.md).
- ⚠️ **mockery как новый инструмент (D-10):** в репо его сейчас нет; `testing.md` фиксирует
  конвенцию, реальная установка/`go:generate` — Phase 4. Не документировать как «уже настроено».
- ⚠️ **Контекст-бюджет (Pitfall 2):** держать каждый док в ~150–200 строк; при росте — дробить
  (authoring.md). `architecture.md` рискует разрастись — следить за объёмом.
- ⚠️ **Кросс-док дубли:** язык кода — только в `style.md`; правила архитектуры — только в
  `architecture.md` (patterns ссылается). Команды сборки/тестов — в `build.md` (testing ссылается).

</code_context>

<specifics>
## Specific Ideas

- Пользователь выбрал **mockery** (`vektra/mockery`) как канонический генератор моков — не из
  предложенных mockgen/counterfeiter. Downstream: `testing.md` ориентируется именно на mockery,
  ресёрч уточняет идиоматичную стыковку mockery-моков с Gomega-ассертами под гексагон-порты.
- Принцип сессии (повтор из Phase 1–2): фундамент правил для ИИ/команды, **не** описание системы —
  поэтому примеры на нейтральном плейсхолдере, без доменных терминов (D-05).
- Enforcement-пометки писать «вперёд» (planned: Phase 4), чтобы Phase 4 только переключал статус (D-11).

</specifics>

<deferred>
## Deferred Ideas

- **`glossary.md` (DOC-07) → domain-milestone** (D-01 Phase 2): доменный ubiquitous language
  (host/VM/owner/SRE/ITDC/namespace/project + роли/«согласованность»). Не в этом milestone.
- **Reference-service walkthrough** (`reference-service.md`) — отложен (Out of Scope REQUIREMENTS.md):
  разблокируется, когда `inventory` начнёт компилироваться. Тогда иллюстративные сниппеты Phase 3
  можно будет заменить/дополнить ссылками на реальный сервис.
- **Реальная настройка тулинга** (golangci-lint/gofumpt/gci, lefthook, commitlint, buf, `go:generate`
  для mockery) и проставление фактического enforcement-статуса (ENF-05) — **Phase 4**.
- ADR-доки, `anti-patterns.md`, `libraries.md`, onboarding, maintenance-протокол — v2 (REQUIREMENTS.md).

</deferred>

---

*Phase: 3-Доки конвенций и архитектуры*
*Context gathered: 2026-06-17*
