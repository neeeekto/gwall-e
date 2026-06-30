---
phase: 05-dev
plan: 05
subsystem: dev-enforcement-canon
tags: [lefthook, knowledge-base, doc-02, go-vet, go-work]
requires: ["05-01", "05-04"]
provides:
  - "pre-push гоняет inventory unit-тесты (go test ./..., без integration build tag, D-15)"
  - "lefthook без GOWORK=off и без исключения inventory (D-01/D-02)"
  - "build.md audit-рецепт go vet ./... (exit 0, DOC-02/SC5)"
  - "каноны build/structure/boundaries описывают inventory как члена go.work (D-04)"
affects: [lefthook.yml, knowledge/, .planning/ROADMAP.md]
tech-stack:
  added: []
  patterns:
    - "audit-валидация без бинаря через go vet ./... (обходит package-main коллизию cmd, Pitfall 2)"
    - "pre-push unit-only; integration за build-тегом, чтобы хук не тянул Docker (D-15)"
key-files:
  created: []
  modified:
    - lefthook.yml
    - knowledge/build.md
    - knowledge/structure.md
    - knowledge/boundaries.md
    - knowledge/testing.md
    - knowledge/README.md
    - .planning/ROADMAP.md
decisions:
  - "go vet ./... как каноническая audit-валидация (go build ./... падает build output cmd already exists)"
  - "stale GOWORK=off убран не только в плановых файлах, но и в testing.md/README.md (canon-drift, T-05-12)"
metrics:
  duration: ~8min
  completed: 2026-06-30
---

# Phase 5 Plan 05: lefthook de-exclusion + каноны под go.work + DOC-02 (go vet) Summary

Снят отменённый `GOWORK=off`/исключение inventory: pre-push теперь гоняет inventory unit-тесты (без integration build tag, D-15), audit-рецепт build.md переведён на эмпирически зелёный `go vet ./...` (DOC-02/SC5), а каноны build/structure/boundaries + ROADMAP приведены к реальности «inventory — полноправный член go.work» (D-01/D-04).

## What Was Built

- **lefthook.yml** — pre-commit `lint-inventory` слит в общий `lint-workspace`-цикл (без `GOWORK=off`); pre-push получил `test-inventory: cd services/inventory && go test ./...` (unit only); удалён комментарий-блок «inventory INTENTIONALLY NOT tested».
- **knowledge/build.md** — audit-рецепт `cd services/audit && go build ./...` → `go vet ./...` (exit 0, проверено на месте); раздел inventory переписан под workspace-build (всегда компилируется, D-03); pre-push описан как включающий inventory unit.
- **knowledge/structure.md** — workspace = четыре члена (+`services/inventory`); раздел inventory из «вне workspace, WIP» в «член workspace».
- **knowledge/boundaries.md** — inventory убран из примера WIP-лесов; карта владения фактами без «исключения inventory из pre-push» и без `GOWORK=off`.
- **.planning/ROADMAP.md** — DOC-02 known-gap помечен resolved (go vet exit 0); SC1 уже не содержал «с GOWORK=off».

## Verification

- `cd services/audit && go vet ./...; echo $?` == `0` (SC5/DOC-02) — verified.
- `grep 'cd services/inventory && go test' lefthook.yml` — присутствует (D-02).
- `grep -r 'GOWORK=off' knowledge/ lefthook.yml .planning/ROADMAP.md` — пусто (D-01/D-04).
- ROADMAP Phase 5 SC1 без «с GOWORK=off» — verified.
- Pre-push smoke (без Docker): `cd services/inventory && go test ./...` exit 0 (D-15) — verified.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug / T-05-12] Stale GOWORK=off в testing.md и README.md**
- **Found during:** Task 3 (sweep `grep -r GOWORK=off`)
- **Issue:** `knowledge/testing.md` (строка 147) и `knowledge/README.md` (строка 33) ссылались на `GOWORK=off` для inventory — факт, удалённый из build.md этим планом. Каноны указывали бы на несуществующую команду (canon-drift, threat T-05-12).
- **Fix:** Переформулированы ссылки на «workspace-build/`go vet`»; README таблица модулей — inventory как член workspace.
- **Files modified:** knowledge/testing.md, knowledge/README.md
- **Commit:** 59358a6

Файлы не были в `files_modified` плана, но правка их обязательна для устранения drift, прямо вызванного D-01/D-04-изменениями (Rule 1: ссылка на удалённый факт = баг канона).

## Notes / Follow-ups

- **Локальная активация хуков:** применение нового lefthook.yml требует разового `lefthook install` в клоне (state-инвентаризация RESEARCH) — это не коммитится. Live-firing inventory pre-push проверяется после bootstrap (см. отложенный 04-UAT).
- `git build ./cmd` остаётся падающим (Pitfall 2, package main в cmd/) — канон сознательно использует `go vet ./...`, бинарь не собирается в audit-рецепте.

## Self-Check: PASSED

- Files: lefthook.yml, knowledge/build.md, knowledge/structure.md, knowledge/boundaries.md, knowledge/testing.md, knowledge/README.md, .planning/ROADMAP.md — все FOUND.
- Commits: e93b924, edd1e22, 59358a6 — все FOUND в git log.
