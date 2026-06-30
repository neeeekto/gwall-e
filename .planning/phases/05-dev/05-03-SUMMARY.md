---
phase: 05-dev
plan: 03
subsystem: testing
tags: [mockery, mockery-v3, ginkgo, gomega, testify, codegen, smoke, unit-test]

# Dependency graph
requires:
  - phase: 05-dev
    provides: "Makefile MOCKERY_VERSION pin (v3.7.1) + generate-mocks таргет — Plan 02"
provides:
  - "Корневой .mockery.yaml (v3-конфиг: template testify, formatter goimports, all:false) → internal/example"
  - "Throwaway ExampleProvisioner интерфейс (smoke-порт; реальные доменные порты — Phase 6/7)"
  - "Сгенерированный mockery v3 мок ExampleProvisioner (testify expecter-API)"
  - "Unit-spec через NewMockExampleProvisioner(GinkgoT()) + Gomega — доказывает SC4-часть (mockery smoke), без build-tag"
affects: [mocks, unit-tests, domain-ports-phase6-7]

# Tech tracking
tech-stack:
  added: [github.com/stretchr/testify v1.9.0, github.com/stretchr/objx v0.5.2]
  patterns:
    - "mockery v3-конфиг: .SrcPackageName/.SrcPackagePath, template testify, моки рядом с кодом в {{.InterfaceDir}}/mocks (Pitfall 5: НЕ v2-ключи)"
    - "Unit-spec: NewMockX(GinkgoT()) (авто-Cleanup сверяет ожидания) + EXPECT().Return; testify/mock — обычный qualified-импорт, ginkgo/gomega — dot-import"

key-files:
  created:
    - .mockery.yaml
    - services/inventory/internal/example/provisioner.go
    - services/inventory/internal/example/mocks/ExampleProvisioner.go
    - services/inventory/internal/example/provisioner_test.go
  modified:
    - services/inventory/go.mod
    - services/inventory/go.sum
    - go.work.sum

key-decisions:
  - "Сигнатура порта Provision(ctx, id ExampleID, name string) error — чтобы типизированный ID (style.md) реально использовался, без dead-code хелперов"
  - "Тест в external-пакете example_test: мок живёт в под-пакете mocks (один файл/пакет, v3), импортируется и example, и mocks"

patterns-established:
  - "mockery v3 .mockery.yaml в корне репозитория, packages → конкретный inventory-пакет с interfaces:{Name:{}}"
  - "Smoke-проверка кодогена моков до появления реальных портов через throwaway example-интерфейс"

requirements-completed: [SVC-06]

# Metrics
duration: ~2 min
completed: 2026-06-30
---

# Phase 05 Plan 03: mockery v3 smoke (.mockery.yaml + ExampleProvisioner) Summary

**Заведён mockery v3 smoke (SVC-06): корневой `.mockery.yaml` (v3-синтаксис — template testify, formatter goimports, all:false, nацелен на `internal/example`) + throwaway `ExampleProvisioner` интерфейс; `make generate-mocks` сгенерировал testify-expecter мок, и unit-spec через `NewMockExampleProvisioner(GinkgoT())` + Gomega зелёный без Docker и без build-tag — доказывает SC4-часть (mockery подключён и проходит smoke-прогон).**

## Performance

- **Duration:** ~2 мин (чистая реализация; задержек/чекпойнтов не было)
- **Started:** 2026-06-30
- **Completed:** 2026-06-30
- **Tasks:** 2 (оба авто)
- **Files:** 4 создано, 3 изменено (go.mod/go.sum/go.work.sum — добавление testify)

