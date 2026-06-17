# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — Фундамент (knowledge base + enforcement)

**Shipped:** 2026-06-17
**Phases:** 4 | **Plans:** 14 | **Sessions:** 1 (intensive 1-day build)

### What Was Built
- База знаний `knowledge/` (10 доков, ~1101 строк): authoring-стандарт, structure/build/git/boundaries, style/testing/architecture/patterns.
- Тонкие точки входа: `AGENTS.md` (канонический источник истины) + урезанный `CLAUDE.md` (~205 → 51 строка) + индекс `knowledge/README.md`.
- Enforcement-слой: `.golangci.yml` v2 (errorlint + depguard ban на `pkg/mediatr`), `lefthook.yml`, commitlint, buf v2-скелет, `Makefile` с пиннингом версий; статус enforcement на каждом механизируемом правиле.

### What Worked
- **No-phantom дисциплина:** доки заземлены на реальный репозиторий (проверенные команды, реальные образцы коммитов), forward-метки честно помечались `planned`/`отложено` вместо выдумывания.
- **Single-source-of-truth:** язык кода живёт только в `style.md`, легенда enforcement — только в `authoring.md`; карта владения фактами в `boundaries.md` предотвращает дублирование.
- **Волновая декомпозиция фаз** (Wave 1/2/3) с явными зависимостями дала чистый порядок исполнения без конфликтов.
- **Depguard как biting ban** на `pkg/mediatr` материализовал решение «без CQRS-шины» в исполнимое правило, а не только декларацию.

### What Was Inefficient
- **DOC-02 проскочил верификацию:** `build.md` зафиксировал `cd services/audit && go build ./...` как проходящую, хотя команда падает (exit 1) — false pass-claim в доке, чей центральный принцип «только проверенные команды». Прямое нарушение собственной премиссы; должно было ловиться при authoring.
- **Frontmatter traceability gaps:** `requirements-completed` пустой в 03-01/03-03 SUMMARY (DOC-04/DOC-05) — VERIFICATION подтвердил satisfied, но 3-source cross-reference ломается.
- **Nyquist validation** не доведён до sign-off ни на одной фазе (1–2 без VALIDATION.md, 3–4 pending) — для doc-фаз «тесты» = presence/link checks, формат не довели до конца.
- **Live-firing хуков отложено:** 7 UAT-сценариев требуют one-time bootstrap (`make tools` + `lefthook install`) — тулинг отсутствовал на машине исполнителя.

### Patterns Established
- MUST/SHOULD/WON'T нормативная разметка с обязательной парной альтернативой («do») для каждого запрета.
- Pointer-over-copy (D-04): `architecture.md` держит инварианты/why, `patterns.md` — how-to рецепты; доки ссылаются, не дублируют.
- Forward-enforcement-метки в knowledge-доках, переключаемые на честный статус (hook / convention-only) при появлении тулинга.
- Версии тулинга пиннятся в `Makefile`, root `go.mod` не трогается.

### Key Lessons
1. **«Проверенная команда» в доке = реально прогнанная команда.** Любой build/test-рецепт в knowledge-доке должен исполняться при authoring, иначе он phantom (см. DOC-02).
2. **Traceability frontmatter — не опционален.** Пустой `requirements-completed` ломает 3-source cross-reference даже когда VERIFICATION зелёный; заполнять синхронно с планом.
3. **Doc/tooling-фазы тоже нуждаются в валидационном контракте** (presence/link/round-trip checks), доведённом до sign-off, а не оставленном PARTIAL.

### Cost Observations
- Model mix: преимущественно opus (quality profile).
- Sessions: 1 интенсивная.
- Notable: 14 планов закрыты за ~1 день; средняя длительность плана ~3 мин — выигрыш от узких, заземлённых на репозиторий задач.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0 | 1 | 4 | Базовый GSD-флоу: discuss→plan→execute→verify по фазам, волновая декомпозиция, no-phantom authoring |

### Cumulative Quality

| Milestone | Tests | Coverage | Zero-Dep Additions |
|-----------|-------|----------|-------------------|
| v1.0 | n/a (doc/tooling milestone) | Nyquist partial | 1 Node devDep (commitlint, exact-pinned) |

### Top Lessons (Verified Across Milestones)

1. No-phantom: документировать только реально проверенные команды/факты — первый же milestone дал контрпример (DOC-02), подтверждающий важность правила.
