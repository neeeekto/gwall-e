# Phase 4: Enforcement-слой (тулинг) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-17
**Phase:** 4-Enforcement-слой (тулинг)
**Mode:** `--auto` (рекомендованный вариант выбран автоматически для каждой gray area;
обоснование — STACK.md, forward-метки D-11 Phase 3, no-phantom)
**Areas discussed:** Multi-module orchestration, commitlint runtime, depguard depth,
enforcement-status taxonomy, buf skeleton + tool pinning

---

## Multi-module orchestration (lint/format/test по go.work + inventory)

| Option | Description | Selected |
|--------|-------------|----------|
| Single root config, workspace-aware + отдельный GOWORK=off проход для inventory; pre-push тесты только in-workspace | Один `.golangci.yml`; pkg/analytics/audit через workspace, inventory отдельно; inventory исключён из pre-push (WIP) | ✓ |
| Per-module конфиги | По `.golangci.yml` на модуль | |
| Гонять всё включая inventory в pre-push | inventory тестируется наравне | |

**Choice:** Single root config + dual-set lint, pre-push только in-workspace (D-01/D-02/D-03).
**Notes:** STACK.md рекомендует единый корневой конфиг + lint per module с GOWORK=off; inventory —
WIP вне workspace, boundaries.md запрещает гонять/чинить WIP-леса → исключён из pre-push явно.

---

## commitlint runtime (Go-репо без package.json)

| Option | Description | Selected |
|--------|-------------|----------|
| Минимальный private package.json + config-conventional, exact-pin, npx --no-install | Следует research: единственная Node-зависимость, изолирована | ✓ |
| Go-native проверка (lefthook-regex / commitlint-rs / conform) | Без Node вообще | |
| Docker-контейнер для commitlint | Изоляция через образ | |

**Choice:** Node devDep, изолированный (D-04).
**Notes:** STACK.md прямо рекомендует commitlint+config-conventional как единственный Node-dep,
exact-pin, документировать. Напряжение с «Lefthook чтобы избегать node_modules» отмечено; Go-native —
в Deferred.

---

## depguard depth (нет кода слоёв)

| Option | Description | Selected |
|--------|-------------|----------|
| Forward-матрица по целевым путям (дремлет) + немедленные баны воскрешения снесённого | Layer-import правила активируются по мере кода; mediatr/TxManager баны кусают сегодня | ✓ |
| Полная матрица слоёв сейчас (аспирационно) | Все правила, хотя кода нет | |
| Закомментированный скелет | depguard выключен | |

**Choice:** Forward-совместимо + concrete WON'T-баны (D-05/D-06).
**Notes:** Слои не существуют как код → import-direction правила дремлют без phantom-претензии; баны
снесённых `pkg/mediatr`/`TxManager` реальны сразу. + errorlint (style.md `%w`).

---

## Enforcement-status taxonomy (CI-gated vs hook при отсутствии CI)

| Option | Description | Selected |
|--------|-------------|----------|
| Truthful relabel: planned:CI-gated → hook; CI-gated зарезервирован до CI | Честные статусы; легенда уточнена в authoring.md | ✓ |
| Оставить CI-gated (lint-конфиг = «CI source of truth») | Помечать как CI-gated, хотя CI нет | |
| Только convention-only | Не разделять hook/CI | |

**Choice:** Honest `hook` сейчас, `CI-gated` зарезервирован (D-07/D-08/D-09).
**Notes:** Полный CI вне скоупа (REQUIREMENTS); no-phantom запрещает заявлять несуществующий CI.
Тот же конфиг переиспользуется в CI позже без изменений.

---

## buf skeleton + tool pinning (нет .proto)

| Option | Description | Selected |
|--------|-------------|----------|
| buf.yaml/buf.gen.yaml v2 с пиннингом плагинов, помечены скелет, НЕ в падающих хуках; версии тулзов через Makefile + таблица | Честный скелет; codegen ручной до появления proto | ✓ |
| Полная codegen-обвязка сейчас | buf в хуках, плагины активны | |
| Пропустить buf | Отложить ENF-04 | |

**Choice:** Скелет + воспроизводимый пиннинг (D-10/D-11).
**Notes:** `.proto` ещё нет → no-phantom: не выдавать codegen за рабочий, buf вне падающих хуков.
Версии — из STACK.md (buf v1.71.x, golangci v2.12.x, lefthook v1.x), пиннить не `@latest`.

---

## Claude's Discretion

- Механика прогона golangci (workspace vs per-module цикл).
- Набор линтеров сверх gofumpt/gci/depguard/errorlint.
- Механизм пиннинга версий (Makefile vs `go.mod tool` Go 1.24 vs mise/.tool-versions).
- Имя/синтаксис commitlint-конфига и `commit-msg` вызова.
- Форма proto-root в buf.yaml и список codegen-плагинов.
- Формулировки обновлённой 3-статусной легенды в authoring.md.

## Deferred Ideas

- Полный CI-pipeline (перевод hook-правил в CI-gated) — будущий эпик.
- Реальные .proto + рабочая кодогенерация (наполнение buf-скелета).
- Go-native commitlint (уход от node_modules полностью).
- Mockery-обвязка целиком (`.mockery.yaml` + `go:generate`).
- Восстановление кода слоёв → активация дремлющих depguard-правил.
- ADR/anti-patterns/libraries/onboarding/maintenance — v2.
