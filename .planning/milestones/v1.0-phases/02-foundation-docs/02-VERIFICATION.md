---
phase: 02-foundation-docs
verified: 2026-06-17T00:00:00Z
status: gaps_found
score: 3/4 must-haves verified
overrides_applied: 0
gaps:
  - truth: "knowledge/build.md даёт команды сборки/запуска/тестов, включая GOWORK=off для inventory, cd pkg && go test, фронтенд npx nx"
    status: failed
    reason: "Два задокументированных рецепта сборки падают с exit code 1. build.md утверждает 'реально проверенные рецепты' и запрещает phantom-команды (**WON'T** документировать то, что не запускается). Нарушение: (1) `cd services/audit && go build ./...` — exit 1 ('go: build output \"cmd\" already exists and is a directory'); build.md line 26-27 заявляет 'сборка проходит'. (2) `cd services/inventory && GOWORK=off go build ./...` — тот же сбой exit 1. DOC-02 Success Criteria требует 'команды сборки ... включая GOWORK=off для inventory' — оба рецепта некорректны."
    artifacts:
      - path: "knowledge/build.md"
        issue: "Line 26-27: 'cd services/audit && go build ./...' — документирована как 'сборка проходит', фактически exit code 1 ('build output cmd already exists and is a directory'). Line 40: 'cd services/inventory && GOWORK=off go build ./...' — аналогичный сбой exit 1. Рабочие альтернативы (exit 0): 'cd services/audit && go build ./cmd', 'cd services/audit && go vet ./...' или 'go build -o /tmp/audit ./...'; для inventory аналогично."
    missing:
      - "Заменить 'go build ./...' в разделе audit на верифицированную команду без коллизии с директорией cmd/ (напр. 'cd services/audit && go build ./cmd' или 'go vet ./...'). Проверить exit code перед обновлением дока."
      - "Обновить WIP-рецепт inventory аналогично: 'cd services/inventory && GOWORK=off go build ./cmd' (или GOWORK=off go vet ./...) — сохранив WIP-пометку."
      - "Скорректировать сопутствующую прозу (line 26-27 'сборка проходит') чтобы соответствовать реально работающей команде."
---

# Phase 02: Стабильные доки-основы — Verification Report

