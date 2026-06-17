---
phase: 03-conventions-architecture-docs
plan: 02
subsystem: testing
tags: [ginkgo, gomega, mockery, knowledge-base, conventions, docs]

# Dependency graph
requires:
  - phase: 03-01
    provides: knowledge/style.md (канон языка/EN-комментариев, цель cross-link)
  - phase: 02
    provides: knowledge/authoring.md (стандарт MUST/SHOULD/WON'T), build.md (команды), boundaries.md (карта владения)
provides:
  - knowledge/testing.md — канон конвенций тестов (Ginkgo v2 + Gomega, suite-бутстрап, структура спеков, mockery-конвенция)
affects: [03-03-architecture, 03-05-readme-boundaries-integration, 04-enforcement-tooling]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Pointer-over-copy: команды → build.md, язык комментариев → style.md (рабочие markdown-ссылки, не копии)"
    - "Forward-enforcement-метки (D-11): convention-only / planned: Phase 4 (go:generate)"
    - "No-phantom (D-10): реальный код-эталон цитируется из pkg/http; mockery-обвязка помечена planned, плейсхолдер Order для иллюстративного мока"

key-files:
  created: [knowledge/testing.md]
  modified: []

key-decisions:
  - "MUST на каркас (suite-бутстрап RegisterFailHandler/RunSpecs, dot-imports, *_test.go рядом с кодом, EN-комментарии); SHOULD на структуру (Describe/Context/It, DescribeTable) — D-09"
  - "mockery (vektra/mockery) закреплён как канонический генератор моков портов + Gomega-ассерты; обвязка (go:generate/.mockery.yaml/установка) помечена planned: Phase 4 — D-10"
  - "Языковое правило тестов живёт только в style.md; testing.md ссылается, не дублирует (INFO-1, карта владения)"

patterns-established:
  - "Реальный код-эталон (pkg/http/*_test.go) цитируется с пометкой Source; иллюстративный mock-сниппет помечен «иллюстрация / целевой вид» на плейсхолдере Order"
  - "Каждое нормативное правило несёт тег силы и forward-enforcement-статус"

requirements-completed: [DOC-03]

# Metrics
duration: 2min
completed: 2026-06-17
---

# Phase 3 Plan 02: testing.md (конвенции тестов) Summary

**knowledge/testing.md фиксирует Ginkgo v2 + Gomega как канон тестов: MUST-каркас suite из реального pkg/http, SHOULD-структуру спеков и mockery-конвенцию мокинга портов (planned: Phase 4, no-phantom).**

## Performance

- **Duration:** 2 min
- **Started:** 2026-06-17T13:46:37Z
- **Completed:** 2026-06-17T13:47:49Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Создан `knowledge/testing.md` (129 строк) — канон конвенций тестов (DOC-03)
- Зафиксирован MUST-каркас suite: `RegisterFailHandler(Fail)` + `RunSpecs(t, "...")`, dot-imports, `*_test.go` рядом с кодом — цитата реального компилируемого эталона `pkg/http/http_test.go`
- Зафиксирована SHOULD-структура спеков (`Describe`/`Context`/`It`, `DescribeTable`) с реальным примером из `pkg/http/middlewares_test.go`
- Закреплён выбор mockery как канонического генератора моков портов + Gomega-ассерты; обвязка помечена `planned: Phase 4` (no-phantom)
- Cross-links рабочими markdown-ссылками: `[style.md](style.md)` (EN-комментарии) и `[build.md](build.md)` (команды) — без дублей

## Task Commits

Each task was committed atomically:

1. **Task 1: Написать knowledge/testing.md (Ginkgo v2 + Gomega + mockery-конвенция)** - `274f93c` (docs)

## Files Created/Modified
- `knowledge/testing.md` - канон конвенций тестов: фреймворк Ginkgo v2 + Gomega, MUST-каркас suite, SHOULD-структура спеков, мокинг портов через mockery (planned Phase 4), cross-links на style.md и build.md

## Decisions Made
None - followed plan as specified. Все решения уже зафиксированы в 03-CONTEXT.md (D-09, D-10, D-11) и применены без отклонений.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required. Установка mockery (`go install`, `.mockery.yaml`, `go:generate`) — задел Phase 4 (ENF-05), не этой фазы.

## Threat Surface

T-03-02 (mitigate) выполнен: mockery-обвязка помечена `planned: Phase 4`, рабочая команда генерации НЕ подаётся как проверенная, mock-сниппет помечен «иллюстрация / целевой вид» на плейсхолдере `Order`. T-03-SC (accept): пакеты не устанавливались (документационная фаза). Новой security-релевантной поверхности файл не вводит.

## Next Phase Readiness
- DOC-03 закрыт; testing.md готов для регистрации в индексе README.md и карте владения boundaries.md (Wave 5, plan 03-05)
- Cross-links на style.md и build.md разрешены (оба файла существуют); финальная uniqueness/link-integrity проверка — Wave 5
- Готов для перекрёстных ссылок из architecture.md/patterns.md, если потребуется ссылка за тест-конвенциями

## Self-Check: PASSED

- FOUND: knowledge/testing.md
- FOUND: .planning/phases/03-conventions-architecture-docs/03-02-SUMMARY.md
- FOUND commit: 274f93c

---
*Phase: 03-conventions-architecture-docs*
*Completed: 2026-06-17*
