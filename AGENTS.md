# gwall-e — гайд для агентов

**AGENTS.md — источник истины** для всех ИИ-агентов (Codex, Cursor, Copilot, Gemini,
Aider и др.) и команды. Тонкий кросс-тульный вход: шапка проекта, навигация в базу знаний
и указатель на стандарт авторинга. Детали — в `knowledge/*.md` (progressive disclosure),
здесь они не дублируются.

## Что это

gwall-e — платформа **Hardware-as-a-Service** для дата-центров: инвентаризация хостов и VM,
просмотр их состояния, выдача SSH-прав, действия над хостами (в т.ч. массовые) и автопочинка —
с сохранением согласованности (никто не может «забрать» чужой хост в обход правил).

Технически — набор Go-микросервисов (бэкенд, DDD + гексагональная архитектура, без CQRS-шины)
и React/Nx фронтенд.

**Core Value:** безопасное и согласованное управление парком серверов как услугой — единый
источник правды о хостах и контролируемый, неконфликтный доступ к действиям между овнерами и
SRE/ITDC.

## Стандарт авторинга

Правила в `knowledge/*.md` тегируются **MUST** / **SHOULD** / **WON'T**, и каждый запрет
сопровождается предписанной альтернативой («do»). Полный стандарт — в
[knowledge/authoring.md](knowledge/authoring.md).

## База знаний

Канон правил проекта живёт в `knowledge/`. Начни навигацию с индекса
[knowledge/README.md](knowledge/README.md) — там полный перечень и порядок чтения.

| Док | О чём | Статус |
|-----|-------|--------|
| [knowledge/README.md](knowledge/README.md) | Индекс базы знаний: карта и порядок чтения | есть |
| [knowledge/authoring.md](knowledge/authoring.md) | Стандарт авторинга: как писать правила | есть |
| [knowledge/structure.md](knowledge/structure.md), [knowledge/build.md](knowledge/build.md), [knowledge/git.md](knowledge/git.md), [knowledge/boundaries.md](knowledge/boundaries.md) | Раскладка, сборка, git, границы | есть |
| `glossary.md` | Домен: ubiquitous language (EN/RU) | отложено (domain-milestone) |
| [knowledge/style.md](knowledge/style.md), [knowledge/testing.md](knowledge/testing.md), [knowledge/architecture.md](knowledge/architecture.md), [knowledge/patterns.md](knowledge/patterns.md) | Стиль кода, тесты, архитектура, рецепты слоёв | есть |

Ссылки даны только на существующие файлы. Будущие доки перечислены **без ссылок** со статусом
«запланировано» — ссылка появится вместе с самим файлом.

## Точки входа

- **AGENTS.md** (этот файл) — канонический источник истины для всех агентов.
- **CLAUDE.md** — тонкий Claude-специфичный вход: ссылается сюда и несёт GSD workflow-блок;
  контент не дублирует.

## Границы (do-not)

- **WON'T** «чинить» WIP-леса в `services/inventory` — это незавершённая работа, а не баг;
  если что-то не собирается, сначала свериться с правилами, а не патчить наугад.
- **WON'T** доверять корневым `README.md`, `Makefile`, `docker-compose.yml` — они устаревшие
  и неавторитетны; источник истины — этот файл и `knowledge/`. Полное правило — в
  [knowledge/boundaries.md](knowledge/boundaries.md).
