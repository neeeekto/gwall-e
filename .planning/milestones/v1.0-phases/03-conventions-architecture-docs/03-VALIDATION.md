---
phase: 3
slug: conventions-architecture-docs
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-17
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> **Documentation phase:** «tests» = shell-проверки, что 4 дока удовлетворяют success criteria.
> Прозу нельзя валидировать unit-тестом — все проверки скриптуемы (grep / `test -f`).
> Источник: [03-RESEARCH.md](03-RESEARCH.md) § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Shell-проверки (`grep` / `test -f`), запускаются из корня репо. Go-эталон в `testing.md` опционально проверяется `go vet`/`go build` |
| **Config file** | none — Wave 0 не требуется (нет тест-фреймворка для прозы) |
| **Quick run command** | `bash` one-liners для затронутого дока (presence + link-check) |
| **Full suite command** | последовательная прогонка всех проверок из таблицы ниже |
| **Estimated runtime** | ~5 секунд (grep по `knowledge/`) |

---

## Sampling Rate

- **After every task commit:** `test -f` + presence-grep для дока этой задачи + link-check затронутых ссылок
- **After every plan wave:** полный набор presence / uniqueness / link-integrity по всем созданным докам
- **Before `/gsd-verify-work`:** все проверки зелёные + ручной просмотр диаграммы импортов (D-06) и меток «иллюстрация» (no-phantom)
- **Max feedback latency:** ~5 секунд

---

## Per-Requirement Verification Map

| Req ID | Behavior | Test Type | Automated Command | File Exists |
|--------|----------|-----------|-------------------|-------------|
| DOC-04 | `style.md` существует + правило языка комментариев | presence | `test -f knowledge/style.md && grep -qiE 'русск\|коммент' knowledge/style.md` | ❌ создаётся фазой |
| DOC-04 | typed ID / sentinel `%w` / DTO→домен присутствуют | presence | `grep -qE '%w' knowledge/style.md && grep -qiE 'типизирован\|typed' knowledge/style.md && grep -qiE 'DTO' knowledge/style.md` | ❌ |
| DOC-03 | `testing.md` фиксирует Ginkgo v2 + Gomega + mockery | presence | `test -f knowledge/testing.md && grep -qi 'ginkgo' knowledge/testing.md && grep -qi 'gomega' knowledge/testing.md && grep -qi 'mockery' knowledge/testing.md` | ❌ |
| DOC-03 | suite-бутстрап MUST задокументирован | presence | `grep -q 'RegisterFailHandler' knowledge/testing.md && grep -q 'RunSpecs' knowledge/testing.md` | ❌ |
| DOC-05 | `architecture.md` явно DDD+гексагон БЕЗ CQRS + MUST NOT | presence | `test -f knowledge/architecture.md && grep -qi 'гексагон' knowledge/architecture.md && grep -qi 'CQRS' knowledge/architecture.md && grep -qiE 'MUST NOT\|WON.T' knowledge/architecture.md` | ❌ |
| DOC-05 | Элементы: Execute / query-lite / UnitOfWork / outbox / PullEvents | presence | `for k in Execute UnitOfWork outbox PullEvents query; do grep -qi "$k" knowledge/architecture.md \|\| echo "MISSING $k"; done` | ❌ |
| DOC-05 | Диаграмма/таблица направления импортов (D-06) | presence + manual | `grep -qiE 'domain.*←\|импорт\|import' knowledge/architecture.md` (диаграмма — ручной просмотр) | ❌ |
| PAT-01 | `patterns.md` — 4 рецепта, ссылается на architecture.md | presence | `test -f knowledge/patterns.md && grep -q 'architecture.md' knowledge/patterns.md` + наличие use case/query/aggregate/repository | ❌ |

---

## Cross-Cutting Validations (no-phantom, dedup, links, authoring)

| Check | Type | Automated Command | Status |
|-------|------|-------------------|--------|
| no-phantom: сниппеты architecture/patterns помечены «иллюстрация» | manual + grep | `grep -qiE 'иллюстрац\|целевой вид' knowledge/patterns.md knowledge/architecture.md` | ❌ |
| Правило языка кода — ТОЛЬКО в `style.md` (нет дубля) | uniqueness | `! grep -rl 'комментарии.*на русском\|RU комментарии' knowledge/ \| grep -v 'style.md\|boundaries.md\|README.md'` | ❌ |
| Link integrity: нет битых markdown-ссылок | link-check | для каждой `[..](X.md)` в `knowledge/*.md` → `test -f knowledge/X` | частично |
| Ownership-map: testing/architecture/patterns зарегистрированы в `boundaries.md` | presence | `grep -q 'testing.md' knowledge/boundaries.md && grep -q 'architecture.md' knowledge/boundaries.md && grep -q 'patterns.md' knowledge/boundaries.md` | частично (style уже есть) |
| README индекс обновлён (ссылки, не «запланировано») | presence | `grep -E '\[testing.md\]\|\[architecture.md\]\|\[patterns.md\]\|\[style.md\]' knowledge/README.md` | частично |
| Размер ~150–200 строк на док | lint | `for f in style testing architecture patterns; do wc -l knowledge/$f.md; done` (флаг при >200) | ❌ |
| Forward enforcement-метки (D-11) на механизируемых правилах | presence | `grep -qiE 'convention-only\|planned.*Phase 4\|CI-gated\|hook' knowledge/style.md knowledge/testing.md knowledge/architecture.md` | ❌ |

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.* — тест-фреймворк не нужен (документационная фаза, проверки = shell one-liners). Link-checker опционален; fallback — `grep` + `test -f` (специфицированы выше).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Диаграмма направления импортов читаема и корректна (D-06) | DOC-05 | ASCII/таблица — смысловая проверка, не grep | Просмотреть блок диаграммы в `architecture.md`: `domain` не импортирует наружу; `usecases`/`api`/`repositories` → `domain` |
| Сниппеты идиоматичны и помечены «иллюстрация» (D-01) | PAT-01 / DOC-05 | grep ловит метку, но не качество кода | Просмотреть Go-сниппеты на нейтральном плейсхолдере; убедиться, что они не выдаются за существующий код |
| Каждое правило с тегом силы, без хеджирования | DOC-03/04/05 | смысловой просмотр нормативных строк | Нет «обычно/желательно/prefer» вместо MUST/SHOULD/WON'T |

---

## Validation Sign-Off

- [ ] Каждый из 4 докиов проходит presence-проверки своего Req ID
- [ ] Cross-cutting: no-phantom, uniqueness, link-integrity, ownership-map, README — зелёные
- [ ] Sampling continuity: после каждого дока прогоняются presence + link-check
- [ ] Размер каждого дока ≤ ~200 строк (иначе дробление per authoring.md)
- [ ] Ручные проверки (диаграмма, метки «иллюстрация») подтверждены
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
