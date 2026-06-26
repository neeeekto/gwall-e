---
phase: 03-conventions-architecture-docs
plan: 05
subsystem: knowledge-base-docs
tags: [docs, integration, link-integrity, no-phantom]
requires: ["03-01", "03-02", "03-03", "03-04"]
provides: ["knowledge-index-complete", "ownership-map-complete", "agents-status-synced"]
affects: [knowledge/README.md, knowledge/boundaries.md, AGENTS.md]
tech_stack:
  added: []
  patterns: [pointer-over-copy, no-phantom-links-and-statuses]
key_files:
  created: []
  modified:
    - knowledge/README.md
    - knowledge/boundaries.md
    - AGENTS.md
decisions:
  - "Финальная волна Phase 3: ссылки на 4 дока добавлены только теперь, когда все файлы существуют (no-phantom для ссылок)"
  - "glossary.md НЕ существует — оставлен без ссылки, честный статус «отложено (domain-milestone)» во всех трёх файлах"
  - "WARNING-1: стале-строка Phase 2 в AGENTS.md разбита — structure/build/git/boundaries получили ссылки и статус «есть», glossary вынесен отдельной строкой без ссылки"
metrics:
  duration: ~3min
  completed: 2026-06-17
---

# Phase 3 Plan 5: Integration — Index, Ownership Map, Status Table Summary

Вписаны 4 дока Phase 3 (style/testing/architecture/patterns) в индекс README.md, карту владения фактами boundaries.md и таблицу статусов AGENTS.md — без битых ссылок; стале-статусы «запланировано (Phase 2/3)» сняты, glossary.md оставлен честно отложенным без ссылки.

## What Was Built

- **knowledge/README.md** — индекс: 4 back-tick-имени заменены на markdown-ссылки `[style.md](style.md)` и т.д., статус «запланировано (Phase 3)» → «существует»; в порядке чтения снят курсив `*(Phase 3)*` и имена сделаны кликабельными.
- **knowledge/boundaries.md** — карта владения фактами: строка style.md → «существует» + ссылка; добавлены 3 строки (testing.md, architecture.md, patterns.md) со статусом «существует».
- **AGENTS.md** — таблица «База знаний»: Phase 3 строка → ссылки + «есть»; стале-строка Phase 2 разбита (structure/build/git/boundaries → ссылки + «есть»); glossary.md вынесен отдельной строкой без ссылки «отложено (domain-milestone)»; дополнительно исправлена стале-ссылка `(Phase 2)` в секции «Границы» на живую ссылку boundaries.md.

## Task Commits

| Task | Name | Commit | Files |
| ---- | ---- | ------ | ----- |
| 1 | Обновить knowledge/README.md (индекс + порядок чтения) | 3c40842 | knowledge/README.md |
| 2 | Обновить knowledge/boundaries.md (карта владения) | a9e209f | knowledge/boundaries.md |
| 3 | Синхронизировать таблицу статусов AGENTS.md (Phase 3 + стале Phase 2) | 4134449 | AGENTS.md |

## Verification

Все cross-cutting проверки из 03-VALIDATION.md зелёные:
- README индекс ссылается на все 4 дока: OK
- Ownership-map регистрирует testing/architecture/patterns: OK
- AGENTS.md синхронизирован — ни «запланировано (Phase 3)», ни «запланировано (Phase 2)» не осталось; стале-ссылка `(Phase 2)` в секции «Границы» тоже снята: OK
- Link integrity: полный скан всех `knowledge/*.md` — ни одной битой ссылки; все 9 `knowledge/*.md`-целей резолвятся: OK
- Uniqueness языка кода (устойчивая к порядку слов проверка): правило живёт только в каноне style.md: OK
- Размеры 4 доков ≤ ~200 строк: style 110, testing 129, architecture 148, patterns 187: OK

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug / no-phantom] Снята стале-ссылка `(Phase 2)` в секции «Границы» AGENTS.md**
- **Found during:** Task 3
- **Issue:** В секции «Границы (do-not)» строка «Полное правило придёт в `knowledge/boundaries.md` (Phase 2)» оставалась фантомной — boundaries.md уже существует, формулировка обещала будущий файл как несуществующий.
- **Fix:** Заменено на «Полное правило — в [knowledge/boundaries.md](knowledge/boundaries.md)» (живая ссылка, без phantom-статуса).
- **Files modified:** AGENTS.md
- **Commit:** 4134449
- **Why:** Тот же no-phantom-инвариант фазы (T-03-07), что и WARNING-1 для таблицы; план явно предписывал убрать стале-статусы Phase 2.

### glossary.md (planned-as-deferred, не девиация)

glossary.md по плану НЕ создаётся в этой фазе (отложен в domain-milestone). Подтверждено `test -f knowledge/glossary.md` = false. Во всех трёх файлах оставлен без ссылки, со статусом «отложено» — link integrity не нарушена.

## Known Stubs

None. Все ссылки ведут на реально существующие файлы; единственный нессылочный пункт (glossary.md) — намеренно отложен в domain-milestone и честно помечен.

## Self-Check: PASSED

- knowledge/README.md: FOUND (modified, commit 3c40842)
- knowledge/boundaries.md: FOUND (modified, commit a9e209f)
- AGENTS.md: FOUND (modified, commit 4134449)
- Commit 3c40842: FOUND
- Commit a9e209f: FOUND
- Commit 4134449: FOUND
- Link integrity scan over all knowledge/*.md: 0 broken links