**Phase Goal:** Команда/ИИ имеют стабильные инфра/процесс-доки навигации — раскладка репозитория, команды сборки/тестов, git-конвенции и границы «не трогать» (без описания доменной модели)
**Verified:** 2026-06-17
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `knowledge/structure.md` описывает раскладку `go.work` и какие модули в/вне workspace (статус `inventory` как WIP) на уровне возможностей, без хрупкой карты путей | VERIFIED | Файл существует (57 строк), содержит `go.work`, называет ровно три workspace-члена (pkg/analytics/audit), помечает inventory как WIP вне workspace, ссылается на build.md для команд, не перечисляет internal/*. gateway/outgate отсутствуют в документе. |
| 2 | `knowledge/build.md` даёт команды сборки/запуска/тестов, включая `GOWORK=off` для `inventory`, `cd pkg && go test`, фронтенд `npx nx` | **FAILED** | `cd pkg && go test ./...` — exit 0 (OK). `npx nx` — инфра-справка (непроверяема без Nx, уровень-рецепта). НО: `cd services/audit && go build ./...` — exit 1 ('build output "cmd" already exists and is a directory'); build.md line 27 утверждает "сборка проходит". `cd services/inventory && GOWORK=off go build ./...` — exit 1 (то же). build.md декларирует **WON'T** документировать phantom-команды — нарушено для audit (жёсткое заявление "сборка проходит") и дополнительно для inventory (WIP-пометка частично покрывает, но команда всё равно документирована как рецепт). |
| 3 | `knowledge/git.md` фиксирует git-конвенции: ветки, Conventional Commits, нормы PR, когда коммитить | VERIFIED | Файл существует (61 строка, < 150). Содержит ветки dev/main, remote neeeekto/gwall-e, Conventional Commits, нормы PR, правило «когда коммитить», краткий GSD-блок (Co-Authored-By, gsd-pr-branch). Все MUST/SHOULD/WON'T тегированы. |
| 4 | `knowledge/boundaries.md` содержит правила «do-not»: не чинить/не расширять WIP-леса; стале `README`/`Makefile`/`docker-compose.yml` не авторитетны; не документировать несуществующие фичи | VERIFIED | Файл существует (71 строка, < 150). Содержит WON'T-правила для WIP-лесов (со ссылкой на structure.md), стале-файлов (README/Makefile/docker-compose), phantom-фич, .planning/ не-канон, не-раздувать CLAUDE.md. Карта владения фактами присутствует. Взаимная пара boundaries.md <-> structure.md замкнута. |

**Score:** 3/4 truths verified

### Deferred Items

Нет — все четыре Success Criteria входят в скоуп Phase 2.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `knowledge/structure.md` | Раскладка go.work на уровне модулей | VERIFIED | 57 строк, go.work + WIP inventory + web/ Nx. Нет gateway/outgate, нет internal/*. |
| `knowledge/build.md` | Команды сборки/тестов (реально работающие) | **STUB (partial)** | Файл существует, `cd pkg && go test ./...` верифицирован (exit 0). Рецепты сборки для audit и inventory документированы как рабочие, фактически exit 1. Нарушение no-phantom. |
| `knowledge/git.md` | Git-конвенции + GSD-блок | VERIFIED | 61 строка, все требования выполнены. |
| `knowledge/boundaries.md` | Правила do-not + карта владения фактами | VERIFIED | 71 строка, все правила, карта, взаимные ссылки. |
| `knowledge/README.md` | Индекс с ссылками на все 4 новых дока | VERIFIED | structure.md, build.md, git.md, boundaries.md — все со ссылками и статусом «существует». glossary.md помечен «отложено (v2 / domain-milestone)». |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `knowledge/README.md` | `knowledge/structure.md` | `[structure.md](structure.md)` | WIRED | Ссылка существует в индексной таблице и в «Порядке чтения» |
| `knowledge/README.md` | `knowledge/build.md` | `[build.md](build.md)` | WIRED | Ссылка существует в индексной таблице и в «Порядке чтения» |
| `knowledge/README.md` | `knowledge/git.md` | `[git.md](git.md)` | WIRED | Ссылка существует, статус «существует» |
| `knowledge/README.md` | `knowledge/boundaries.md` | `[boundaries.md](boundaries.md)` | WIRED | Ссылка существует, статус «существует» |
| `knowledge/structure.md` | `knowledge/build.md` | относительная ссылка | WIRED | Ссылка присутствует трижды (build.md как канон команд) |
| `knowledge/structure.md` | `knowledge/boundaries.md` | `[boundaries.md](boundaries.md)` | WIRED | Рабочая ссылка в разделе inventory WIP |
| `knowledge/boundaries.md` | `knowledge/structure.md` | `[structure.md](structure.md)` | WIRED | Ссылка в разделе WIP-лесов (канон WIP-статуса inventory) |
| `knowledge/boundaries.md` | `knowledge/build.md` | `[build.md](build.md)` | WIRED | Ссылка в разделе WIP-лесов и карте владения |

Финальная проверка битых ссылок: все relative-ссылки во всех пяти доках (README.md, structure.md, build.md, git.md, boundaries.md) ведут на существующие файлы — PASS.

### Data-Flow Trace (Level 4)

Неприменимо — фаза производит только markdown-документацию, нет компонентов с рендерингом динамических данных.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| `cd pkg && go test ./...` (задокументирован как зелёный) | `cd /Users/yymoroz3/Projects/personal/gwall-e/pkg && go test ./...` | exit 0, `ok github.com/gwall-e/pkg/http (cached)` | PASS |
| `cd services/audit && go build ./...` (задокументирован как "сборка проходит") | `cd /Users/yymoroz3/Projects/personal/gwall-e/services/audit && go build ./...` | exit 1: `go: build output "cmd" already exists and is a directory` | **FAIL** |
| `cd services/inventory && GOWORK=off go build ./...` (задокументирован как WIP-рецепт) | `cd /Users/yymoroz3/Projects/personal/gwall-e/services/inventory && GOWORK=off go build ./...` | exit 1: `go: build output "cmd" already exists and is a directory` | **FAIL** |

Корневая причина (детерминированная, не flakiness): оба сервиса имеют единственный пакет в директории `cmd/`. `go build ./...` с одним резолвенным пакетом пытается записать бинарник `cmd` в cwd, что конфликтует с существующей директорией `cmd/`. Рабочие альтернативы (exit 0): `go build ./cmd`, `go vet ./...`, `go build -o /tmp/audit ./...`.

### Probe Execution

Специальных probe-скриптов для этой фазы нет — фаза документационная. Spot-checks выше являются эквивалентом.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| DOC-01 | 02-01-PLAN.md | `knowledge/structure.md` — раскладка go.work, модули в/вне workspace, WIP inventory | SATISFIED | structure.md существует, содержит все требуемые элементы, factual claims верифицированы |
| DOC-02 | 02-01-PLAN.md | `knowledge/build.md` — команды сборки/тестов, GOWORK=off, cd pkg && go test, npx nx | **BLOCKED** | build.md существует, `cd pkg && go test` работает. Но `go build ./...` для audit (с жёстким заявлением "сборка проходит") и inventory failing exit 1. Нарушает stated premise DOC-02 "только верифицированные, реально запускаемые команды". |
| DOC-06 | 02-02-PLAN.md | `knowledge/git.md` — ветки, Conventional Commits, нормы PR, когда коммитить | SATISFIED | git.md содержит все требуемые элементы; factual claims (ветки dev/main, remote neeeekto/gwall-e) верифицированы |
| DOC-08 | 02-03-PLAN.md | `knowledge/boundaries.md` — правила do-not (WIP-леса, стале-файлы, phantom) | SATISFIED | boundaries.md содержит все три основных правила, оба посева Phase 1, карту владения |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `knowledge/boundaries.md` | 13 | `TODO` | Info | Слово `TODO` является частью прозы описывающей код-паттерны WIP-сервисов (`nil-deps, \`TODO\`, пустые internal/-каталоги`) — это не маркер незавершённости в самом документе, а описание признаков WIP-кода. Не является BLOCKER. |
| `knowledge/build.md` | 26-27, 40 | Документированные команды с exit code 1 | **BLOCKER** | `go build ./...` для audit и inventory возвращают exit 1. build.md декларирует WON'T для phantom-команд. Прямое нарушение stated premise DOC-02. |

**Blocker anti-patterns: 1.** Unreferenced TBD/FIXME/XXX: отсутствуют.

### Human Verification Required

Нет — все четыре Success Criteria верифицируемы программно. Статус gaps_found из-за подтверждённого провала команд.

### WR-01 Assessment (из code review): root-level go.mod не задокументирован

Репозиторий содержит корневой `go.mod` (`module github.com/gwall-e`, go 1.23.6), который не входит в `go.work` и не упомянут в `structure.md`. Также присутствуют `agents/informator/go.mod`, `agents/skytor/go.mod`, `services/outgate/go.mod`, `services/gateway/go.mod` — ни один не задокументирован в structure.md.

Оценка: `structure.md` заявляет "Активные модули workspace — ровно три" — это фактически верно применительно к go.work. Но doc позиционирует себя как "карта репозитория на уровне модулей" — пропуск корневого go.mod (с реальным go.sum и зависимостями) и agents/ создаёт неполноту. Это WARNING (как и оценил review), не BLOCKER сам по себе.

Однако этот WARNING не затрагивает Success Criteria Phase 2 (которые описывают workspace/inventory WIP, а не исчерпывающий список всех модулей). Поэтому он не выводит DOC-01 в FAILED — DOC-01 остаётся SATISFIED. Дефект фиксируется как область улучшения.

### IN-01 Assessment (из code review): непоследовательный scope в примере

`git.md` line 23 использует образец `docs(02-01):` (sub-phase scope), тогда как GSD-блок line 57 предписывает `docs(NN):`. Scope формы не согласованы в одном документе. Это INFO — функциональность git.md не нарушена, DOC-06 SATISFIED.

### Gaps Summary

Один критический gap блокирует phase goal:

**CR-01 (BLOCKER): build.md документирует команды, которые падают с exit 1.**

Центральное обещание `build.md` — "реально работающие команды" с явным WON'T против phantom-команд — нарушено:

1. `cd services/audit && go build ./...` → exit 1 (детерминированно, не WIP/flakiness). build.md line 26-27: "сборка проходит" — фактически неверно.
2. `cd services/inventory && GOWORK=off go build ./...` → exit 1. Для inventory WIP-пометка частично покрывает ожидаемость сбоев, но команда всё равно документирована в разделе с MUST-рецептом без disclamer что exit 1 — норма.

Корневая причина: каждый модуль содержит единственный пакет в директории `cmd/`. `go build ./...` пытается эмитировать бинарник с именем `cmd` в cwd, что конфликтует с существующей директорией. Детерминированный сбой на любом чистом дереве.

Рабочие альтернативы (все exit 0, подтверждено):
- `cd services/audit && go build ./cmd`
- `cd services/audit && go vet ./...`
- `cd services/audit && go build -o /tmp/audit ./...`
- Для inventory: те же с `GOWORK=off` prefix

DOC-02 Success Criteria прямо указывает на команды сборки как must-have. Три из четырёх задокументированных основных рецептов работают; один (audit build) с жёстким заявлением о прохождении — не работает.

Остальные три docs (structure.md, git.md, boundaries.md) полностью соответствуют своим Success Criteria. README индекс синхронизирован. Все relative-ссылки валидны. Все нормативные правила тегированы MUST/SHOULD/WON'T. Все git-коммиты верифицированы.

---

_Verified: 2026-06-17_
_Verifier: Claude (gsd-verifier)_
