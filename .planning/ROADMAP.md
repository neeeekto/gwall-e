# Roadmap: gwall-e

## Milestones

- ✅ **v1.0 Фундамент** — Phases 1-4 (shipped 2026-06-17) — база знаний `knowledge/` + enforcement-тулинг

Full v1.0 detail archived in [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md).

## Phases

<details>
<summary>✅ v1.0 Фундамент (Phases 1-4) — SHIPPED 2026-06-17</summary>

- [x] Phase 1: Раскладка базы знаний и точки входа (2/2 plans) — completed 2026-06-17
- [x] Phase 2: Стабильные доки-основы (3/3 plans) — completed 2026-06-17
- [x] Phase 3: Доки конвенций и архитектуры (5/5 plans) — completed 2026-06-17
- [x] Phase 4: Enforcement-слой (тулинг) (4/4 plans) — completed 2026-06-17

**Known gap (tech debt):** DOC-02 — `build.md` audit-рецепт `cd services/audit && go build ./...` падает (exit 1); рабочие формы `go build ./cmd` / `go vet ./...`. См. [MILESTONES.md](MILESTONES.md).

</details>

### 🔜 Next milestone (planned)

Запускается через `/gsd-new-milestone` (questioning → research → requirements → roadmap). Кандидаты: DOC-07 (glossary, domain-milestone) и/или первый бизнес-эпик платформы. Tech debt v1.0 (DOC-02 fix, Nyquist sign-off, live-firing UAT) — внести в скоуп следующего цикла.

## Progress

| Phase | Milestone | Plans Complete | Status   | Completed  |
| ----- | --------- | -------------- | -------- | ---------- |
| 1. Раскладка базы знаний и точки входа | v1.0 | 2/2 | Complete | 2026-06-17 |
| 2. Стабильные доки-основы | v1.0 | 3/3 | Complete | 2026-06-17 |
| 3. Доки конвенций и архитектуры | v1.0 | 5/5 | Complete | 2026-06-17 |
| 4. Enforcement-слой (тулинг) | v1.0 | 4/4 | Complete | 2026-06-17 |
