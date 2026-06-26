# Phase 2: Стабильные доки-основы - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-17
**Phase:** 2-Стабильные доки-основы
**Areas discussed:** Глоссарий (→ декскоуп), structure.md, git.md + boundaries.md

---

## Выбор областей для обсуждения

| Option | Description | Selected |
|--------|-------------|----------|
| Глоссарий: глубина и forward-looking | Стабильные термины vs forward-looking фичи | ✓ |
| structure.md: глубина слоёв + stale-каталоги | Модули/workspace vs целевые слои; трактовка stale-dirs | ✓ |
| build.md: какие команды документировать | Рабочие vs WIP/frontend с пометками | |
| git.md + boundaries.md: GSD-специфика и do-not | Глубина GSD; конкретность WIP; кросс-док владение | ✓ |

**User's choice:** structure.md, Глоссарий, git.md + boundaries.md (build.md — на усмотрение Claude)

---

## Глоссарий (исходное обсуждение → переосмыслено в декскоуп)

| Option | Description | Selected |
|--------|-------------|----------|
| Ядро + роли/доступ | host, VM, owner, SRE, ITDC, namespace, project + роли | ✓ (затем отменено) |
| Ядро + forward-looking фичи | + actions/autorepair/массовые работы | |
| Минимум по коду | только встречающееся в коде сейчас | |

Глубина: «Термин + EN/RU + связи/инварианты» (✓). Роли: «в глоссарии» (✓).

**Переосмысление скоупа (вопрос пользователя):** «Зачем сейчас обсуждаем глоссарий? Цель —
сделать базу для ИИ; описывать систему будем в отдельном Milestone.»

| Option | Description | Selected |
|--------|-------------|----------|
| Отложить в domain-milestone | Убрать glossary.md из Phase 2; домен фиксируем при проектировании системы | ✓ |
| Мини-глоссарий без бизнес-домена | Только инфра/архитектурные термины (UnitOfWork, outbox, port...) | |
| Оставить как в роадмапе | Полный доменный глоссарий сейчас | |

**User's choice:** Отложить glossary.md (DOC-07) в domain-milestone.
**Notes:** Доменная модель не спроектирована — фиксация сейчас = риск расхождения; противоречит цели
milestone'а (правила для ИИ, не описание системы). ROADMAP/REQUIREMENTS обновлены.

---

## structure.md

| Option | Description | Selected |
|--------|-------------|----------|
| Только модули/workspace | Раскладка на уровне модулей; слои → architecture.md | ✓ |
| Модули + карта слоёв со ссылкой | + абзац о слоях с указателем | |
| Полная слоистая раскладка здесь | Слои сервиса прямо в structure.md | |

**User's choice:** Подготовить инфра-базу без доменных доков; детали системы — следующий Milestone
(= модули/workspace уровень).
**Notes:** Stale-каталоги `inventory/internal/` → «WIP-скелет, не авторитетен» (✓).
gateway/outgate → «не упоминать» (✓).

---

## git.md + boundaries.md

| Option | Description | Selected |
|--------|-------------|----------|
| git.md: Общие + кратко GSD | Conventional Commits/ветки/PR + краткий блок GSD-практик | ✓ |
| git.md: Только общие конвенции | без GSD-поведения | |
| git.md: Полно GSD | весь GSD git-workflow | |

| Option | Description | Selected |
|--------|-------------|----------|
| boundaries.md: Возможности + примеры | правила на уровне возможностей, без хрупких путей | ✓ |
| boundaries.md: Именованный список путей | явные пути | |

| Option | Description | Selected |
|--------|-------------|----------|
| Явная карта владения | таблица «факт → канон» (pointer-over-copy) | ✓ |
| Без формальной карты | следовать pointer-over-copy по ходу | |

**User's choice:** git.md — общие + кратко GSD; boundaries.md — возможности+примеры; явная карта владения.

---

## Claude's Discretion

- **build.md (DOC-02)** — не выбрана для детального обсуждения; наполнение на усмотрение планировщика
  в рамках принципов сессии (реально рабочие команды + WIP-пометки для `inventory`, без phantom-команд).
- Точное расположение «карты владения» (boundaries.md vs README.md).
- Дробление доков при превышении ~150–200 строк (authoring.md).

## Deferred Ideas

- `glossary.md` (DOC-07) → domain-milestone (доменный ubiquitous language).
- Целевая слоистая раскладка сервиса → architecture.md (Phase 3) / domain-milestone.
- Полный GSD git-workflow — не дублируем в git.md.
