# Phase 4: Enforcement-слой (тулинг) - Context

**Gathered:** 2026-06-17
**Status:** Ready for planning

> **Режим:** контекст собран в `--auto` — для каждой gray area выбран рекомендованный
> вариант, обоснованный `.planning/research/STACK.md` (тулинг уже зафиксирован ресёрчем),
> forward-метками D-11 Phase 3 и принципом no-phantom (`knowledge/boundaries.md`).
> Решения ниже — дефолты; человек может скорректировать перед/во время планирования.

<domain>
## Phase Boundary

Фаза подкрепляет **механизируемые** правила базы знаний реальным тулингом и проставляет
каждому такому правилу **фактический enforcement-статус** — база перестаёт быть только
декларативной. В скоупе ровно 5 артефактов (ENF-01..05):

1. **`.golangci.yml`** — golangci-lint **v2** (схема v2, `linters.default`), **gofumpt** как
   форматтер, **gci** для порядка импортов; единый корневой конфиг, консистентный с
   мульти-модульным `go.work`.
2. **`lefthook.yml`** — git-хуки: `pre-commit` (lint + format), `pre-push` (тесты),
   `commit-msg` (commitlint).
3. **Конфиг commitlint** (Conventional Commits), подключённый к `commit-msg` хуку.
4. **Скелет `buf.yaml` + `buf.gen.yaml`** для proto (lint / breaking / codegen) — **скелет**,
   т.к. `.proto` в репо ещё нет.
5. **ENF-05:** каждое механизируемое правило в `knowledge/*.md` помечено **фактическим**
   статусом enforcement (переключение forward-меток D-11, а не ретрофит с нуля).

**Вне скоупа** (REQUIREMENTS Out of Scope — НЕ расширять):
- **Полный CI-pipeline** (выбор раннера, матрицы, workflow-файлы) — Phase 4 делает только
  **локальную** проводку (lefthook + конфиги линтеров); CI — будущий эпик.
- **Реальные `.proto` и рабочая кодогенерация** — buf только скелет; codegen не «работает»
  (нет схем), хуки не должны падать на пустом вводе.
- **Восстановление кода `inventory`/слоёв** (`domain/usecases/...`) — это эпик реализации;
  здесь слои ещё не существуют как код.
- **Mockery-обвязка как «уже настроенная фича»** сверх минимального задела (testing.md уже
  фиксирует mockery как канон; `go:generate`/`.mockery.yaml` — только если попадает в ENF, без
  phantom-претензий).

Требования: **ENF-01** (`.golangci.yml`), **ENF-02** (`lefthook.yml`), **ENF-03** (commitlint),
**ENF-04** (buf-скелет), **ENF-05** (enforcement-статусы в доках).

</domain>

<decisions>
## Implementation Decisions

### Мульти-модульная оркестрация lint/format/test (ENF-01, ENF-02)
- **D-01:** **Единый корневой `.golangci.yml`** (схема v2: `linters.default: standard` + opt-in,
  блок `formatters` для gofumpt+gci) — один источник правды формата/линта. **Запуск** учитывает
  раскладку workspace: для модулей **в** `go.work` (`./pkg`, `./services/analytics`,
  `./services/audit`) — обычный прогон; для **`inventory`** (намеренно вне workspace) — отдельный
  проход с `GOWORK=off` из каталога модуля. Конкретная механика прогона (один вызов с
  workspace-режимом vs цикл по модулям) — на усмотрение планировщика/ресёрча, но **оба**
  множества модулей должны линтоваться.
- **D-02:** **`pre-commit`** = lint + format. Форматирование (gofumpt + gci) подключено **через
  golangci-lint v2 `formatters`** (формат == линт == будущий CI идентичны), не отдельным
  параллельным вызовом gofumpt. Держать хук **быстрым** (по возможности на staged-файлах);
  CI — будущий источник истины (вне скоупа).
- **D-03:** **`pre-push`** = тесты **только для in-workspace модулей** (`pkg` через `cd pkg && go test`,
  `analytics`, `audit`). **`inventory` ИСКЛЮЧЁН** из pre-push: он WIP, вне workspace, `internal/`
  снесён — гонять/«чинить» WIP-леса запрещено (`boundaries.md`). Это документируется явно, чтобы
  исключение не выглядело как пропуск. Ginkgo — там, где есть suite (реальный эталон — `pkg/http`).

### commitlint в Go-репо без package.json (ENF-03)
- **D-04:** Следуем ресёрчу (STACK.md): **минимальный корневой `package.json`** (`private: true`,
  только `devDependencies`, **exact-pin** `@commitlint/cli` + `@commitlint/config-conventional`) +
  `commitlint.config.js` (extends `config-conventional`). `commit-msg` хук вызывает
  `npx --no-install commitlint --edit {1}`. Установка (`npm install`) документируется рядом с
  `lefthook install` как разовый bootstrap-шаг. **Контекст-напряжение:** Lefthook выбран чтобы
  избегать node_modules — но commitlint остаётся **единственной** Node-зависимостью, изолированной
  в dev-тулинг (de-facto стандарт Conventional Commits). Go-native альтернатива — в Deferred.

