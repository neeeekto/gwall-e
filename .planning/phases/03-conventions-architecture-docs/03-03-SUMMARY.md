---
phase: 03-conventions-architecture-docs
plan: 03
subsystem: knowledge-base
tags: [docs, architecture, ddd, hexagonal, no-cqrs]
requires:
  - knowledge/style.md  # 03-01: typed IDs / errors / DTO→домен cross-links
  - knowledge/authoring.md  # authoring standard
provides:
  - knowledge/architecture.md  # канон целевой архитектуры (DOC-05)
affects:
  - knowledge/patterns.md  # 03-04 будет ссылаться сюда за правилами
tech-stack:
  added: []
  patterns: [DDD, hexagonal-architecture, interactor-Execute, query-lite, UnitOfWork-port, transactional-outbox, relay, aggregate-factory, PullEvents]
key-files:
  created:
    - knowledge/architecture.md
  modified: []
decisions:
  - "architecture.md держит инварианты/why (правила); how-to рецепты — в patterns.md (D-04, pointer-over-copy)"
  - "Файл не дробился: 148 строк, в пределах ~80–200; блок событий оставлен в основном файле"
  - "Цель depguard-меток — planned CI-gated Phase 4; review-enforced правила — convention-only (D-11)"
metrics:
  duration: ~4m
  completed: 2026-06-17
---

# Phase 3 Plan 3: architecture.md (DDD + гексагон без CQRS) Summary

Создан `knowledge/architecture.md` — канон целевой архитектуры gwall-e (DOC-05): DDD +
гексагональная архитектура БЕЗ CQRS-шины, с правилами слоёв, текстовой диаграммой +
таблицей направления импортов, инвариантами (Execute, query-lite, порт UnitOfWork,
outbox+relay, фабрики + PullEvents) и явным MUST NOT возрождать CQRS-диспетчер /
pkg/mediatr / TxManager.

## What Was Built

- **knowledge/architecture.md** (148 строк, NEW): канон целевой архитектуры.
  - Шапка явно заявляет «гексагональная архитектура БЕЗ CQRS-шины»; рецепты — ссылка на
    patterns.md (pointer-over-copy), правила языка/ошибок — ссылка на style.md.
  - **Слои и импорты (D-06):** MUST зависимости внутрь на `domain`; `domain` не импортирует
    наружу (только порты). Текстовая ASCII-диаграмма + таблица «кто кого импортирует»
    (`api`/`repositories`/`query`/`cron` → `domain`; `app`=composition root ручной DI;
    `cmd`=main). Enforcement: planned CI-gated Phase 4 (depguard).
  - **Write-side (Execute):** MUST 1 use case = struct + `Execute(ctx, in) (out, error)`;
    маппинг DTO→домен — ссылка на style.md; иллюстративный `RegisterOrderUseCase`.
  - **UnitOfWork:** MUST транзакционная граница через порт `UnitOfWork` (`Do(ctx, fn)`) в
    `domain`; агрегат + outbox в одной tx.
  - **Read-side:** MUST query-lite напрямую в DTO (CQRS-lite, мимо агрегатов).
  - **События:** MUST фабрики агрегатов + `PullEvents()` → transactional outbox в той же tx
    + async relay (нет dual-write, at-least-once).
  - **Валидация на edge:** SHOULD (protovalidate в api/), домен не валидирует транспорт —
    помечено как конвенция, не реализация (security no-phantom).
  - **MUST NOT CQRS:** WON'T возрождать CommandDispatcher/QueryDispatcher / pkg/mediatr /
    TxManager / tx.go — парность запрет→do (inbound зовёт use case напрямую + порт UnitOfWork).
  - Все Go-сниппеты помечены «целевой вид / иллюстрация — НЕ из компилируемого файла» на
    плейсхолдере `Order` (D-01, D-05); нет ссылок на несуществующие пути.

## Verification

DOC-05 presence (дискретные asserted-grep, без for-loop/||-маскировки — WARNING-2):
все зелёные (`гексагон`, `CQRS`, `MUST NOT|WON'T`, `Execute`, `UnitOfWork`, `outbox`,
`PullEvents`, `query`, диаграмма импортов, метка иллюстрации, `mediatr|TxManager|диспетчер`,
enforcement-метки). `UnitOfWork` встречается 7 раз (artifact `contains` удовлетворён).
Размер 148 строк (≤ ~200, дробление не требуется). Cross-link на style.md присутствует
(2 ссылки, key_link удовлетворён). Все markdown-ссылки ведут на существующие файлы.

**Manual-Only гейт (обязательный, до verify — WARNING-3 accept):**
- D-06 диаграмма: пройдено — `domain` без исходящих стрелок («— (никого)»), все остальные
  слои направлены внутрь к `domain`; направление корректно.
- D-01 метки «иллюстрация»: пройдено — оба Go-сниппета на плейсхолдере `Order` помечены и
  не выдаются за существующий код; нет phantom-путей.
- Теги силы: каждое нормативное правило несёт ровно один MUST/SHOULD/WON'T, без хеджирования.

## Deviations from Plan

None — plan executed exactly as written. Файл не дробился (148 строк в пределах целевого
~80–200), так что выделение блока событий в architecture-events.md не понадобилось.

## Out-of-Scope / Deferred (Wave 5 cross-cutting)

`files_modified` плана — только `knowledge/architecture.md`. Регистрация architecture.md в
ownership-map `boundaries.md` и обновление индекса `README.md` (с «запланировано» на ссылку)
— это cross-cutting задачи финального Wave 5 (см. 03-VALIDATION.md Cross-Cutting), не входят
в scope этого плана; не трогались здесь.

## Known Stubs

None — документационный артефакт, без кода/данных. Все снищеты явно помечены как
иллюстративные плейсхолдеры (это не stubs, а нормированный no-phantom формат).

## Self-Check: PASSED

- FOUND: knowledge/architecture.md
- FOUND: .planning/phases/03-conventions-architecture-docs/03-03-SUMMARY.md
- FOUND: commit 74923ab
