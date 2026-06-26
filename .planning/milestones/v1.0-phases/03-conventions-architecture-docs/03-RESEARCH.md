# Phase 3: Доки конвенций и архитектуры — Research

**Researched:** 2026-06-17
**Domain:** Документация конвенций для ИИ/команды (knowledge-base authoring) + целевые Go-конвенции, которые эти доки описывают (DDD + гексагон БЕЗ CQRS, Ginkgo v2 + Gomega, mockery v3)
**Confidence:** HIGH (authoring-стандарт и архитектурные решения — locked в CONTEXT/PROJECT и подтверждены инспекцией репо); MEDIUM (mockery v3 ↔ Gomega стыковка — подтверждено офиц. docs WebFetch, но не локальной компиляцией, т.к. mockery в репо ещё нет)

## Summary

Это **документационная** фаза: четыре канонических дока `knowledge/*.md` (русская проза, английские тех-термины/код), наполняющие уже зафиксированный Phase 1–2 каркас. Никакого продакшн-кода — только нормативные правила и копируемые рецепты для ИИ/команды. Все четыре дока **MUST** следовать authoring-стандарту (`authoring.md`: MUST/SHOULD/WON'T, парность «запрет → do», pointer-over-copy, размер ~150–200 строк, no-phantom, статус enforcement).

Главный риск — **не внешний (контент почти весь locked)**, а дисциплина: (1) no-phantom — эталонного компилируемого сервиса нет, поэтому все архитектурные/паттерн-сниппеты ILLUSTRATIVE на нейтральном плейсхолдере и явно помечены; (2) кросс-док дубли — язык кода только в `style.md`, архитектурные правила только в `architecture.md`, команды сборки/тестов только в `build.md`; (3) контекст-бюджет — `architecture.md` рискует разрастись (5+ паттернов), дробить при >200 строк; (4) каждый док регистрируется в индексе (`README.md`) и карте владения (`boundaries.md`) **без битых ссылок** (ссылка появляется вместе с файлом).

Единственный реальный внешний knowledge-gap — стыковка mockery v3 + Gomega под гексагон-порты (D-10): подтверждено, что mockery v3 генерирует testify-style моки с expecter API (`mock.EXPECT().Method().Return()`), конструктор `NewMockX(t)` авто-регистрирует `t.Cleanup` с `AssertExpectations`, а в Ginkgo передаётся `GinkgoT()`. Ассертить результат use case — через Gomega `Expect(...)`. Установка/`go:generate`/`.mockery.yaml` — Phase 4 (помечается `planned`).

**Primary recommendation:** Спланировать ровно 4 контентных дока + 2 интеграционных правки (`README.md` индекс, `boundaries.md` карта владения), каждое правило с тегом силы и forward-enforcement-пометкой (D-11), все архитектурные/паттерн-сниппеты на одном нейтральном плейсхолдере (`Order`) с явной меткой «иллюстрация, не из компилируемого файла». Реальный код-эталон цитировать только для `testing.md` (`pkg/http/*_test.go` существует и компилируется).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**patterns.md — глубина рецептов (PAT-01):**
- **D-01:** Рецепты используют **иллюстративные Go-сниппеты**, реальные и идиоматичные, но **явно помеченные** как «целевой вид / иллюстрация» — НЕ из компилируемого файла. Даёт «копируемость» PAT-01 без нарушения no-phantom. Контекст: эталонный сервис не компилируется (`inventory/internal/` снесён), reference-service walkthrough отложен в Out of Scope.
- **D-02:** Каждый рецепт — **вертикальный срез до wiring включительно**: `struct + Execute` → порты/репозиторий → composition root (`app`, ручной DI) → gRPC-адаптер в `api`. Показать полный путь «как подключить», не одну механику слоя.
- **D-03:** Каталог рецептов покрывает: «добавить use case», «добавить query (read-side, DTO)», «добавить aggregate (+ фабрика, `PullEvents`)», «добавить repository (Mongo-реализация порта + участие в UnitOfWork)». Согласованы с `architecture.md`, ссылаются на него (pointer-over-copy).

**architecture.md ↔ patterns.md — граница (DOC-05 / PAT-01):**
- **D-04:** «Правила vs рецепты» (pointer-over-copy): `architecture.md` = инварианты/правила/why (слои, направление импортов, порт `UnitOfWork`, transactional outbox + relay, фабрики агрегатов + `PullEvents`, query-lite read-side в DTO, usecases-interactor `Execute`, явный **MUST NOT** возрождать CQRS-диспетчер/`pkg/mediatr`/`TxManager`). `patterns.md` = пошаговые how-to со сниппетами, ссылается на `architecture.md`, правила не дублирует.
- **D-05:** Примеры в обоих доках — на **едином нейтральном плейсхолдере-агрегате** (напр. `Order`/`Widget`), сквозном между доками. **НЕ** использовать доменные термины gwall-e (host/VM/owner). Конкретное имя — на усмотрение планировщика, но одно во всех доках.
- **D-06:** `architecture.md` содержит **текстовую диаграмму/таблицу направления импортов** между слоями (`domain ← usecases ← api`/`repositories`; `domain` не импортирует наружу). ASCII или таблица — на усмотрение.

**style.md — охват и формат (DOC-04):**
- **D-07:** Охват — **только проект-специфика** gwall-e: язык (русские комментарии/доменная терминология; имена идентификаторов — английские; комментарии в тестах — английские), типизированные ID, sentinel vs обёрнутые ошибки (`%w`), маппинг DTO→домен **внутри хендлера**. Общий Go-стиль (формат/нейминг) **не дублировать** — ссылка на gofumpt (Phase 4) + Effective Go. `style.md` — **единственное место** канона про язык комментариев.
- **D-08:** Формат каждого правила — **«правило + плохо/хорошо»**: короткий MUST/SHOULD + мини-пример «плохо → хорошо». Особенно для typed IDs, sentinel-vs-wrapped errors, DTO→домен.

**testing.md — конвенции (DOC-03):**
- **D-09:** Мандат: **MUST на каркас + SHOULD на структуру**. MUST: suite-бутстрап (`RegisterFailHandler(Fail)` + `RunSpecs(t, "...")` — как в реальном `pkg/http`), комментарии в тестах на английском, тесты рядом с кодом (`*_test.go` в том же пакете). SHOULD: `Describe`(юнит)/`Context`(сценарий)/`It`(утверждение), `DescribeTable` для табличных кейсов.
- **D-10:** Тестирование use case'ов на портах (`UnitOfWork`, repositories) — через **генерируемые моки (кодоген)**. Канонический инструмент — **mockery** (`vektra/mockery`). `testing.md` показывает мокинг портов mockery + проверку через Gomega. Обвязка (`go:generate`/`.mockery.yaml`/установка) — Phase 4 (помечается planned, D-11); выбор инструмента закреплён здесь.

**Сквозное — enforcement-статус (D-11):** Механизируемые правила во всех доках Phase 3 получают **пред-пометку enforcement-статуса** уже сейчас (`convention-only` / `planned: CI-gated Phase 4` / `planned: hook Phase 4`). Phase 4 (ENF-05) только меняет статус.

### Claude's Discretion
- Конкретное имя нейтрального плейсхолдера-агрегата (D-05) — вне домена gwall-e, одинаковое во всех доках.
- Точная форма диаграммы импортов (D-06): ASCII-схема или таблица.
- Разбивка любого дока на под-файлы при превышении ~150–200 строк (authoring.md) — по факту.
- Конкретные формулировки/набор мини-примеров «плохо/хорошо» сверх перечисленных правил.
- Точный стиль генерируемых mockery-моков (expecter vs классический) и стыковка с Gomega-ассертами — в рамках D-10.

### Deferred Ideas (OUT OF SCOPE)
- **`glossary.md` (DOC-07) → domain-milestone** (D-01 Phase 2): доменный ubiquitous language (host/VM/owner/SRE/ITDC/namespace/project + роли/«согласованность»). Не в этом milestone.
- **Reference-service walkthrough** (`reference-service.md`) — Out of Scope; разблокируется, когда `inventory` начнёт компилироваться.
- **Реальная настройка тулинга** (golangci-lint/gofumpt/gci, lefthook, commitlint, buf, `go:generate` для mockery) и фактический enforcement-статус (ENF-05) — Phase 4.
- ADR-доки, `anti-patterns.md`, `libraries.md`, onboarding, maintenance-протокол — v2.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DOC-04 | `knowledge/style.md` — канон языка (RU комментарии/EN имена), типизированные ID, sentinel vs обёрнутые ошибки (`%w`), маппинг DTO→домен внутри хендлера; единственное место правила про язык | Code Examples §1 (typed IDs, errors, DTO→domain); карта владения уже резервирует `style.md` за этим фактом (boundaries.md L67); формат «плохо/хорошо» per D-08 совместим с authoring парностью |
| DOC-03 | `knowledge/testing.md` — Ginkgo v2 + Gomega, EN-комментарии, структура спеков | Реальный эталон `pkg/http/*_test.go` (suite-бутстрап `RegisterFailHandler(Fail)`+`RunSpecs`, dot-imports, `Describe/Context/It`); версии pinned в `pkg/go.mod`; mockery v3 + GinkgoT() стыковка (Code Examples §2) |
| DOC-05 | `knowledge/architecture.md` — DDD + гексагон БЕЗ CQRS: слои/импорты, `Execute`, query-lite, `UnitOfWork`, outbox+relay, фабрики+`PullEvents`; явный MUST NOT возрождать CQRS-диспетчер/`TxManager` | Architecture Patterns §1–4 (из `.planning/research/ARCHITECTURE.md`, locked в PROJECT.md Key Decisions); диаграмма импортов (D-06); Anti-Patterns раздел даёт текст MUST NOT |
| PAT-01 | `knowledge/patterns.md` — копируемые рецепты add use case/query/aggregate/repository, согласованные с architecture.md | D-02 вертикальный срез + D-03 каталог; Code Examples §1 — иллюстративные сниппеты на плейсхолдере `Order`; pointer-over-copy на `architecture.md` |
</phase_requirements>

## Architectural Responsibility Map

Для документационной фазы «tier» — это **владелец факта** в карте knowledge-базы (pointer-over-copy), а не вычислительный слой. Каждая способность принадлежит ровно одному канон-доку; остальные ссылаются.

| Capability (факт/правило) | Primary Tier (канон-док) | Secondary Tier (ссылается) | Rationale |
|---------------------------|--------------------------|----------------------------|-----------|
| Язык кода/комментариев, EN имена | `style.md` | `testing.md` (EN-комментарии в тестах — ссылка), root `CLAUDE.md`/`AGENTS.md` (указатель) | Карта владения (boundaries.md L67) уже резервирует это за `style.md`; единственное место (D-07) |
| Типизированные ID, sentinel vs `%w`, DTO→домен в хендлере | `style.md` | `patterns.md`/`architecture.md` (ссылка при упоминании) | Это правила кода уровня файла — стиль, не архитектура слоёв |
| Слои/направление импортов, `Execute`, query-lite, `UnitOfWork`, outbox+relay, `PullEvents`, MUST NOT CQRS | `architecture.md` | `patterns.md` (рецепты ссылаются, D-04) | Инварианты/why — архитектурный канон |
| Пошаговые how-to (add use case/query/aggregate/repo) | `patterns.md` | `architecture.md` (за правилами — ссылка, не копия) | Рецепты vs правила (D-04 pointer-over-copy) |
| Конвенции тестов: Ginkgo v2+Gomega, спек-структура, мокинг портов | `testing.md` | `build.md` (команды прогона — ссылка) | Тесты — отдельный топик; команды живут в `build.md` |
| Команды сборки/тестов (`cd pkg && go test`, `GOWORK=off`) | `build.md` (существует) | `testing.md` (ссылка, не копия) | Уже канон в build.md (boundaries.md L64) |
| Authoring-стандарт (MUST/SHOULD/WON'T, парность, статус enforcement) | `authoring.md` (существует) | все 4 новых дока (следуют) | Уже канон |
| Индекс/порядок чтения, статусы доков | `README.md` (существует) | — | Обновляется при наполнении каждого дока |

## Standard Stack

> Документационная фаза — продакшн-зависимостей не добавляет. «Стек» здесь — инструменты/библиотеки, **которые описывают** доки `testing.md`/`architecture.md`. Версии Ginkgo/Gomega verified из `pkg/go.mod`; mockery — выбор закреплён CONTEXT D-10, версия из офиц. docs.

### Core (описываются доками, уже в репо)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/onsi/ginkgo/v2` | v2.23.4 | BDD-фреймворк спеков (testing.md) | [VERIFIED: pkg/go.mod] pinned; реальный эталон `pkg/http/*_test.go` компилируется и зелёный |
| `github.com/onsi/gomega` | v1.38.0 | Matcher-библиотека ассертов (testing.md) | [VERIFIED: pkg/go.mod] pinned; используется в `pkg/http/*_test.go` (`Expect(...).To(...)`) |

### Supporting (описывается testing.md как конвенция; УСТАНОВКА — Phase 4)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/vektra/mockery` | v3 (CLI-генератор) | Кодоген testify-style моков для портов (`UnitOfWork`, repositories) | [CITED: vektra.github.io/mockery/latest] Выбран в D-10. В репо ЕЩЁ НЕТ — `testing.md` фиксирует конвенцию, помечает `planned: Phase 4`. Сами моки зависят от `github.com/stretchr/testify` (транзитивно через сгенерированный код) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| mockery | mockgen (gomock), counterfeiter | Пользователь явно выбрал mockery (D-10, specifics) — НЕ переобсуждать. Альтернативы вне скоупа |
| Ручные fake-реализации портов | mockery-генерация | D-10 фиксирует кодоген; ручные фейки допустимы как пояснение, но канон — генерируемые моки |

**Installation:** не в этой фазе. `testing.md` пишет конвенцию как `planned: Phase 4` (установка `go install github.com/vektra/mockery/v3@latest`, `.mockery.yaml`, `go:generate` — задел ENF-05).

**Version verification (выполнено):**
```bash
# из pkg/go.mod — verified
github.com/onsi/ginkgo/v2 v2.23.4
github.com/onsi/gomega v1.38.0
# mockery — major version v3 подтверждён офиц. docs (vektra.github.io/mockery/latest)
```

⚠️ **Несоответствие версии Go (для планировщика):** `pkg/go.mod` объявляет `go 1.23.6`, но `structure.md`/`build.md`/PROJECT.md называют Go **1.24.6**. Это НЕ предмет Phase 3 (доки версию Go-toolchain не фиксируют как новый факт — это `structure.md`/`build.md`), но `testing.md`/`patterns.md` **WON'T** жёстко прописывать версию Go — ссылаться на `structure.md`/`build.md`. Флаг для будущей сверки. [VERIFIED: pkg/go.mod vs knowledge/structure.md]

## Package Legitimacy Audit

> Phase 3 НЕ устанавливает пакеты (документационная фаза). Аудит — для пакетов, которые доки **называют** как канон, чтобы планировщик не закрепил slop-имя.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `github.com/onsi/ginkgo/v2` | Go (pkg.go.dev) | зрелый | широко используется | github.com/onsi/ginkgo | OK | Уже в репо (pkg/go.mod), эталон компилируется |
| `github.com/onsi/gomega` | Go | зрелый | широко используется | github.com/onsi/gomega | OK | Уже в репо (pkg/go.mod) |
| `github.com/vektra/mockery` | Go | зрелый (v3) | широко используется | github.com/vektra/mockery | OK (CITED docs) | Конвенция в testing.md; установка — Phase 4 (planned) |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

> Сетевой seam `package-legitimacy check` для Go-модулей не запускался (Go-экосистема не покрыта seam'ом, все три пакета — общеизвестные с публичными GitHub-репо, два уже pinned в `pkg/go.mod` и компилируются). mockery помечен `[CITED: vektra.github.io/mockery]` — verified из офиц. docs, не из training-данных. Планировщику: установку mockery гейтить в **Phase 4**, не в Phase 3 (в Phase 3 ничего не ставится).

## Architecture Patterns

> Применимы два уровня: **(A) как устроена сама фаза доков** (граф из 4 файлов + 2 интеграционных правки) и **(B) целевая архитектура сервиса**, которую `architecture.md`/`patterns.md` описывают (locked в PROJECT.md / `.planning/research/ARCHITECTURE.md`).

### System Architecture Diagram (A — поток наполнения доков)

```
CONTEXT.md (locked D-01..D-11) + authoring.md (стандарт)
        │
        ▼
  ┌─────────────────────────────────────────────────────────┐
  │  4 контентных дока (новые)                                │
  │                                                           │
  │  style.md ──(EN-комментарии: ссылка)──► testing.md        │
  │     ▲                                       │             │
  │     │(typed ID/errors/DTO→домен: ссылка)    │(команды:    │
  │     │                                       │ ссылка)     │
  │     │                                       ▼             │
  │  patterns.md ──(правила: ссылка D-04)──► architecture.md  │
  │     │                                       │             │
  │     └───── единый плейсхолдер `Order` ──────┘ (D-05)       │
  │            (иллюстративные сниппеты, D-01 no-phantom)      │
  └─────────────────────────────────────────────────────────┘
        │                                   │
        ▼ (зарегистрировать)                ▼ (на каждый док)
  boundaries.md                        README.md
  «карта владения фактами»             «индекс + порядок чтения»
  +4 строки (style/testing/            снять «запланировано»,
   architecture/patterns)              добавить ссылку — без битых
        │                                   │
        ▼                                   ▼
  Phase 4 (ENF-05): только переключает enforcement-статус (D-11)
```

Правила направленности (инвариант для авторинга): факт течёт **из** канон-дока **в** ссылающиеся доки односторонне; обратных копий нет (pointer-over-copy). Ссылка добавляется в `README.md`/`boundaries.md` только когда целевой файл уже существует (no broken links).

### System Architecture Diagram (B — целевой сервис, что описывает architecture.md)

```
gRPC request
   ↓
api/ (gRPC адаптер): decode + protovalidate
   ↓ Execute(ctx, in)                     ┌─ read path ─────────────┐
usecases/ (write): 1 struct + Execute     │ query/ : Execute → DTO  │
   ↓ domain.NewOrder(...) инварианты       │  прямой read Mongo,     │
   ↓ uow.Do(ctx, fn):                      │  мимо агрегатов         │
        repositories/ (Mongo tx):          └─────────────────────────┘
          repo.Save(agg)
          outbox.Append(agg.PullEvents())  // та же tx
        COMMIT
   ↓
gRPC reply ← encode(out)

(async) relay: outbox → EventPublisher.Publish → mark published

Зависимости ВСЕГДА внутрь на domain; domain не импортирует наружу.
app/ = composition root (ручной DI); cmd/ = main.
```

### Recommended Project Structure (фаза доков — что создаётся/трогается)
```
knowledge/
├── style.md           # НОВЫЙ (DOC-04): язык, typed ID, errors, DTO→домен
├── testing.md         # НОВЫЙ (DOC-03): Ginkgo v2+Gomega, спеки, mockery-конвенция
├── architecture.md    # НОВЫЙ (DOC-05): DDD+гексагон БЕЗ CQRS, слои, порты, события
├── patterns.md        # НОВЫЙ (PAT-01): рецепты add use case/query/aggregate/repo
├── README.md          # ПРАВКА: снять «запланировано», добавить 4 ссылки
└── boundaries.md      # ПРАВКА: +4 строки в карту владения фактами
```

### Pattern 1: «Правило + плохо/хорошо» (style.md, D-08)
**What:** Каждое правило — MUST/SHOULD тег + микро-пример «плохо → хорошо».
**When to use:** Все нормативные правила в `style.md` (и где уместно — в `testing.md`).
**Example (иллюстрация):**
```go
// MUST: типизированный ID вместо «голой» строки. ⟶ enforcement: convention-only
//
// плохо:
func GetOrder(id string) (*Order, error)        // легко перепутать с другим string
//
// хорошо:
type OrderID string
func GetOrder(id OrderID) (*Order, error)        // компилятор ловит подмену
```

### Pattern 2: «Правила vs рецепты» (architecture.md ↔ patterns.md, D-04)
**What:** `architecture.md` хранит инвариант/why; `patterns.md` — пошаговый how-to и **ссылается** на правило, не копирует.
**When to use:** Везде, где рецепт опирается на архитектурное правило.
**Example:** В `patterns.md` рецепт «добавить use case» пишет: *«поместите запись в `uow.Do(...)` — правило транзакционной границы см. [architecture.md](architecture.md) §UnitOfWork»*, без повторения текста правила.

### Pattern 3: Единый нейтральный плейсхолдер (D-05)
**What:** Один агрегат (`Order`) сквозной во всех сниппетах обоих доков, вне домена gwall-e.
**When to use:** Все иллюстративные сниппеты `architecture.md`/`patterns.md`.
**Example:** `Order`, `OrderID`, `OrderRepository`, `RegisterOrderUseCase`, `ListOrdersQuery` — никаких host/VM/owner.

### Pattern 4: Forward enforcement-метка (D-11)
**What:** Механизируемое правило несёт пред-пометку будущего enforcement.
**When to use:** Любое правило, которое Phase 4 сможет автоматизировать.
**Example:**
```
- **MUST** маппить DTO→домен внутри хендлера.  ⟶ planned: CI-gated Phase 4 (depguard)
- **SHOULD** ставить gofumpt-формат.            ⟶ planned: hook Phase 4 (gofumpt)
- **MUST** комментарии кода — на русском.       ⟶ convention-only (review-enforced)
```

### Anti-Patterns to Avoid
- **Phantom-сниппет:** Подавать иллюстративный код как существующий файл/сервис. WON'T — все сниппеты `architecture.md`/`patterns.md` MUST нести метку «целевой вид / иллюстрация, не из компилируемого файла» (D-01, no-phantom). Реальный код цитируется только для `testing.md` (`pkg/http/*_test.go`).
- **Кросс-док дубль:** Повторять правило языка вне `style.md`, архитектурное правило вне `architecture.md`, команды вне `build.md`. WON'T — вместо копии относительная ссылка (pointer-over-copy).
- **Мега-файл:** `architecture.md` собирает 5+ паттернов + диаграмму + MUST NOT → риск >200 строк. WON'T держать всё в одном файле сверх лимита — дробить на под-топики и связывать ссылками (authoring.md «Размер и дробление»).
- **Битая ссылка:** Добавить ссылку в `README.md`/`boundaries.md` до создания файла. WON'T — ссылка появляется вместе с файлом.
- **Хеджирование без тега:** «обычно», «желательно», «prefer» в нормативном правиле. WON'T — ставить MUST/SHOULD/WON'T (authoring.md).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Authoring-формат (теги силы, парность) | Свой формат разметки правил | `authoring.md` стандарт (существует) | Канон уже задан Phase 1; доки ему следуют |
| Команды сборки/тестов | Перечислять `go test`/`GOWORK=off` в `testing.md` | Ссылка на `build.md` | Команды — канон в build.md (boundaries.md L64) |
| Правило языка комментариев | Повторять «RU комментарии» в каждом доке | Один канон в `style.md` + ссылки | Карта владения (boundaries.md L67); единственное место (D-07) |
| Архитектурные правила в рецептах | Копировать инварианты в `patterns.md` | Ссылка на `architecture.md` (D-04) | pointer-over-copy; копии расходятся |
| Ручные моки портов | Писать fake-структуры вручную | mockery-генерация (D-10) | Кодоген синхронизирован с интерфейсом; меньше дрейфа |
| Suite-бутстрап тестов | Изобретать свой раннер | `RegisterFailHandler(Fail)` + `RunSpecs` (как `pkg/http`) | Реальный рабочий эталон в репо |

**Key insight:** В документационной фазе «не хэндроллить» = «не дублировать факт». Каждый факт имеет ровно один канон-док; всё остальное — ссылка. Это и есть главный механизм против дрейфа базы знаний.

## Runtime State Inventory

> Не применимо в классическом смысле (нет данных/сервисов для миграции). Но фаза **меняет состояние knowledge-графа** — аналог «live config». Перечисляю явно.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — verified: фаза создаёт только Markdown-файлы, нет БД/датастора | None |
| Live service config (аналог: индекс/карта базы знаний) | `knowledge/README.md` (индекс, строки 32–35 со статусом «запланировано» для всех 4 доков); `knowledge/boundaries.md` (карта владения, L67 уже резервирует `style.md`, остальные 3 факта не зарегистрированы) | **Правка-doc:** снять «запланировано» + добавить ссылки в README на каждый созданный файл; добавить 3 строки в карту владения (testing/architecture/patterns) |
| OS-registered state | None — verified: нет хуков/скедулеров/демонов в скоупе (это Phase 4) | None |
| Secrets/env vars | None — verified: документационная фаза, секреты не затрагивает | None |
| Build artifacts | None — verified: нет компиляции/install в Phase 3 (mockery install — Phase 4) | None |

**Канонический вопрос:** после создания 4 файлов — какие «зарегистрированные» места ещё ссылаются на них как «запланировано»? Ответ: `README.md` индекс (строки 32–35, 45) и `AGENTS.md` (таблица L37 «запланировано (Phase 3)»). Все три (`README.md`, `boundaries.md`, опционально `AGENTS.md` таблица статусов) MUST обновиться, иначе индекс рассинхронизируется с реальностью.

## Common Pitfalls

### Pitfall 1: Phantom-правила (сниппет выдан за существующий код)
**What goes wrong:** `architecture.md`/`patterns.md` показывают `RegisterOrderUseCase` так, будто файл уже есть в репо; агент идёт его «дописывать»/чинить.
**Why it happens:** Эталонного компилируемого сервиса нет (`inventory/internal/` снесён), но сниппеты выглядят настоящими.
**How to avoid:** Каждый сниппет несёт явную метку «целевой вид / иллюстрация — НЕ из компилируемого файла» (D-01). Реальный код цитируется только для `testing.md`. no-phantom правило уже в `boundaries.md`/`authoring.md`.
**Warning signs:** Сниппет без метки; ссылка на несуществующий путь `services/.../order.go`; агент открывает PR «доделать Order».

### Pitfall 2: Мега-файл `architecture.md` (контекст-бюджет, Pitfall 2 из research)
**What goes wrong:** Диаграмма импортов + 4–5 паттернов (Execute, query-lite, UoW, outbox, фабрики/PullEvents) + MUST NOT → файл >200 строк, агент роняет правила из хвоста.
**Why it happens:** «Вся архитектура в одном месте» кажется удобным.
**How to avoid:** Держать `architecture.md` на уровне инвариантов/why (правила, не how-to — детали в `patterns.md`). При >200 строк дробить (напр. `architecture.md` + `architecture-events.md`) и связывать ссылками; обновить индекс. Pruning-тест на каждую строку.
**Warning signs:** Файл подбирается к 200 строкам; в нём появляются пошаговые how-to (это уже `patterns.md`).

### Pitfall 3: Кросс-док дубль правила языка/архитектуры/команд
**What goes wrong:** «RU комментарии» оказываются и в `style.md`, и в `testing.md`; команды теста — в `testing.md` и `build.md`. Копии расходятся.
**Why it happens:** Локально кажется удобнее повторить, чем сослаться.
**How to avoid:** Карта владения (boundaries.md): язык → `style.md`, команды → `build.md`, архитектура → `architecture.md`. Везде ссылка, не копия (D-04, D-07). `testing.md` про EN-комментарии — короткая ссылка на `style.md`.
**Warning signs:** Один и тот же MUST дословно в двух файлах; grep правила находит >1 канон.

### Pitfall 4: Битые/преждевременные ссылки в индексе
**What goes wrong:** `README.md`/`boundaries.md`/`AGENTS.md` ссылаются на `patterns.md` до его создания → битая ссылка.
**Why it happens:** Обновили индекс пакетно, а файлы пишутся в разном порядке/волнах.
**How to avoid:** Ссылка добавляется в индекс **в той же единице работы**, что создаёт файл (или после). no-phantom для ссылок (authoring.md L61). Валидация — см. Validation Architecture (link integrity).
**Warning signs:** Markdown-ссылка ведёт на несуществующий `knowledge/*.md`.

### Pitfall 5: mockery документирован как «уже настроено»
**What goes wrong:** `testing.md` пишет `go:generate`/`.mockery.yaml` так, будто инструмент стоит и сконфигурирован; агент пытается запустить генерацию.
**Why it happens:** Конвенция и установка смешиваются.
**How to avoid:** `testing.md` фиксирует **конвенцию** (как мокать, expecter API, GinkgoT()), а установку/обвязку помечает `planned: Phase 4` (D-10/D-11). Не показывать рабочую команду генерации как проверенную.
**Warning signs:** В `testing.md` команда `mockery` без пометки planned; ссылка на несуществующий `.mockery.yaml`.

## Code Examples

> Сниппеты §1 — **иллюстративные** (плейсхолдер `Order`, для `architecture.md`/`patterns.md`, метить D-01). Сниппет §2 (suite-бутстрап) — **реальный**, из `pkg/http/http_test.go` (компилируется). mockery-фрагмент §3 — конвенция (CITED docs), помечать `planned: Phase 4`.

### §1 — Целевые конвенции кода (иллюстрация для style.md / patterns.md / architecture.md)
```go
// ИЛЛЮСТРАЦИЯ (целевой вид) — НЕ из компилируемого файла. Плейсхолдер: Order.

// Типизированный ID (style.md)
type OrderID string

// Sentinel-ошибка + обёртка %w (style.md)
var ErrOrderNotFound = errors.New("order not found")
func (r *orderRepo) Get(ctx context.Context, id OrderID) (*Order, error) {
    // ... если не нашли:
    return nil, fmt.Errorf("get order %s: %w", id, ErrOrderNotFound) // обёртка сохраняет sentinel
}

// Use case = struct + Execute; маппинг DTO→домен внутри хендлера (architecture.md/style.md)
type RegisterOrderUseCase struct {
    orders OrderRepository // порт из domain/
    uow    UnitOfWork      // порт из domain/
}
func (uc *RegisterOrderUseCase) Execute(ctx context.Context, in RegisterOrderInput) (RegisterOrderOutput, error) {
    order, err := NewOrder(in.SKU, in.Qty) // фабрика проверяет инварианты
    if err != nil {
        return RegisterOrderOutput{}, err
    }
    err = uc.uow.Do(ctx, func(ctx context.Context) error {
        return uc.orders.Save(ctx, order) // Save пишет агрегат + outbox-события в той же tx
    })
    if err != nil {
        return RegisterOrderOutput{}, fmt.Errorf("register order: %w", err)
    }
    return RegisterOrderOutput{ID: order.ID()}, nil
}

// Порт UnitOfWork в domain/ (architecture.md)
type UnitOfWork interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// Доменные события: накопить → отдать через PullEvents (architecture.md)
// агрегат: order.RecordEvent(OrderRegistered{...}); затем outbox.Append(order.PullEvents())
```
Источник паттернов: `.planning/research/ARCHITECTURE.md` (locked в PROJECT.md Key Decisions). [VERIFIED: .planning/research/ARCHITECTURE.md + PROJECT.md]

### §2 — Suite-бутстрап Ginkgo v2 (РЕАЛЬНЫЙ, из репо — для testing.md)
```go
// Source: pkg/http/http_test.go (компилируется, тест зелёный)
package http

import (
    "testing"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestHTTPSuite(t *testing.T) {
    RegisterFailHandler(Fail)        // MUST — каркас (D-09)
    RunSpecs(t, "HTTP Package Suite")
}
```
Структура спека (РЕАЛЬНЫЙ, `pkg/http/middlewares_test.go`): внешний пакет `http_test`, dot-imports, `Describe("...", func(){ Context("...", func(){ It("should ...", func(){ Expect(...).To(...) }) }) })`. EN-комментарии. [VERIFIED: pkg/http/*_test.go]

### §3 — mockery v3 + Gomega под порт (КОНВЕНЦИЯ, planned: Phase 4)
```go
// КОНВЕНЦИЯ для testing.md — установка/генерация: planned Phase 4 (D-10/D-11).
// mockery v3 генерирует testify-style мок с expecter API.

var _ = Describe("RegisterOrderUseCase", func() {
    var (
        uc   *RegisterOrderUseCase
        repo *MockOrderRepository // сгенерирован mockery
        uow  *MockUnitOfWork
    )
    BeforeEach(func() {
        // GinkgoT() удовлетворяет testify TestingT и поддерживает Cleanup —
        // NewMockX авто-регистрирует AssertExpectations на teardown спека.
        repo = NewMockOrderRepository(GinkgoT())
        uow  = NewMockUnitOfWork(GinkgoT())
        uc   = &RegisterOrderUseCase{orders: repo, uow: uow}
    })
    It("saves the order within the unit of work", func() {
        // expecter API
        uow.EXPECT().Do(mock.Anything, mock.Anything).
            RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
                return fn(ctx)
            })
        repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Once()

        out, err := uc.Execute(context.Background(), RegisterOrderInput{SKU: "X", Qty: 1})

        Expect(err).ToNot(HaveOccurred())  // Gomega ассерт результата
        Expect(out.ID).ToNot(BeEmpty())
        // AssertExpectations выполнится автоматически через GinkgoT().Cleanup
    })
})
```
Источник: [CITED: vektra.github.io/mockery/latest] (expecter `EXPECT().Method().Return()`, `NewMockX(t)` + `t.Cleanup`), [CITED: onsi.github.io/ginkgo] (`GinkgoT()` в core DSL). Конкретный стиль (expecter vs классический) — на усмотрение планировщика (D-10 discretion). [ASSUMED] деталь: `GinkgoT()` совместим с testify `TestingT`+`Cleanup` — общеизвестно, но не скомпилировано в этой сессии (mockery в репо нет).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CQRS-шина / диспетчер (`pkg/mediatr`, `CommandDispatcher`/`QueryDispatcher`) | Inbound-адаптеры зовут use case'ы напрямую; query-lite для чтения | Снесено (PROJECT.md) | `architecture.md` MUST NOT возрождать диспетчер/mediatr |
| `TxManager` / `tx.go` | Порт `UnitOfWork`, транзакция в `ctx` | Снесено (PROJECT.md) | `architecture.md` MUST NOT возрождать `TxManager` |
| mockery v2 (`with-expecter` флаг) | mockery v3 (`.mockery.yaml` packages-синтаксис, expecter по умолчанию, `template: testify`) | mockery v3 | `testing.md`/Phase 4 ориентируются на v3-синтаксис |
| Ginkgo v1 | Ginkgo v2 (`onsi/ginkgo/v2`) | pinned v2.23.4 | `testing.md` MUST на v2; репо уже на v2 |

**Deprecated/outdated:**
- `pkg/mediatr`, `CommandDispatcher`, `QueryDispatcher`, `TxManager`, `tx.go` — удалены, невалидны. `architecture.md` фиксирует явный MUST NOT их возрождать.
- Ginkgo v1, mockery v2 синтаксис (`--with-expecter` CLI флаг вместо `.mockery.yaml`) — не использовать.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `GinkgoT()` совместим с testify `TestingT` и поддерживает `Cleanup`, поэтому `NewMockX(GinkgoT())` авто-ассертит на teardown | Code Examples §3 | LOW — общеизвестная связка; если нюанс иной, поправится в Phase 4 при реальной установке mockery. `testing.md` метит mockery `planned`, так что незакомпилированная деталь не блокирует |
| A2 | Имя плейсхолдера `Order` (использовано в примерах) | Code Examples, Patterns | NONE — D-05 явно отдаёт выбор имени планировщику; `Order` — лишь иллюстрация в research |
| A3 | mockery v3 default filename `mocks_test.go`, `template: testify` | Standard Stack, State of the Art | LOW — CITED из офиц. docs; точная обвязка всё равно Phase 4 |
| A4 | Go-toolchain версия (1.23.6 в go.mod vs 1.24.6 в доках) — расхождение реально | Standard Stack ⚠️ | LOW — не предмет Phase 3 (версия Go — факт `structure.md`/`build.md`); доки P3 на неё не ссылаются жёстко |

**Эти `[ASSUMED]` пункты НЕ блокируют планирование:** A1/A3 относятся к mockery, который в Phase 3 только описывается как конвенция (`planned: Phase 4`) — фактическая компиляция произойдёт в Phase 4. A2 — явная discretion. A4 — вне скоупа фазы.

## Open Questions

1. **Дробить ли `architecture.md` заранее?**
   - What we know: 5 паттернов + диаграмма + MUST NOT рискуют дать >200 строк (authoring лимит).
   - What's unclear: уложится ли в один файл при достаточно сжатой подаче (правила, не how-to — детали в `patterns.md`).
   - Recommendation: Писать как один файл на уровне инвариантов; если на ревью >200 строк — выделить события (outbox/relay/PullEvents) в под-топик и связать ссылкой. Решение — планировщику (discretion).

2. **Нужно ли обновлять таблицу статусов в корневом `AGENTS.md` (L37)?**
   - What we know: `AGENTS.md` L37 перечисляет 4 дока со статусом «запланировано (Phase 3)».
   - What's unclear: считается ли `AGENTS.md` частью скоупа Phase 3 или только `knowledge/`.
   - Recommendation: Да — синхронность обязательна (no-phantom для статусов). Включить мелкую правку статуса в `AGENTS.md` в скоуп фазы (снять «запланировано» для наполненных доков), либо явно отметить в плане как интеграционную правку.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Цитирование/проверка `pkg/http/*_test.go` (опц.) | ✓ (репо собирается) | go.mod: 1.23.6 (доки: 1.24.6 — расхождение) | — |
| Ginkgo v2 / Gomega | Реальность эталона `testing.md` | ✓ | v2.23.4 / v1.38.0 (pkg/go.mod) | — |
| mockery v3 | Генерация моков (для testing.md как конвенция) | ✗ (нет в репо) | — (CITED v3) | Фаза 3 только ОПИСЫВАЕТ конвенцию (`planned: Phase 4`); установка не требуется |
| Markdown link-checker (опц.) | Валидация ссылок | ? (не проверено) | — | Ручной grep + проверка `test -f` для каждой ссылки (см. Validation Architecture) |

**Missing dependencies with no fallback:** none (документационная фаза).
**Missing dependencies with fallback:** mockery (не нужен — только конвенция); link-checker (заменяется grep/`test -f`).

## Validation Architecture

> nyquist_validation включён (config.json). Это документационная фаза — «тесты» = проверки, что 4 дока удовлетворяют success criteria. Все проверки скриптуемы из bash (нет тест-фреймворка для прозы).

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Shell-проверки (grep / `test -f`) — прозу нельзя валидировать unit-тестом; Go-эталон в `testing.md` проверяется `go vet`/`go build` опционально |
| Config file | none — проверки инлайн в VALIDATION.md (Wave 0 не нужен) |
| Quick run command | `bash` one-liners (см. ниже), запускаются из корня репо |
| Full suite command | последовательность всех проверок ниже |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DOC-04 | `style.md` существует и содержит правило языка комментариев | presence | `test -f knowledge/style.md && grep -qiE 'русск|коммент' knowledge/style.md` | ❌ создаётся фазой |
| DOC-04 | Правило языка — ТОЛЬКО в `style.md` (нет дубля) | uniqueness | `! grep -rl 'комментарии.*на русском\|RU комментарии' knowledge/ | grep -v 'style.md\|boundaries.md\|README.md'` (только канон+карта+индекс) | ❌ |
| DOC-04 | typed ID / sentinel `%w` / DTO→домен присутствуют | presence | `grep -qE '%w' knowledge/style.md && grep -qiE 'типизирован|typed' knowledge/style.md && grep -qiE 'DTO' knowledge/style.md` | ❌ |
| DOC-03 | `testing.md` существует, фиксирует Ginkgo v2+Gomega, EN-комментарии, mockery | presence | `test -f knowledge/testing.md && grep -qi 'ginkgo' knowledge/testing.md && grep -qi 'gomega' knowledge/testing.md && grep -qi 'mockery' knowledge/testing.md` | ❌ |
| DOC-03 | suite-бутстрап MUST задокументирован | presence | `grep -q 'RegisterFailHandler' knowledge/testing.md && grep -q 'RunSpecs' knowledge/testing.md` | ❌ |
| DOC-05 | `architecture.md` явно DDD+гексагон БЕЗ CQRS + MUST NOT | presence | `test -f knowledge/architecture.md && grep -qi 'гексагон' knowledge/architecture.md && grep -qiE 'CQRS' knowledge/architecture.md && grep -qiE 'MUST NOT|WON.T' knowledge/architecture.md` | ❌ |
| DOC-05 | Элементы: Execute / query-lite / UnitOfWork / outbox / PullEvents | presence | `for k in Execute UnitOfWork outbox PullEvents query; do grep -qi "$k" knowledge/architecture.md || echo "MISSING $k"; done` | ❌ |
| DOC-05 | Диаграмма/таблица направления импортов (D-06) | presence | `grep -qiE 'domain.*←|импорт|import' knowledge/architecture.md` (визуальная проверка диаграммы вручную) | ❌ |
| PAT-01 | `patterns.md` — 4 рецепта, ссылается на architecture.md | presence | `test -f knowledge/patterns.md && grep -c 'architecture.md' knowledge/patterns.md` (>0) и наличие use case/query/aggregate/repository | ❌ |
| cross | no-phantom: все сниппеты architecture/patterns помечены «иллюстрация» | manual + grep | `grep -qiE 'иллюстрац|целевой вид' knowledge/patterns.md knowledge/architecture.md` | ❌ |
| cross | Link integrity: нет битых markdown-ссылок | link-check | для каждой `[..](X.md)` в knowledge/*.md → `test -f knowledge/X` | существует/частично |
| cross | Ownership-map: 4 факта зарегистрированы в boundaries.md | presence | `grep -q 'testing.md' knowledge/boundaries.md && grep -q 'architecture.md' knowledge/boundaries.md && grep -q 'patterns.md' knowledge/boundaries.md` | частично (style.md уже есть) |
| cross | README индекс обновлён (ссылки, не «запланировано» Phase 3) | presence | `grep -E '\[testing.md\]|\[architecture.md\]|\[patterns.md\]|\[style.md\]' knowledge/README.md` (ссылки вместо back-ticks) | частично |
| cross | Размер ~150–200 строк на док | lint | `for f in style testing architecture patterns; do wc -l knowledge/$f.md; done` (флаг при >200) | ❌ |
| cross | Каждое правило с тегом силы | lint | визуально: нет хеджирования без MUST/SHOULD/WON'T (`grep -iE 'обычно|желательно|prefer'` → должно быть пусто в нормативных строках) | ❌ |
| cross | Forward enforcement-метки (D-11) на механизируемых правилах | presence | `grep -qiE 'convention-only|planned.*Phase 4|CI-gated|hook' knowledge/style.md knowledge/testing.md knowledge/architecture.md` | ❌ |

### Sampling Rate
- **Per task commit:** `test -f` + presence-grep для дока этой задачи + link-check затронутых ссылок.
- **Per wave merge:** полный набор presence/uniqueness/link-integrity по всем созданным докам.
- **Phase gate:** все проверки зелёные + ручной просмотр диаграммы импортов (D-06) и меток «иллюстрация» (no-phantom) перед `/gsd-verify-work`.

### Wave 0 Gaps
- None — тест-инфраструктура не требуется (документационная фаза, проверки = shell one-liners). Link-checker опционален; fallback — grep + `test -f` (уже специфицированы выше).

## Security Domain

> `security_enforcement: true`, ASVS level 1. Phase 3 — документационная (нет кода, ввода, аутентификации, крипто, сети). ASVS-категории к **самой фазе** не применимы.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — (нет кода/аутентификации в фазе) |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | no (для фазы) | — Но `architecture.md` SHOULD упомянуть: валидация на edge (protovalidate в `api/`), домен не валидирует транспорт. Это конвенция, не реализация |
| V6 Cryptography | no | — `style.md` SHOULD: WON'T хэндроллить крипто (общая Go-конвенция, ссылка); реализации в фазе нет |

### Known Threat Patterns for {stack}
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Документ описывает несуществующую защиту как реальную (phantom security rule) | Repudiation/Spoofing (доверие к ложному правилу) | no-phantom: не документировать бизнес-фичи/защиту, которых нет (boundaries.md); security-конвенции домена — в будущих эпиках |
| Утечка инфраструктуры/транспорта в домен (описано как анти-паттерн) | Tampering | `architecture.md` Anti-Pattern 3: только порты в домене, реализации в адаптерах |

> Реальные security-контроли платформы (ownership-race, SSH-grants, audit destructive actions — `.planning/research/PITFALLS.md` Security) — **будущие эпики, вне Phase 3**. `architecture.md` НЕ должен документировать их как существующие правила (no-phantom).

## Sources

### Primary (HIGH confidence)
- `pkg/http/http_test.go`, `pkg/http/middlewares_test.go`, `pkg/http/client_test.go` — реальный компилируемый Ginkgo v2+Gomega эталон (suite-бутстрап, Describe/Context/It, dot-imports, EN-комментарии)
- `pkg/go.mod` — pinned ginkgo v2.23.4, gomega v1.38.0
- `.planning/PROJECT.md` Key Decisions, `.planning/research/ARCHITECTURE.md` — целевая архитектура (DDD+гексагон БЕЗ CQRS, Execute, query-lite, UnitOfWork, outbox+relay, PullEvents) — locked
- `knowledge/authoring.md`, `knowledge/boundaries.md`, `knowledge/README.md`, `knowledge/structure.md`, `knowledge/build.md`, `knowledge/git.md` — каркас/стандарт/карта владения (Phase 1–2)
- `.planning/phases/03-conventions-architecture-docs/03-CONTEXT.md` — locked D-01..D-11
- `.planning/research/PITFALLS.md` — pitfalls (мега-файл, MUST/SHOULD+do, phantom-правила, кросс-док дубли)

### Secondary (MEDIUM confidence)
- vektra.github.io/mockery/latest (overview + configuration) — mockery v3, expecter API, `.mockery.yaml` packages-синтаксис, `template: testify`, `NewMockX(t)` + Cleanup [CITED]
- onsi.github.io/ginkgo — `GinkgoT()` в core DSL [CITED]

### Tertiary (LOW confidence)
- [ASSUMED] деталь: `GinkgoT()` ↔ testify `TestingT`+`Cleanup` совместимость (общеизвестно, но не скомпилировано в сессии; mockery в репо нет) — помечено в Assumptions Log A1

## Metadata

**Confidence breakdown:**
- Authoring/scope/архитектурные правила: HIGH — locked в CONTEXT/PROJECT, подтверждены инспекцией репо и существующих доков
- testing.md эталон (Ginkgo/Gomega): HIGH — реальный компилируемый код в `pkg/http`
- mockery v3 ↔ Gomega стыковка: MEDIUM — CITED из офиц. docs, не скомпилировано (mockery нет в репо; фаза описывает как `planned`)
- Validation подход: HIGH — shell-проверки прямо отражают success criteria

**Research date:** 2026-06-17
**Valid until:** ~2026-07-17 (стабильно; mockery v3/Ginkgo v2 — медленно меняются; locked-решения проекта не меняются вне discuss)
