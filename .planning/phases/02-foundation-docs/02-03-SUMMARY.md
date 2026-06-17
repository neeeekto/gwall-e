---
phase: 02-foundation-docs
plan: 03
subsystem: docs
tags: [knowledge-base, boundaries, do-not-rules, fact-ownership, authoring]

# Dependency graph
requires:
  - phase: 02-foundation-docs
    plan: 01
    provides: knowledge/structure.md (WIP-статус inventory, упоминание boundaries.md), knowledge/build.md, README.md index
  - phase: 02-foundation-docs
    plan: 02
    provides: knowledge/git.md, README.md index wired to git.md (boundaries.md ещё запланировано)
  - phase: 01-knowledge-base-layout
    provides: knowledge/authoring.md (MUST/SHOULD/WON'T, парность запрет→do, pointer-over-copy, no-phantom)
provides:
  - knowledge/boundaries.md — правила «do-not» (WIP-леса, стале-файлы, phantom) на уровне возможностей+примеры; оба посева Phase 1; карта владения фактами
  - knowledge/structure.md — рабочая обратная ссылка на boundaries.md (взаимная пара замкнута)
  - knowledge/README.md index — boundaries.md со ссылкой и статусом «существует»
affects: [onboarding, architecture.md, style.md, enforcement-phase]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Do-not правила на уровне возможностей + примеры (inventory/gateway/outgate как ПРИМЕРЫ, не жёсткий path-список)"
    - "Карта владения фактами: один факт = один канон-док, остальные ссылаются (pointer-over-copy)"
    - "Взаимные относительные ссылки boundaries.md <-> structure.md по WIP-статусу inventory"

key-files:
  created:
    - knowledge/boundaries.md
  modified:
    - knowledge/structure.md
    - knowledge/README.md

key-decisions:
  - "boundaries.md следует D-06: do-not на уровне возможностей+примеры, без хрупких path-карт; inventory/gateway/outgate — примеры"
  - "Перенесены оба посева Phase 1 (D-07): .planning/ не канон; WON'T ре-раздувать CLAUDE.md через generate-claude-md"
  - "Карта владения фактами (D-08) как markdown-таблица: structure.md/build.md/git.md/authoring.md — ссылки; style.md — «запланировано» без ссылки (no-phantom)"
  - "Пара boundaries.md <-> structure.md замкнута (D-03/D-08): голое упоминание в structure.md превращено в рабочую ссылку"

patterns-established:
  - "Pattern: карта владения фактами таблицей со статусом «существует/запланировано» — ссылки только на реальные файлы"

requirements-completed: [DOC-08]

# Metrics
duration: 1min
completed: 2026-06-17
---

# Phase 02 Plan 03: boundaries.md Summary

**Границы «не трогать» и единая карта владения фактами закреплены в knowledge/boundaries.md: WON'T-правила для WIP-лесов вне go.work (inventory/gateway/outgate как примеры), стале README/Makefile/docker-compose не авторитетны, phantom-фичи не документируются; перенесены оба посева Phase 1 (.planning/ не канон; WON'T ре-раздувать CLAUDE.md через generate-claude-md); карта владения фактами (один факт = один канон) ссылается на structure.md/build.md/git.md/authoring.md и помечает style.md запланированным. Взаимная пара boundaries.md <-> structure.md по WIP-статусу inventory замкнута; индекс README обновлён, битых ссылок нет.**

## Performance

- **Duration:** ~1 min
- **Completed:** 2026-06-17
- **Tasks:** 3
- **Files modified:** 3 (1 created, 2 modified)

