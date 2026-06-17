---
phase: 03-conventions-architecture-docs
plan: 01
subsystem: knowledge-base
tags: [docs, conventions, go-style, language, errors, dto]
requires: [knowledge/authoring.md, knowledge/boundaries.md, knowledge/structure.md, knowledge/build.md]
provides:
  - "knowledge/style.md — канон языка кода/комментариев, typed IDs, sentinel-vs-wrapped errors (%w), DTO→домен в хендлере"
  - "Нейтральный плейсхолдер Order, зафиксированный для downstream architecture.md/patterns.md (D-05)"
affects:
  - "testing.md (02) — будет ссылаться сюда за правилом EN-комментариев"
  - "architecture.md (03) / patterns.md (04) — переиспользуют плейсхолдер Order и ссылаются на правила typed ID/errors/DTO→домен"
  - "boundaries.md L67 / README.md / AGENTS.md — индекс/карта владения (правка в плане 05)"
tech-stack:
  added: []
  patterns:
    - "Правило + плохо/хорошо (D-08)"
    - "Forward enforcement-метки (D-11): convention-only / planned: CI-gated Phase 4 / planned: hook Phase 4"
    - "Pointer-over-copy (общий Go-стиль → gofumpt/Effective Go; команды → build.md; архитектура → architecture.md)"
key-files:
  created:
    - knowledge/style.md
  modified: []
decisions:
  - "D-07: охват только проект-специфика; общий Go-стиль не дублируется (ссылка на gofumpt Phase 4 + Effective Go)"
  - "D-08: каждое правило в формате «правило + плохо/хорошо»"
  - "D-11: механизируемые правила несут forward-enforcement-метки"
  - "D-05: нейтральный плейсхолдер Order (вне домена gwall-e) зафиксирован для downstream-доков"
metrics:
  duration: ~4min
  completed: 2026-06-17
---

# Phase 3 Plan 1: style.md (язык кода + проект-специфичный Go-стиль) Summary

Создан `knowledge/style.md` — единственный канон правила про язык кода/комментариев
(RU-комментарии/доменная терминология, EN-имена идентификаторов, EN-комментарии в тестах)
и проект-специфичных правил Go-стиля gwall-e: типизированные ID, sentinel-vs-обёрнутые
ошибки (`%w`), маппинг DTO→домен внутри хендлера. Каждое правило — в формате «правило +
плохо/хорошо» с forward-enforcement-меткой; общий Go-стиль не дублируется (ссылка на
gofumpt Phase 4 + Effective Go).

## What Was Built

- **knowledge/style.md** (110 строк) со структурой:
  - Шапка-абзац: позиционирование как единственного канона языка; общий Go-стиль не
    дублируется (ссылка на Effective Go + gofumpt planned Phase 4); ссылка на authoring.md
    за форматом; явная no-phantom-метка про иллюстративный плейсхолдер `Order`.
  - **Язык кода и комментариев** (D-07): три MUST (RU-комментарии/доменная терминология,
    EN-имена идентификаторов, EN-комментарии в тестах) + пример плохо/хорошо; явно заявлено,
    что это единственное место правила (карта владения).
  - **Типизированные ID** (D-08): MUST typed ID вместо голой строки + пример.
  - **Sentinel vs обёрнутые ошибки**: MUST sentinel + MUST `%w`-обёртка; пример плохо (`%v`
    рвёт цепочку) → хорошо (`%w`).
  - **Маппинг DTO→домен внутри хендлера**: MUST decode/validate на edge + маппинг в
    хендлере; пример плохо (DTO протекает в use case) → хорошо.
  - Все механизируемые правила с метками `convention-only (review-enforced)` /
    `planned: CI-gated Phase 4` / `planned: hook Phase 4`.
  - Все Go-сниппеты на нейтральном плейсхолдере `Order` (D-05), с no-phantom-меткой.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Написать knowledge/style.md | 4960b47 | knowledge/style.md |

## Verification

Plan automated verification — все зелёные:
- `test -f knowledge/style.md` ✓
- DOC-04 presence: `русск|коммент` ✓, `%w` ✓, `типизирован|typed` ✓, `DTO` ✓
- D-11 enforcement-метки: `convention-only|planned.*Phase 4|CI-gated|hook` ✓
- D-07 non-duplication: `gofumpt|Effective Go` ✓
- Authoring теги: `MUST|SHOULD|WON'T` ✓
- Размер: 110 строк (≤ ~200) ✓
- Нет хеджирования: `обычно|желательно|prefer` — пусто ✓
- Плейсхолдер: единственное вхождение host/VM/owner — в строке-отрицании («НЕ домен… ни
  host, ни VM, ни owner»), не как плейсхолдер-агрегат ✓

## Deviations from Plan

None — план выполнен ровно как написано.

## Notes for Downstream Plans

- Плейсхолдер `Order` (с `OrderID`, `OrderRepository`, `RegisterOrderUseCase`,
  `UnitOfWork`) зафиксирован — architecture.md (03) / patterns.md (04) MUST переиспользовать
  то же имя (D-05).
- testing.md (02) за правилом EN-комментариев в тестах MUST ссылаться сюда (карта владения),
  а не повторять.
- Карта владения в boundaries.md L67 уже резервирует style.md за этим фактом; правка
  статуса «запланировано» → «существует» в boundaries.md/README.md/AGENTS.md — в плане 05.

## Self-Check: PASSED

- FOUND: knowledge/style.md
- FOUND: commit 4960b47
