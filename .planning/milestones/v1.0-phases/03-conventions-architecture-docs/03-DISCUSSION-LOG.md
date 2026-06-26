# Phase 3: Доки конвенций и архитектуры - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-17
**Phase:** 3-Доки конвенций и архитектуры
**Areas discussed:** patterns.md глубина, architecture.md ↔ patterns.md, style.md, testing.md

---

## patterns.md — глубина рецептов

| Option | Description | Selected |
|--------|-------------|----------|
| Иллюстративные сниппеты | Реальные Go-сниппеты, явно помеченные как «целевой вид/иллюстрация» (не из компилируемого файла) | ✓ |
| Скелеты + проза | Минимальные скелеты (сигнатуры без тел) + пошаговые шаги словами | |
| Прозовые чек-листы | Только процедуры словами + ссылки на architecture.md, без кода | |

**User's choice:** Иллюстративные сниппеты
**Notes:** Эталонный сервис не компилируется (inventory снесён) — пометка снимает конфликт с no-phantom (boundaries.md), сохраняя «копируемость» PAT-01.

## patterns.md — пример для иллюстраций

| Option | Description | Selected |
|--------|-------------|----------|
| Нейтральный плейсхолдер | Абстрактный агрегат вне домена gwall-e (Order/Widget/Foo) — только механика слоёв | ✓ |
| Реальный домен (Host) | Иллюстрировать на Host/Owner — нагляднее, но предрешает доменную модель до domain-milestone | |
| Без единого примера | Каждый рецепт со своим ад-хок примером | |

**User's choice:** Нейтральный плейсхолдер
**Notes:** Согласовано с D-01 Phase 2 (домен отложен). Единый сквозной плейсхолдер для patterns.md и architecture.md.

## patterns.md — охват рецепта

| Option | Description | Selected |
|--------|-------------|----------|
| До wiring включительно | struct+Execute → порты/репо → composition root (app, DI) → gRPC-адаптер (api) | ✓ |
| Только слой | Механика своего слоя без DI/gRPC | |

**User's choice:** До wiring включительно
**Notes:** Вертикальный срез — реально копируемый путь.

---

## architecture.md ↔ patterns.md — граница

| Option | Description | Selected |
|--------|-------------|----------|
| Правила vs рецепты | architecture.md = инварианты/правила/why; patterns.md = how-to со сниппетами, ссылается, не дублирует | ✓ |
| Arch самодостаточен | architecture.md = правила + мини-примеры; patterns.md = расширенный каталог | |
| Один файл | Объединить в architecture.md | |

**User's choice:** Правила vs рецепты
**Notes:** pointer-over-copy.

## architecture.md — диаграмма импортов

| Option | Description | Selected |
|--------|-------------|----------|
| Да, таблица/ASCII | Явная текстовая карта направления импортов между слоями | ✓ |
| Только текстом | Правила импортов словами/списком MUST, без диаграммы | |

**User's choice:** Да, таблица/ASCII

---

## Сквозное — enforcement-статус (готовит ENF-05 Phase 4)

| Option | Description | Selected |
|--------|-------------|----------|
| Пред-пометка (planned) | Механизируемые правила помечать сразу (convention-only / planned: Phase 4); Phase 4 только меняет статус | ✓ |
| Без пометок | Писать без статуса; всё проставляет Phase 4 ретроактивно | |

**User's choice:** Пред-пометка (planned)

---

## style.md — охват

| Option | Description | Selected |
|--------|-------------|----------|
| Только проект-специфика | Язык, typed IDs, sentinel vs обёрнутые ошибки, DTO→домен; общий формат → gofumpt + Effective Go | ✓ |
| Полный стайлгайд | Плюс общий нейминг/формат/организация пакетов | |

**User's choice:** Только проект-специфика
**Notes:** Без bloat, не дублирует линтер Phase 4.

## style.md — формат правил

| Option | Description | Selected |
|--------|-------------|----------|
| Правило + «плохо/хорошо» | MUST/SHOULD + мини-пример плохо → хорошо (согласовано с «запрет→do») | ✓ |
| Правило + пример хорошего | MUST/SHOULD + один пример правильного кода | |
| Только правило | Только формулировка без кода | |

**User's choice:** Правило + «плохо/хорошо»

---

## testing.md — степень мандата

| Option | Description | Selected |
|--------|-------------|----------|
| MUST каркас + SHOULD структура | MUST: suite-бутстрап, англ. комментарии, тесты рядом; SHOULD: Describe/Context/It, DescribeTable | ✓ |
| Всё MUST | Жёсткий мандат на всю структуру | |
| Convention-only | Эталон pkg/http как рекомендация, без жёстких MUST | |

**User's choice:** MUST каркас + SHOULD структура

## testing.md — моки портов

| Option | Description | Selected |
|--------|-------------|----------|
| Ручные fakes | Ручные fake/stub-реализации портов в _test-пакетах | |
| Кодоген моков | Генерируемые моки из интерфейсов портов | ✓ |
| Отдать ресёрчу | Не фиксировать; researcher предложит | |

**User's choice:** Кодоген моков

## testing.md — инструмент кодогена

| Option | Description | Selected |
|--------|-------------|----------|
| Отдать ресёрчу | Сравнить mockgen/counterfeiter/moq | |
| go.uber.org/mock (mockgen) | Самый распространённый, go:generate | |
| counterfeiter | Fakes в стиле Ginkgo-экосистемы | |
| mockery (vektra/mockery) | Выбор пользователя (free text) | ✓ |

**User's choice:** mockery (`vektra/mockery`)
**Notes:** Обвязка go:generate/.mockery.yaml — Phase 4; выбор инструмента закреплён здесь.

---

## Claude's Discretion

- Конкретное имя нейтрального плейсхолдера-агрегата (вне домена gwall-e, одинаковое во всех доках).
- Форма диаграммы импортов (ASCII vs таблица).
- Разбивка дока на под-файлы при превышении ~150–200 строк.
- Конкретные формулировки/набор мини-примеров «плохо/хорошо».
- Стиль генерируемых mockery-моков (expecter vs классический) и стыковка с Gomega.

## Deferred Ideas

- `glossary.md` (DOC-07) → domain-milestone.
- Reference-service walkthrough → когда `inventory` начнёт компилироваться.
- Реальная настройка тулинга (golangci-lint/gofumpt/gci, lefthook, commitlint, buf, go:generate mockery)
  и фактический enforcement-статус (ENF-05) → Phase 4.
- ADR-доки, anti-patterns.md, libraries.md, onboarding, maintenance-протокол → v2.
