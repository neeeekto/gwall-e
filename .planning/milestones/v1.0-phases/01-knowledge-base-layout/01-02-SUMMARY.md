---
phase: 01-knowledge-base-layout
plan: 02
subsystem: docs
tags: [entry-points, agents-md, claude-md, thin-hybrid, progressive-disclosure, ru-docs]

# Dependency graph
requires:
  - "knowledge/README.md — индекс базы знаний (Plan 01)"
  - "knowledge/authoring.md — стандарт авторинга (Plan 01)"
provides:
  - "AGENTS.md — канонический тонкий кросс-тульный вход (источник истины, D-05)"
  - "CLAUDE.md урезан до тонкого гибрида (<150 строк), указывает на AGENTS.md/knowledge, сохранён GSD workflow-блок (D-06)"
affects: [phase-2-topic-docs, phase-3-topic-docs, phase-4-enforcement]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "AGENTS.md как открытый кросс-тульный стандарт (источник истины) + CLAUDE.md как тонкий Claude-специфичный указатель"
    - "Progressive disclosure: тонкий вход → topic-доки в knowledge/ через таблицу-ссылки"
    - "Pointer-over-copy: CLAUDE.md ссылается на AGENTS.md/knowledge, контент не дублируется (D-07)"
    - "Re-bloat guard: HTML-комментарий-предупреждение против generate-claude-md в CLAUDE.md"

key-files:
  created:
    - AGENTS.md
  modified:
    - CLAUDE.md

key-decisions:
  - "Шапка AGENTS.md взята из PROJECT.md «What This Is» + Core Value (Claude's Discretion из CONTEXT.md)"
  - "В AGENTS.md и CLAUDE.md будущие topic-доки перечислены без markdown-ссылок со статусом «запланировано» (Pitfall 8)"
  - "В CLAUDE.md сохранён компактный GSD:project-блок, но детали заменены ссылкой на AGENTS.md/knowledge — без дублирования (D-07)"
  - "Удалённые блоки stack/conventions/architecture/skills вырезаны вместе с GSD-маркерами; добавлен HTML-warning против re-bloat (T-01-04)"

patterns-established:
  - "Источник истины — AGENTS.md; CLAUDE.md = тонкий указатель + GSD workflow/profile блоки"
  - "Ссылки только на существующие файлы; boundary-правила (WON'T) с предписанной альтернативой"

requirements-completed: [KB-01, KB-03]

# Metrics
duration: 4min
completed: 2026-06-17
---

# Phase 1 Plan 02: Тонкие точки входа (AGENTS.md + урезанный CLAUDE.md) Summary

**Создан AGENTS.md как канонический тонкий кросс-тульный источник истины (шапка + Core Value + таблица-ссылки в knowledge/ + указатель на authoring.md) и урезан корневой CLAUDE.md с ~205 до 51 строк (тонкий гибрид: указатель на AGENTS.md/knowledge + сохранённый GSD workflow-блок), без phantom-ссылок и дублирования.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-06-17T10:58:59Z
- **Tasks:** 2
- **Files modified:** 2 (AGENTS.md создан, CLAUDE.md урезан)

## Accomplishments
- `AGENTS.md` (54 строки) — канонический источник истины (D-05): тонкая шапка из PROJECT.md + Core Value, указатель на стандарт авторинга, таблица-ссылки в knowledge/ (только на существующие README.md/authoring.md; будущие доки — без ссылок со статусом «запланировано»), блок «Точки входа» и boundary-правила (WON'T) на WIP inventory и stale-файлы.
- `CLAUDE.md` урезан до 51 строки (KB-01, D-06): удалены тяжёлые авто-секции stack/conventions/architecture/skills вместе с GSD-маркерами; сохранены GSD:workflow и GSD:profile; верх переписан в тонкий указатель на AGENTS.md (источник истины) и knowledge/README.md; добавлен HTML-warning против re-bloat через generate-claude-md.
- Ни одной битой/phantom-ссылки и ни одного упоминания memory-bank в обоих файлах; все knowledge/*.md ссылки разрешаются в реально существующие файлы.

## Task Commits

Each task was committed atomically:

1. **Task 1: AGENTS.md — канонический тонкий вход (KB-03, D-05/D-07)** - `cc7a262` (docs)
2. **Task 2: Урезать CLAUDE.md до тонкого гибрида (KB-01, D-06/D-07)** - `09c751a` (docs)

**Plan metadata:** see final docs commit.

## Files Created/Modified
- `AGENTS.md` - Канонический кросс-тульный тонкий вход: шапка + Core Value + указатель на authoring.md + таблица-ссылки в knowledge/ + точки входа + границы (do-not).
- `CLAUDE.md` - Тонкий гибрид: HTML-warning против re-bloat + шапка + указатель на AGENTS.md/knowledge + сохранённые GSD:project (компактный)/workflow/profile блоки.

## Decisions Made
- Шапка AGENTS.md собрана из PROJECT.md «What This Is» + Core Value (Claude's Discretion), сжата до 2 абзацев без stack-дампа.
- В CLAUDE.md оставлен компактный GSD:project-блок, но его содержимое заменено ссылкой на AGENTS.md/knowledge — чтобы не дублировать контент (D-07) и не дать generate-claude-md «зацепку» для раздувания.
- Будущие topic-доки в обоих файлах перечислены текстом со статусом «запланировано (Phase 2/3)», без ссылок (Pitfall 8).

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Тонкие точки входа готовы: AGENTS.md (источник истины) и тонкий CLAUDE.md указывают в knowledge/ через таблицу-ссылки.
- Topic-доки Phase 2–3 подключаются добавлением строки-ссылки в таблицы README/AGENTS вместе с самим файлом.
- Риск re-bloat CLAUDE.md задокументирован HTML-предупреждением; правило закрепляется в knowledge/boundaries.md (Phase 2). НЕ запускать generate-claude-md.
- Блокеров нет.

## Self-Check: PASSED

- FOUND: AGENTS.md
- FOUND: CLAUDE.md
- FOUND: .planning/phases/01-knowledge-base-layout/01-02-SUMMARY.md
- FOUND commit: cc7a262 (Task 1)
- FOUND commit: 09c751a (Task 2)

---
*Phase: 01-knowledge-base-layout*
*Completed: 2026-06-17*