## Accomplishments
- Создан корневой `.mockery.yaml` в v3-синтаксисе (Pitfall 5): `all: false`, `formatter: goimports`, `template: testify`, `dir: "{{.InterfaceDir}}/mocks"`, `filename: "{{.InterfaceName}}.go"`, `pkgname: "mocks"`; `packages:` → `internal/example` с `interfaces: { ExampleProvisioner: {} }`.
- Создан throwaway `ExampleProvisioner` интерфейс (`internal/example/provisioner.go`) с типизированным `ExampleID` и sentinel `ErrExampleProvisionFailed` (style.md), помечен русскими комментариями как пример-порт для smoke (реальные порты — Phase 6/7).
- `make generate-mocks` (пиннутый mockery v3.7.1) сгенерировал `internal/example/mocks/ExampleProvisioner.go` (testify expecter-API: `EXPECT()`, `NewMockExampleProvisioner`).
- Написан unit-spec (`provisioner_test.go`, пакет `example_test`, без build-tag): suite-каркас `RegisterFailHandler(Fail)` + `RunSpecs`; dot-import ginkgo/gomega; обычный импорт `testify/mock`; кейсы success (`Return(nil)`) и failure (`Return(sentinel)` + `MatchError`). `go test` зелёный, `go vet` чистый.

## Task Commits

1. **Task 1: .mockery.yaml (v3) + throwaway ExampleProvisioner** - `3497adb` (feat)
2. **Task 2: generate ExampleProvisioner mock + Gomega unit-spec** - `96e5e1a` (feat)

**Plan metadata:** см. финальный `docs(05-03)`-коммит.

## Files Created/Modified
- `.mockery.yaml` (создан) — корневой v3-конфиг mockery, нацелен на `internal/example`.
- `services/inventory/internal/example/provisioner.go` (создан) — throwaway `ExampleProvisioner` + `ExampleID` + sentinel-ошибка.
- `services/inventory/internal/example/mocks/ExampleProvisioner.go` (создан, кодоген) — mockery v3 testify-мок.
- `services/inventory/internal/example/provisioner_test.go` (создан) — Ginkgo/Gomega unit-spec против мока.
- `services/inventory/go.mod`, `services/inventory/go.sum`, `go.work.sum` (изменены) — добавлен `stretchr/testify` (+ `objx`), втянутый тест-кодом (T-05-07 accept).

## Decisions Made
- **Сигнатура `Provision(ctx, id ExampleID, name string) error`:** добавлен параметр типизированного ID, чтобы конвенция «типизированные ID» (style.md) реально использовалась в throwaway-интерфейсе, без искусственных `var _ = helper` заглушек.
- **Тест — external-пакет `example_test`:** mockery v3 кладёт мок в под-пакет `mocks` (один файл на пакет), поэтому spec импортирует и `example`, и `example/mocks` извне.

## Deviations from Plan

None — план выполнен как написано. Параметр `id ExampleID` в сигнатуре метода — это применение явно-разрешённой плановой опции («если фигурирует ID — типизированный»), а не отклонение; `testify` в go.mod ожидаем (T-05-07 accept, тянется тест-кодом).

## Threat Surface
- T-05-06 (Tampering, mockery codegen): mitigate — бинарь пиннут `MOCKERY_VERSION := v3.7.1` (Plan 02), кодоген закоммичен.
- T-05-07 (Tampering, testify): accept — зрелый `stretchr/testify`, попал в `go.sum` с хешами.

Новой security-релевантной поверхности (endpoints, auth, file/network access, schema на trust-boundary) план не вводит. Threat-флагов нет.

## Issues Encountered
- Локальный `mockery` в PATH — homebrew, но той же пиннутой версии v3.7.1, поэтому кодоген совпадает с ожидаемым (Makefile-таргет `generate-mocks` выполнился без ошибок).

## User Setup Required
None — `make generate-mocks` использует пиннутый mockery v3.7.1; тест гоняется обычным `go test` без Docker и без build-tag.

## Next Phase Readiness
- mockery v3 доказан рабочим: при появлении реальных доменных портов (Phase 6/7) достаточно добавить их пакеты/интерфейсы в `packages:` `.mockery.yaml` и перегенерировать.
- example-пакет — throwaway: может быть удалён/заменён реальными портами в Phase 6/7 без потери покрытия.

## Self-Check: PASSED

- .mockery.yaml — FOUND
- services/inventory/internal/example/provisioner.go — FOUND
- services/inventory/internal/example/mocks/ExampleProvisioner.go — FOUND
- services/inventory/internal/example/provisioner_test.go — FOUND
- Commit 3497adb — FOUND
- Commit 96e5e1a — FOUND

---
*Phase: 05-dev*
*Completed: 2026-06-30*