### depguard при отсутствии кода слоёв (ENF-01, готовит ENF-05)
- **D-05:** depguard конфигурируется **forward-совместимо**: правила направления импортов из
  `architecture.md` (домен не импортирует наружу; `usecases → domain`; `api`/`repositories →
  usecases`/`domain`) пишутся против **целевых путей слоёв** (`.../internal/domain`, `.../usecases`
  и т.д.). Слоёв-кода сейчас нет → эти правила **дремлют** (ничего не матчат) и активируются по мере
  появления кода — без phantom-претензии, что они уже что-то проверяют.
- **D-06:** Параллельно — **немедленно кусающие** баны «воскрешения снесённого» (concrete WON'T из
  `architecture.md`/`PROJECT.md`): запрет импорта удалённых `pkg/mediatr` (CQRS-шина),
  `TxManager`/`tx`-диспетчера. Эти guard'ы реальны **сегодня** (сработают на любой реинтродукции).
  Плюс `errorlint` для sentinel-vs-wrapped (`%w`/`errors.Is`) — flips style.md «planned: CI-gated
  Phase 4 (errorlint)». Точный набор линтеров (depguard, errorlint, + опц. под typed-IDs) — на
  усмотрение планировщика в рамках forward-меток Phase 3.

### Таксономия enforcement-статуса — truthful reconciliation (ENF-05)
- **D-07:** Phase 4 проводит **локальные lefthook-хуки**, а **полный CI вне скоупа**. Чтобы не
  утверждать несуществующий CI (no-phantom), статусы определяются **честно**:
  - `hook` — правило проверяется git-хуком локально (lint/format/commit-msg/test) через
    закоммиченный конфиг — **доступно сегодня**.
  - `convention-only` — review-enforced, без автоматизации.
  - `CI-gated` — **зарезервировано** до появления CI-пайплайна (будущий эпик); тот же
    `.golangci.yml`/`buf` переиспользуются в CI без изменений.
- **D-08:** Соответственно **forward-метки D-11 переключаются**: `planned: CI-gated Phase 4
  (depguard/errorlint/...)` → **`hook (lint: depguard/...)`** (НЕ `CI-gated` — CI пока нет);
  `planned: hook (gofumpt)` → `hook (format)`; `convention-only (review-enforced)` остаётся как есть.
  `architecture.md` D-11-метки про CI-gated depguard перелейблятся в `hook`, с **одной строкой**
  заметки, что конфиг станет CI-gated при появлении CI.
- **D-09:** Легенда статусов уточняется **в одном каноне** — `knowledge/authoring.md` §«Статус
  enforcement» (определить 3 статуса точно + правило «не помечать CI-gated без CI»). Topic-доки
  ссылаются на легенду, не переопределяют. Снять формулировку «Phase 1 фиксирует только стандарт»
  на «статусы проставлены (Phase 4)».

### buf-скелет + пиннинг инструментов (ENF-04, ENF-01..03)
- **D-10:** `buf.yaml` (v2: `lint` + `breaking` конфиг, модуль указывает на будущий proto-root) и
  `buf.gen.yaml` (v2: плагины с **пиннингом версий** — `protoc-gen-go`, `protoc-gen-go-grpc`; опц.
  protovalidate per STACK). Помечены как **скелет**: `.proto` ещё нет, поэтому buf **НЕ** включается
  в падающие lefthook-хуки (codegen/lint по proto — ручной/opt-in до появления схем). no-phantom: не
  выдавать кодоген за «работающий».
- **D-11:** Версии всех инструментов (golangci-lint v2.12.x, gofumpt, gci, lefthook v1.x, buf
  v1.71.x, commitlint exact) **пиннятся воспроизводимо**: install-таргеты в `Makefile` (`make tools`)
  со ссылкой на таблицу версий + документирование в `knowledge/build.md` (одноразовый bootstrap:
  `make tools` + `lefthook install` + `npm install`). Точный механизм (Makefile vs `go.mod` `tool`
  directive Go 1.24 vs `.tool-versions`/mise) — на усмотрение планировщика/ресёрча; commitlint
  пиннится через `package.json` exact. Версии берутся из STACK.md (HIGH-confidence, июнь 2026).

### Claude's Discretion
- Точная механика прогона golangci по workspace vs per-module цикл (D-01) — планировщик/ресёрч.
- Конкретный набор линтеров сверх gofumpt/gci/depguard/errorlint (D-06) — в рамках forward-меток.
- Механизм пиннинга версий тулзов (Makefile vs `go.mod tool` vs mise) (D-11).
- Имя файла commitlint-конфига (`commitlint.config.js` vs `.commitlintrc.*`) и точный синтаксис
  `commit-msg` вызова (D-04).
