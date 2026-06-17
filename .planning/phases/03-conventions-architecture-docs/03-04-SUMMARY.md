---
phase: 03-conventions-architecture-docs
plan: 04
subsystem: docs
tags: [knowledge-base, ddd, hexagonal, patterns, recipes, go, authoring]

# Dependency graph
requires:
  - phase: 03-conventions-architecture-docs
    provides: "architecture.md (03-03) — канон правил слоёв/UnitOfWork/outbox/PullEvents, плейсхолдер Order; style.md (03-01) — язык кода/typed IDs/ошибки/DTO→домен"
provides:
  - "knowledge/patterns.md — копируемые рецепты add use case/query/aggregate/repository (PAT-01), вертикальный срез до wiring, pointer-over-copy на architecture.md/style.md"
affects: [phase-04-enforcement-tooling, knowledge-base-index-wave-5, boundaries-ownership-map]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Правила vs рецепты (D-04): architecture.md держит инварианты/why, patterns.md — how-to со ссылками, без дублирования"
    - "Иллюстративные помеченные сниппеты на едином плейсхолдере Order (D-01/D-05, no-phantom)"
    - "Вертикальный срез до wiring (D-02): struct+Execute → порты → composition root (app) → gRPC-адаптер (api)"

key-files:
  created:
    - knowledge/patterns.md
  modified: []

key-decisions:
  - "patterns.md ссылается на architecture.md (14 ссылок) и style.md (7 ссылок) за правилами; правила НЕ дублирует (pointer-over-copy, D-04)"
  - "Все 5 Go-сниппетов помечены «целевой вид / иллюстрация — НЕ из компилируемого файла» на плейсхолдере Order; единственное упоминание services/.../order.go — явный no-phantom disclaimer (D-01)"
  - "4 рецепта поданы как вертикальный срез до wiring (struct+Execute → порты → app → api), а не механика одного слоя (D-02)"

patterns-established:
  - "Рецепт-док как how-to-каталог: каждый рецепт = тег силы + нумерованные шаги + помеченный иллюстративный сниппет + ссылки на канон правил"

requirements-completed: [PAT-01]

# Metrics
duration: 4min
completed: 2026-06-17
---

# Phase 3 Plan 04: patterns.md Summary

**knowledge/patterns.md — 4 копируемых рецепта (add use case / query / aggregate / repository) вертикальным срезом до wiring, ссылающихся на architecture.md и style.md без дублирования правил**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-06-17
- **Completed:** 2026-06-17
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Создан `knowledge/patterns.md` (187 строк, ≤200): шапка-абзац «правила vs рецепты» + 4 рецепта.
- Каждый рецепт — вертикальный срез до wiring (D-02): `struct + Execute` → порты/репозиторий → composition root (`app`, ручной DI) → gRPC-адаптер в `api`.
- Покрыты все 4 рецепта PAT-01 (D-03): add use case (write-side), add query (read-side DTO), add aggregate (фабрика + `PullEvents`), add repository (Mongo-реализация порта + UnitOfWork + outbox в той же tx).
- Pointer-over-copy (D-04): 14 ссылок на architecture.md за архитектурными правилами, 7 ссылок на style.md за языком/ошибками/DTO→домен; правила не продублированы.
- No-phantom (D-01/D-05): 5 иллюстративных сниппетов на едином плейсхолдере `Order`, каждый помечен «целевой вид / иллюстрация — НЕ из компилируемого файла»; нет ссылок на несуществующие пути как на реальные.

## Task Commits

Each task was committed atomically:

1. **Task 1: Написать knowledge/patterns.md (4 рецепта, вертикальный срез, ссылки на architecture.md)** - `a401b2a` (feat)

**Plan metadata:** см. финальный docs-коммит.

## Files Created/Modified
- `knowledge/patterns.md` - Копируемый каталог рецептов add use case/query/aggregate/repository; вертикальный срез до wiring; ссылки на architecture.md (правила) и style.md (язык/ошибки); иллюстративные помеченные сниппеты на плейсхолдере Order.

## Decisions Made
- Имя плейсхолдера-агрегата — `Order` (переиспользовано из style.md/architecture.md, согласно D-05); никаких host/VM/owner.
- Каждый нормативный шаг несёт тег силы (MUST/WON'T) и forward-enforcement-метку `convention-only (review-enforced)` (D-11); архитектурные/языковые правила оставлены ссылками, не скопированы.
- Финальный блок добавлен явным «WON'T возрождать CQRS-диспетчер/pkg/mediatr/TxManager» со ссылкой на architecture.md §«MUST NOT» — рецепты не открывают лазейку к снесённым подсистемам.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- `gsd-tools` не на PATH (exit 127); вызывался через `node $HOME/.claude/gsd-core/bin/gsd-tools.cjs`. State-операции выполнены этим способом.

## Manual-Only Gate (D-01) — passed before verify
Обязательный фазовый гейт (VALIDATION.md Manual-Only, WARNING-3 accept): ручной просмотр Go-сниппетов подтвердил —
- все 5 сниппетов несут метку «целевой вид / иллюстрация — НЕ из компилируемого файла»;
- используется единый плейсхолдер `Order` (вне домена gwall-e);
- рецепты ссылаются на architecture.md/style.md за правилами, не выдают сниппеты за существующий код;
- единственное упоминание `services/.../order.go` — явный no-phantom disclaimer («таких файлов нет»).

## Verification
- Automated (PLAN verify one-liner): зелёный — presence + 4 рецепта + метки иллюстрации + authoring-теги.
- `architecture.md` refs: 14; `style.md` refs: 7 (pointer-over-copy).
- Размер: 187 строк (≤ ~200).
- Link integrity: architecture.md / style.md / authoring.md / boundaries.md — все существуют.
- Vertical-slice grep (`composition root|app|gRPC|api`): зелёный.

## Next Phase Readiness
- PAT-01 удовлетворён; `patterns.md` создан и согласован с architecture.md/style.md.
- Wave 5 (индекс/карта владения): снять статус «запланировано» у `patterns.md` в `knowledge/README.md` и добавить строку в карту владения `knowledge/boundaries.md` (рецепты → patterns.md). Эти интеграционные правки — вне скоупа этого плана.
- Phase 4 (ENF-05): forward-метки `convention-only (review-enforced)` готовы к переключению на фактический enforcement-статус.

## Self-Check: PASSED

- FOUND: knowledge/patterns.md
- FOUND: .planning/phases/03-conventions-architecture-docs/03-04-SUMMARY.md
- FOUND commit: a401b2a

---
*Phase: 03-conventions-architecture-docs*
*Completed: 2026-06-17*
