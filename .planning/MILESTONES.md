# Milestones

## v1.0 Фундамент (Shipped: 2026-06-17)

**Delivered:** версионируемая база знаний `knowledge/` (правила для ИИ/команды) + проводка enforcement-тулинга — фундамент конвенций до бизнес-фич.

**Phases completed:** 4 phases, 14 plans, ~23 tasks
**Requirements:** 17/17 v1 satisfied (DOC-07 glossary deferred to domain-milestone)
**Artifacts:** 10 docs в `knowledge/` (~1101 строк) + `AGENTS.md` + тонкий `CLAUDE.md` (51 строка); `.golangci.yml` v2, `lefthook.yml`, commitlint, buf-скелет, `Makefile`
**Timeline:** intensive 1-day build (2026-06-17)

**Key accomplishments:**

- **Раскладка + точки входа (Phase 1):** authoring-стандарт MUST/SHOULD/WON'T (каждый запрет с парной «do»), индекс `knowledge/README.md`, `AGENTS.md` как канонический тонкий источник истины, `CLAUDE.md` урезан с ~205 до 51 строк — без phantom-ссылок и дублирования.
- **Foundation-доки (Phase 2):** `structure.md` (раскладка `go.work`: pkg/analytics/audit в workspace, inventory WIP вне), `build.md` (команды сборки/тестов), `git.md` (Conventional Commits, ветки dev/main), `boundaries.md` (do-not правила + карта владения фактами).
- **Конвенции + архитектура (Phase 3):** `style.md` (единый канон языка RU-комментарии/EN-имена, typed IDs, sentinel/`%w`), `testing.md` (Ginkgo v2 + Gomega + mockery), `architecture.md` (DDD+гексагон БЕЗ CQRS, MUST NOT возрождать диспетчер/`TxManager`), `patterns.md` (4 копируемых рецепта).
- **Enforcement-слой (Phase 4):** `.golangci.yml` v2 (errorlint + depguard ban на `pkg/mediatr`, gofumpt/gci), `lefthook.yml` (pre-commit lint/format, pre-push тесты, commit-msg), commitlint (config-conventional), buf v2-скелет; `Makefile` пиннит версии тулинга.
- **Статус enforcement (ENF-05):** каждое механизируемое правило в `knowledge/*.md` помечено честным статусом (hook / convention-only; ничего не CI-gated — CI ещё нет), легенда канонизирована в `authoring.md`.

**Known Gaps (accepted as tech debt):**

- **DOC-02** — `build.md:61-62` документирует `cd services/audit && go build ./...` как проходящую сборку, но команда падает (exit 1: `build output "cmd" already exists and is a directory`). Рабочие формы: `go build ./cmd`, `go vet ./...`, `go build -o /tmp/audit ./...`. Единственный блокер аудита; принят как tech debt.

**Known deferred items at close: 5** (see STATE.md Deferred Items) — DOC-02 build claim, Phase 02/04 verification gaps, 7 Phase 04 live-firing UAT scenarios (`make tools` + `lefthook install`), Nyquist sign-off pending Phases 1–4.

---