- Точная форма proto-root в `buf.yaml` и список codegen-плагинов (D-10).
- Формулировки обновлённой легенды статусов в `authoring.md` (D-09) в рамках 3-статусной модели.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Источник истины по тулингу и версиям (ГЛАВНОЕ для этой фазы)
- `.planning/research/STACK.md` — пиннутые версии и обоснования: golangci-lint **v2.12.x** (схема v2,
  `linters.default`, gofumpt+gci в `formatters`, lint per module с `GOWORK=off`), **Lefthook v1.x**
  (over Husky; pre-commit/pre-push/commit-msg), **commitlint + config-conventional** (единственная
  Node-зависимость, exact-pin, изолировать), **buf v1.71.x** (`buf.gen.yaml` пиннит плагины, lint+breaking),
  protobuf APIv2, protovalidate; install-команды; Version Compatibility (Go 1.24); анти-паттерны
  (golangci v1-схема, Husky в Go-репо, raw protoc).

### Источник истины по скоупу и решениям
- `.planning/REQUIREMENTS.md` — **ENF-01..ENF-05** (формулировки success-критериев); **Out of Scope**
  (полный CI-pipeline; восстановление inventory; reference-service walkthrough — пока inventory не
  собирается). v2 (ADR/anti-patterns/libraries) — не сейчас.
- `.planning/ROADMAP.md` — Phase 4 goal + 5 success criteria (golangci/lefthook/commitlint/buf/статусы).
- `.planning/PROJECT.md` — Key Decisions (БЕЗ CQRS-шины: `pkg/mediatr` удалён; `TxManager`/`tx.go`
  удалён → база для depguard-банов D-06), Constraints (Go 1.24.6 workspace, `inventory` вне `go.work`,
  Git remote `neeeekto/gwall-e`, ветки `dev`/`main`).

### Правила, которым проставляются статусы (вход для ENF-05) — содержат forward-метки D-11
- `knowledge/authoring.md` §«Статус enforcement» — **канон легенды** статусов (`CI-gated`/`hook`/
  `convention-only`); ЗДЕСЬ уточняется таксономия (D-07/D-09).
- `knowledge/style.md` — метки: typed IDs `planned: CI-gated Phase 4 (linter)`, sentinel/`%w`
  `planned: CI-gated Phase 4 (errorlint)`, DTO→домен `planned: CI-gated Phase 4 (depguard)`, gofumpt
  `planned: hook`, плюс `convention-only (review-enforced)` правила языка.
- `knowledge/architecture.md` — метки: направление импортов/слои `planned: CI-gated Phase 4 (depguard)`,
  запрет импорта снесённых пакетов `planned: CI-gated Phase 4 (depguard)`, прочее `convention-only`.
- `knowledge/testing.md` — mockery `planned: Phase 4 (go:generate)`; suite/структура `convention-only`.
- `knowledge/patterns.md` — `convention-only (review-enforced)` метки (рецепты ссылаются на architecture/style).
- `knowledge/boundaries.md` — **no-phantom** (не документировать несуществующее как работающее; WIP не
  эталон) + **карта владения фактами** — buf/codegen-скелет и исключение inventory должны быть честными.
- `knowledge/build.md` — команды сборки/тестов (`cd pkg && go test`, `GOWORK=off` для inventory);
  сюда добавляется bootstrap-инструкция тулинга (D-11).
- `knowledge/structure.md` — раскладка `go.work` (модули в/вне workspace) — основа D-01/D-03.

### Прецедент (как фиксировались решения этого milestone)
- `.planning/phases/03-conventions-architecture-docs/03-CONTEXT.md` — **D-11** (forward-метки писались
  «вперёд», Phase 4 только переключает статус), D-10 (mockery как канон), плейсхолдер `Order` (D-05),
  pointer-over-copy/карта владения.
- `.planning/phases/02-foundation-docs/02-CONTEXT.md` — декскоуп glossary, модульный structure.md,
  карта владения фактами.

### Эталон кода (для D-03 pre-push тестов)
- `pkg/http/*_test.go` — **реальные** Ginkgo v2 + Gomega тесты (suite-бутстрап); единственный
  компилируемый тест-эталон. `pkg/go.mod` фиксирует ginkgo v2.23.4 / gomega v1.38.0.
- `go.work` — `use (./pkg ./services/analytics ./services/audit)`; `inventory` намеренно отсутствует.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **Все 10 `knowledge/*.md` существуют** и уже несут forward enforcement-метки (D-11 Phase 3) —
  ENF-05 их **переключает**, а не пишет с нуля. `authoring.md` уже содержит §«Статус enforcement».
