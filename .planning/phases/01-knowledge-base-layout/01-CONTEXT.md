# Phase 1: Раскладка базы знаний и точки входа - Context

**Gathered:** 2026-06-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Фаза создаёт **структуру** базы знаний и **тонкие точки входа**, но НЕ наполняет topic-доки контентом (это Phase 2–3). В скоупе: каталог `knowledge/` с индексом `README.md`, файл authoring-стандарта, тонкий гибридный `CLAUDE.md` и канонический `AGENTS.md`. Цель — чтобы все последующие доки писались по единому стандарту и подключались через progressive disclosure.

Требования: **KB-01** (тонкий `CLAUDE.md` <~150 строк → ссылки на `knowledge/*.md`), **KB-02** (`knowledge/README.md` индекс), **KB-03** (`AGENTS.md` тонкий кросс-тульный вход), **KB-04** (authoring-стандарт MUST/SHOULD/WON'T + «do» к каждому запрету).

</domain>

<decisions>
## Implementation Decisions

### Структура knowledge/
- **D-01:** Плоская структура — все доки лежат в `knowledge/*.md`, без подпапок. Подпапки (`decisions/`, `conventions/`) добавим позже, когда число файлов вырастет.
- **D-02:** Имена файлов — kebab-case, английские: `architecture.md`, `style.md`, `glossary.md`, `build.md`, `structure.md`, `testing.md`, `git.md`, `boundaries.md`, `patterns.md`, `authoring.md`. (Стабильно, как в research и в экосистеме CLAUDE.md/AGENTS.md.)
- **D-03:** Язык **содержимого** доков — русский (как комментарии кода). Тех-термины и примеры кода — английские; доменные термины канонизируются в `glossary.md` с маппингом EN/RU.
- **D-04:** `knowledge/README.md` — формат **таблица**: `Файл | Назначение | Когда читать` + явный порядок чтения (reading order). Это индекс-якорь для всех доков.

### Точки входа (CLAUDE.md / AGENTS.md)
- **D-05:** **`AGENTS.md` — источник истины** (канонический тонкий вход): короткая шапка проекта + Core Value + таблица-ссылки на `knowledge/*.md` + указатель на `authoring.md`. Кросс-тульный открытый стандарт.
- **D-06:** **`CLAUDE.md` — тонкий гибрид / указатель**: короткая шапка + ссылка на `AGENTS.md` и `knowledge/` + **сохранённый GSD workflow-блок** (он специфичен для Claude Code/GSD-исполнения — остаётся в CLAUDE.md). Тяжёлый авто-дамп GSD-секций `stack`/`architecture`/`conventions` **убрать** (контент дублирует `knowledge/`). Целевой объём — <~150 строк (KB-01).
- **D-07:** `AGENTS.md` и `CLAUDE.md` — **отдельные файлы, не symlink** (т.к. CLAUDE.md несёт GSD workflow-блок, которого нет смысла тащить в AGENTS.md). Контент НЕ дублируется: CLAUDE.md ссылается на AGENTS.md/knowledge.

### Authoring-стандарт (KB-04)
- **D-08:** Authoring-стандарт — **отдельный файл `knowledge/authoring.md`**: как писать правила (каждое нормативное правило помечается MUST / SHOULD / WON'T; каждый запрет сопровождается предписанной альтернативой «do»). В `README.md` — краткая памятка + ссылка на `authoring.md`. Все доки Phase 2–4 обязаны следовать этому стандарту.

### Claude's Discretion
- Точные колонки/формулировки таблицы в `README.md` (сверх минимума `Файл | Назначение | Когда читать`).
- Конкретная разметка силы правил (например, жирный префикс `**MUST**` / `**SHOULD**` / `**WON'T**` vs бейджи) — определить в `authoring.md`.
- Точное содержание шапки `AGENTS.md` (взять из PROJECT.md «What This Is» + Core Value).
- Создавать ли файлы-стабы для будущих topic-доков (Phase 2–3) или только перечислять их в README со статусом «запланировано» — на усмотрение планировщика, но **без битых ссылок** (если ссылка есть — файл/стаб должен существовать).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Источник истины по решениям и скоупу
- `.planning/PROJECT.md` — Key Decisions: целевая архитектура (DDD+гексагон БЕЗ CQRS), язык кода (RU комментарии/термины, EN имена), список решений для будущих ADR.
- `.planning/REQUIREMENTS.md` — требования KB-01..04 (скоуп Phase 1), а также карта v1/v2 (что НЕ документировать сейчас).

### Research (раскладка базы знаний и грабли)
- `.planning/research/ARCHITECTURE.md` — «Layer 1: структура базы знаний» (тонкие точки входа + progressive disclosure, граф зависимостей доков).
- `.planning/research/FEATURES.md` — Part A: table-stakes доки, дифференциаторы, anti-features (что НЕ класть в базу); порядок зависимостей доков.
- `.planning/research/PITFALLS.md` — Pitfall 2 (mega-file/контекст-бюджет), Pitfall 5 (MUST/SHOULD/WON'T + «do» к запретам), Pitfall 8 (phantom-правила), Pitfall 1 (язык — решён: русский).
- `.planning/research/SUMMARY.md` — Phase 1 implications (раскладка + authoring-стандарт первыми).

### Текущее состояние входных файлов
- `CLAUDE.md` (корень) — сейчас GSD-сгенерён, ~205 строк, управляемые маркеры `<!-- GSD:project/stack/conventions/architecture/skills -->`. Подлежит урезанию (D-06).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `knowledge/` — каталог уже существует, но **пустой**. Стартуем с нуля.
- `CLAUDE.md` — содержит готовый GSD workflow-блок и шапку проекта; из него берём шапку и оставляем workflow-блок, остальное (stack/architecture/conventions/skills дампы) убираем.

### Established Patterns
- GSD-маркеры в CLAUDE.md (`<!-- GSD:*-start/end -->`) — генерируются `gsd-tools query generate-claude-md` из PROJECT.md/research. Тонкий гибрид должен оставить нужные блоки и не тянуть тяжёлые.
- Progressive disclosure (тонкий вход → topic-файлы) — зафиксировано в PROJECT.md и research.

### Integration Points
- Корень репо: `CLAUDE.md`, новый `AGENTS.md`, каталог `knowledge/`.
- Будущие доки Phase 2–4 подключаются через таблицу в `knowledge/README.md` и ссылки из `AGENTS.md`.

### Risks / Constraints for planner
- ⚠️ **Регенерация CLAUDE.md:** `generate-claude-md` повторно раздувает CLAUDE.md тяжёлыми секциями из research. Планировщик должен это учесть: либо впредь не запускать авто-генерацию для этого проекта, либо ограничить/отключить тяжёлые секции (stack/architecture/conventions), чтобы тонкий гибрид (D-06) не перезатирался. Зафиксировать как правило в `boundaries.md` (Phase 2).
- `memory-bank/` на диске нет; старый CLAUDE.md из git HEAD ссылался на несуществующие memory-bank-доки — новые ссылки вести только на реально создаваемые `knowledge/*.md` (без phantom-ссылок, Pitfall 8).

</code_context>

<specifics>
## Specific Ideas

- Пользователь ранее предложил: в `boundaries.md` (Phase 2) добавить пункт «что НЕ кладём в knowledge/» (черновики/спайки/планы — не канон), а в `knowledge/README.md` (Phase 1) — короткую карту «что где живёт» (knowledge/ = канон, .planning/ = процесс). Учесть при планировании Phase 1 README и Phase 2 boundaries.
- ADR отложены в v2; решения пока живут в `PROJECT.md → Key Decisions`. README может упомянуть это как «лог решений — в PROJECT.md (ADR придут в v2)».

</specifics>

<deferred>
## Deferred Ideas

- ADR-доки (`knowledge/decisions/`), `anti-patterns.md`, `libraries.md`, onboarding-гайд, maintenance-протокол — отложены в v2 (см. REQUIREMENTS.md v2).
- Подпапки внутри `knowledge/` — ввести позже при росте числа доков (D-01).
- Наполнение topic-доков контентом — Phase 2 (glossary/structure/build/git/boundaries) и Phase 3 (style/testing/architecture/patterns).

</deferred>

---

*Phase: 1-Раскладка базы знаний и точки входа*
*Context gathered: 2026-06-17*
