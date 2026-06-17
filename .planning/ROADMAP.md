# Roadmap: gwall-e

## Milestones

- ✅ **v1.0 Фундамент** — Phases 1-4 (shipped 2026-06-17) — база знаний `knowledge/` + enforcement-тулинг
- 🗺️ **v2.0 L2-видение** — карта 12 доменов + направление (без кода); см. [L2-ARCHITECTURE.md](L2-ARCHITECTURE.md). Эпики режутся отдельными milestone'ами

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

### 🗺️ v2.0 — L2-видение платформы (epic-уровень)

> **Тип milestone:** L2-видение (без кода). Это НЕ нарезка фаз — это карта эпиков и направление.
> Каждый эпик = отдельный будущий milestone, который режется через `/gsd-new-milestone`.
> Полная архитектура — [L2-ARCHITECTURE.md](L2-ARCHITECTURE.md). Порядок индикативный,
> уточняем по ходу.

**Принцип:** сервис = домен (DDD). Идентичность Project/Host/VM владеет Inventory; домены
синхронизируются через Kafka (choreography для идентичности + orchestration-саги для длинных
процессов).

**Эпики (each → отдельный milestone):**

| # | Эпик (домен) | Цель (L2) | Зависит от |
|---|--------------|-----------|------------|
| E1 | **Inventory + Event-backbone** | Источник истины Project/Host/VM (идентичность/ЖЦ, железо, локация, owner-ref); solo/sync-инвентарь; outbox→relay→**Kafka** | — *(фундамент)* |
| E2 | **Access** | Вся авторизация: owner-роли, права/роли, временные гранты, IDM-sync, SSH-гранты | E1 |
| E3 | **Network** | Свитчи, VLAN, IPAM, сетевые шаблоны, смена VLAN | E1 |
| E4 | **Health / Monitoring** | Runtime, health-checks, config-compliance (+ ingest от Host Agent) | E1 |
| E5 | **Coordination** ⭐ | approve-before-start с Owner CMS, CMS-конфиг+вызов, локи, предохранители/лимиты | E1, E2 |
| E6 | **Integrations** | Адаптеры к внешним провайдерам операций (profile/reimage/network/reboot) | E1 |
| E7 | **Actions** | Единичные операции над хостом + каталог наливок | E5, E6, E3 |
| E8 | **Orchestrator** | Кросс-доменные lifecycle-саги (provision, decommission с вето) | E1, E5, E7 |
| E9 | **Scenarios** | Кампании плановых массовых работ (окна, drain, shutdown на учения, move owner) | E7, E5, E8 |
| E10 | **Remediation** | Авто-починка по правилам SRE (Automation Plot), self-healing | E4, E7 |
| E11 | **Audit** | Лог всех действий в системе (consumer событий) | E1 |
| E12 | **Analytics** | Аналитика парка | E1 |
| E13 | **Search** | Поверх OpenSearch, много индексов, event-fed; для людей и машин | E1 |
| E14 | **API Gateway / BFF** | Вход для фронтенда | по мере доменов |
| E15 | **Notifications** | Уведомления | по мере доменов |

**Отложенный кластер:** **Host Agent** (агент на хосте: сбор данных → Health, исполнение ← Actions,
раздача SSH ← Access; + серверная часть). Решение «домен vs часть Health» — при проектировании E2/E4.

**Индикативный порядок:** `E1` → (`E2` ∥ `E3` ∥ `E4`) → `E5` → (`E6` → `E7`) → `E8` → `E9` → `E10`;
платформенные (`E11`–`E15`) вплетаются по мере появления данных/фронта.

**Tech debt v1.0** (внести при нарезке E1): DOC-02 fix, Nyquist sign-off, live-firing UAT, DOC-07 glossary.

> Следующий шаг: `/gsd-new-milestone` для нарезки **E1 Inventory** в требования и фазы.

## Progress

| Phase | Milestone | Plans Complete | Status   | Completed  |
| ----- | --------- | -------------- | -------- | ---------- |
| 1. Раскладка базы знаний и точки входа | v1.0 | 2/2 | Complete | 2026-06-17 |
| 2. Стабильные доки-основы | v1.0 | 3/3 | Complete | 2026-06-17 |
| 3. Доки конвенций и архитектуры | v1.0 | 5/5 | Complete | 2026-06-17 |
| 4. Enforcement-слой (тулинг) | v1.0 | 4/4 | Complete | 2026-06-17 |