- `go.work` (pkg/analytics/audit) + договорённость `GOWORK=off` для inventory — готовая основа
  оркестрации линта/тестов (D-01/D-03), задокументирована в `structure.md`/`build.md`.
- `pkg/http/*_test.go` + `pkg/go.mod` — живой Ginkgo+Gomega эталон для pre-push (D-03).

### Established Patterns
- **No-phantom / pointer-over-copy** (`boundaries.md`, `authoring.md`): один факт = один канон;
  не документировать несуществующее как рабочее. Прямо диктует D-07/D-08/D-10 (честные статусы,
  buf-скелет без претензии на codegen).
- **Forward-then-flip** (D-11): метки ставились заранее → Phase 4 минимально переключает статус.
- **Progressive disclosure**: легенда статусов — в `authoring.md`, topic-доки ссылаются (D-09).

### Integration Points
- `authoring.md` §«Статус enforcement» ↔ все topic-доки (легенда → ссылки) — D-09.
- `.golangci.yml` (depguard/errorlint) ↔ метки `architecture.md`/`style.md` — переключаются вместе (D-08).
- `lefthook.yml` ↔ `.golangci.yml` (pre-commit), `pkg`/`analytics`/`audit` тесты (pre-push),
  commitlint (commit-msg) — D-02/D-03/D-04.
- `build.md` ↔ bootstrap тулинга (`make tools`/`lefthook install`/`npm install`) — D-11.
- `boundaries.md` карта владения — зарегистрировать enforcement-конфиги/исключение inventory честно.

### Risks / Constraints for planner
- ⚠️ **Нет кода слоёв:** `domain/usecases/query/repositories/api` не существуют → depguard-правила
  направления импортов **дремлют** (D-05); не выдавать за активную проверку. inventory `internal/` пуст.
- ⚠️ **CI вне скоупа:** не помечать ничего `CI-gated` (нет CI) — только `hook`/`convention-only` (D-07).
  Не создавать `.github/workflows`/CI-конфиги.
- ⚠️ **Нет `.proto`:** buf — скелет; не включать buf в падающие хуки, не утверждать рабочий codegen (D-10).
- ⚠️ **Node-напряжение:** commitlint тянет минимальный Node-стек в Go-репо — держать изолированным
  (private package.json, только devDeps), документировать (D-04); не расширять Node-тулинг.
- ⚠️ **inventory в pre-push:** не гонять/«чинить» WIP-леса (boundaries.md) — исключение явно
  задокументировать как осознанное, не как пропуск (D-03).
- ⚠️ **Версии:** брать из STACK.md (golangci v2.12.x, buf v1.71.x, lefthook v1.x) — пиннить, не `@latest`.

</code_context>

<specifics>
## Specific Ideas

- Тулинг **уже выбран ресёрчем** (STACK.md, HIGH-confidence июнь 2026): golangci-lint **v2**, gofumpt+gci
  как `formatters`, **Lefthook** (не Husky), **commitlint+config-conventional** (единственный Node-dep),
  **buf v1.71** — это не переоткрывается, фаза только **проводит и пиннит** их.
- Принцип сессии (повтор Phase 1–3): фундамент правил для ИИ/команды, **не описание системы** и **не
  выдавать незавершённое за работающее** — диктует честные статусы и скелет-buf.
- ENF-05 — это **переключение forward-меток D-11**, а не новая разметка; ключевой нюанс — таксономия
  `hook` vs `CI-gated` при отсутствии CI (D-07/D-08).

</specifics>

<deferred>
## Deferred Ideas

- **Полный CI-pipeline** (раннер/матрицы/workflow-файлы, перевод `hook`-правил в `CI-gated`) — будущий
  эпик (Out of Scope REQUIREMENTS); те же `.golangci.yml`/`buf` переиспользуются без изменений.
- **Реальные `.proto` + рабочая кодогенерация** (наполнение buf-скелета, включение buf в хуки) —
  разблокируется при появлении схем/собираемого сервиса.
- **Go-native commitlint** (lefthook-regex / commitlint-rs / conform) как замена Node-стека — вариант
  на будущее, если репо захочет уйти от node_modules полностью (сейчас следуем research — Node devDep).
- **Mockery-обвязка целиком** (`.mockery.yaml` + `go:generate` + установка) сверх минимального ENF-задела —
  дозревает с восстановлением кода слоёв/inventory.
- **Восстановление кода слоёв** (`domain/usecases/...`), после чего дремлющие depguard-правила (D-05)
  становятся активными — отдельный эпик реализации.
- **ADR-доки, `anti-patterns.md`, `libraries.md`, onboarding, maintenance-протокол** — v2 (REQUIREMENTS).

</deferred>

---

*Phase: 4-Enforcement-слой (тулинг)*
*Context gathered: 2026-06-17*
