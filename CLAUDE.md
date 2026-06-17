<!-- ВНИМАНИЕ: CLAUDE.md намеренно урезан до тонкого гибрида (Phase 1, KB-01/D-06).
     НЕ запускать `gsd-tools query generate-claude-md` — он повторно раздует удалённые
     секции stack/architecture/conventions/skills из research. Детали — в knowledge/.
     Правило закрепляется в knowledge/boundaries.md (Phase 2). -->

# gwall-e

gwall-e — платформа **Hardware-as-a-Service** для дата-центров: инвентаризация хостов и VM,
SSH-права, действия над хостами (в т.ч. массовые) и автопочинка с сохранением согласованности.

**Core Value:** безопасное и согласованное управление парком серверов как услугой.

## Точки входа

**Источник истины для конвенций — [AGENTS.md](AGENTS.md)** (канонический кросс-тульный вход).
Детальные правила — в `knowledge/` (progressive disclosure); начни с
[knowledge/README.md](knowledge/README.md). Этот файл не дублирует их контент — только
ссылается; он несёт лишь Claude-специфичный GSD workflow-блок ниже.

<!-- GSD:project-start source:PROJECT.md -->

## Project

**gwall-e** — Go-микросервисы (бэкенд, DDD + гексагональная архитектура, без CQRS-шины) и
React/Nx фронтенд. Подробности проекта, Core Value, tech stack и конвенции — в
[AGENTS.md](AGENTS.md) и [knowledge/](knowledge/README.md); здесь не повторяются (D-07).

<!-- GSD:project-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:

- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