## Accomplishments
- `knowledge/boundaries.md` (71 строка, < 150) — по стандарту authoring.md, каждый запрет тегирован **WON'T** и парен с предписанной **MUST**-альтернативой (9× WON'T, 6× MUST):
  - не «чинить» WIP-леса вне `go.work` (nil-deps, пустые `internal/`, не компилирующиеся пакеты); `inventory` как пример со ссылкой на `structure.md` за WIP-статусом; `gateway`/`outgate` как примеры заглушек;
  - стале `README`/`Makefile`/`docker-compose.yml` не авторитетны — опираться на `knowledge/`-канон;
  - phantom-фичи не документировать (per authoring no-phantom);
  - посев Phase 1 (а): `.planning/` — процесс, не канон;
  - посев Phase 1 (б): **WON'T** ре-раздувать `CLAUDE.md` / запускать `generate-claude-md` для тяжёлых секций — держать тонкий гибрид;
  - карта владения фактами таблицей: `structure.md` (WIP/раскладка), `build.md` (команды), `git.md` (git), `authoring.md` (стандарт) — со ссылками; `style.md` — «запланировано (Phase 3)» без ссылки.
- `knowledge/structure.md` — голое упоминание `boundaries.md` в разделе про inventory WIP превращено в рабочую относительную ссылку `[boundaries.md](boundaries.md)`; взаимная пара замкнута.
- `knowledge/README.md` — индекс-строка boundaries.md → рабочая ссылка + статус «существует»; в «Порядке чтения» снят маркер «(Phase 2)»; строки Phase 3 (style.md и др.) остаются «запланировано» без ссылок. Финальная проверка: ни одной битой relative-ссылки в README.md, structure.md, build.md, git.md, boundaries.md.

## Task Commits

Each task was committed atomically:

1. **Task 1: Создать knowledge/boundaries.md** - `9cc2ee1` (docs)
2. **Task 2: Замкнуть обратную ссылку в structure.md** - `9a7d33c` (docs)
3. **Task 3: Обновить README.md индекс + финальная проверка ссылок** - `8c8b9ae` (docs)

## Files Created/Modified
- `knowledge/boundaries.md` - Правила «do-not» + оба посева Phase 1 + карта владения фактами; канон границ
- `knowledge/structure.md` - Рабочая обратная ссылка на boundaries.md (взаимная пара замкнута)
- `knowledge/README.md` - Индекс-ссылка и статус «существует» для boundaries.md; «Порядок чтения» обновлён

## Decisions Made
- Следовал D-06/D-07/D-08 точно: do-not на уровне возможностей+примеры (без path-карт), оба посева Phase 1, карта владения фактами как таблица «факт → канон».
- Карта владения размещена в `boundaries.md` (усмотрение планировщика per D-08 / Claude's Discretion); расширяет блок «Что где живёт» в README ссылочно, не копируя.
- Ссылки в карте — только на существующие файлы; `style.md` указан со статусом «запланировано» без ссылки (no-phantom, authoring).

## Deviations from Plan

None - plan executed exactly as written. Все три task-верификации и acceptance criteria прошли с первой попытки; авто-фиксы (Rules 1-3) и архитектурные решения (Rule 4) не потребовались.

## Issues Encountered
None. Реальное состояние совпало с предположениями плана: structure.md из 02-01 оставил голое упоминание `boundaries.md` (без ссылки), как и ожидал Task 2 — превращено в рабочую ссылку.

## Threat Surface
T-02-01 (Information Disclosure) mitigated: boundaries.md просканирован на секреты/токены/приватные ключи/внутренние хостнеймы (`.internal`)/private IP/ssh-ключи — чисто; документированы только публичные факты репо (границы, имена доков, карта владения). Исполняемой поверхности не введено (T-02-02 accept).

## User Setup Required
None - только документация, внешняя конфигурация не требуется.

## Next Phase Readiness
- DOC-08 закрыт; границы «не трогать» и карта владения фактами зафиксированы.
- Все четыре инфра/процесс-дока Phase 2 (structure.md, build.md, git.md, boundaries.md) наполнены и связаны в индексе без битых ссылок; Phase 2 (foundation-docs) полностью наполнена.
- Phase 3 доки (style.md/testing.md/architecture.md/patterns.md) остаются «запланировано» — карта владения уже резервирует `style.md` как канон языка кода.

## Self-Check: PASSED

All created/modified files exist on disk; all three task commits (`9cc2ee1`, `9a7d33c`, `8c8b9ae`) are present in git history.

---
*Phase: 02-foundation-docs*
*Completed: 2026-06-17*
