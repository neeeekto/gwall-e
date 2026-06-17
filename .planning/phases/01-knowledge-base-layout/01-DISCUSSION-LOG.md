# Phase 1: Раскладка базы знаний и точки входа - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-17
**Phase:** 1-Раскладка базы знаний и точки входа
**Areas discussed:** Структура knowledge/, CLAUDE.md и GSD-блоки, CLAUDE.md vs AGENTS.md, Формат authoring-стандарта

---

## Структура knowledge/ — раскладка

| Option | Description | Selected |
|--------|-------------|----------|
| Плоско | Всё в `knowledge/*.md`, навигация через README | ✓ |
| Подпапки сразу | decisions/, conventions/ и т.д. | |

**User's choice:** Плоско
**Notes:** Подпапки добавим при росте числа файлов.

## Структура knowledge/ — имена файлов

| Option | Description | Selected |
|--------|-------------|----------|
| kebab-case, англ. | `architecture.md`, `style.md`, `glossary.md` | ✓ |
| Русские/транслит | кириллица/транслит в именах | |

**User's choice:** kebab-case, английские

## Структура knowledge/ — язык содержимого

| Option | Description | Selected |
|--------|-------------|----------|
| Русский | Как комментарии кода; термины EN через glossary | ✓ |
| Английский | Всё на английском | |
| Смешанно | RU проза + EN тех-термины | |

**User's choice:** Русский

## Структура knowledge/ — формат README

| Option | Description | Selected |
|--------|-------------|----------|
| Таблица | Файл \| Назначение \| Когда читать + порядок чтения | ✓ |
| Простой список | Маркированный список ссылок | |

**User's choice:** Таблица

---

## CLAUDE.md и GSD-блоки

| Option | Description | Selected |
|--------|-------------|----------|
| Тонкий гибрид | Шапка + ссылки на knowledge/ + GSD workflow-блок; тяжёлый stack/architecture-дамп убрать | ✓ |
| Чистый минимум | Только ручной индекс; GSD-автогенерацию отключить | |
| Оставить + ссылки | Не урезать, добавить ссылки (нарушает <150 строк) | |

**User's choice:** Тонкий гибрид
**Notes:** GSD workflow-блок логично оставить в CLAUDE.md (специфичен для Claude Code/GSD). Учтён риск повторного раздувания через generate-claude-md.

---

## CLAUDE.md vs AGENTS.md — источник истины

| Option | Description | Selected |
|--------|-------------|----------|
| CLAUDE.md — источник | AGENTS.md — тонкий указатель | |
| Symlink | AGENTS.md → symlink на CLAUDE.md | |
| AGENTS.md — источник | Кросс-тульный стандарт; CLAUDE.md — указатель | ✓ |

**User's choice:** AGENTS.md — источник истины

---

## Формат authoring-стандарта

| Option | Description | Selected |
|--------|-------------|----------|
| Отдельный файл | `knowledge/authoring.md` + памятка в README | ✓ |
| Секция в README | Стандарт как секция README.md | |

**User's choice:** Отдельный файл

---

## Claude's Discretion

- Точные колонки/формулировки таблицы README сверх минимума.
- Конкретная разметка силы правил (жирный префикс vs бейджи) — в `authoring.md`.
- Содержание шапки AGENTS.md (из PROJECT.md).
- Создавать ли стабы будущих доков или перечислять со статусом «запланировано» (без битых ссылок).

## Deferred Ideas

- ADR-доки, anti-patterns.md, libraries.md, onboarding, maintenance-протокол — v2.
- Подпапки внутри knowledge/ — при росте.
- Наполнение topic-доков контентом — Phase 2–3.
- Пункт «что НЕ кладём в knowledge/» (boundaries.md, Phase 2) + карта «что где живёт» (README, Phase 1).
