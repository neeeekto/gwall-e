# Phase 6: Доменная модель Inventory - Context

**Gathered:** 2026-06-30
**Status:** Ready for planning

<domain>
## Phase Boundary

Спроектированный домен **Inventory** как Go-агрегаты с инвариантами идентичности/ЖЦ и
**семантическими доменными событиями** — фундамент всего event-backbone (без событий
нечего класть в outbox). Вертикаль доходит до **usecases** (interactor'ы), но **без**
реальной инфраструктуры.

**В скоупе (Success Criteria из ROADMAP + требования):**
- Агрегаты `Project`, `Host`, и локации `DC`/`Module`/`Rack`; `HostHardware` VO внутри Host
  (RAM/CPU/Drives/NIC/PSU/storage-controller/внутренний GPU/chassis/motherboard/IPMIMac)
- Идентичность: внутренний постоянный непереиспользуемый `ID` — единственный носитель
  идентичности (INV-01…03, INV-09); внешние идентификаторы — `string` (HW-06)
- ЖЦ `shadow → registered → decommissioned` + терминальный `deleted`; `decommissioned ≠ deleted`;
  re-add = новый `ID` без авто-мерджа; FQDN-конфликт среди active = доменный конфликт (INV-04…08)
- Локации DC→Module→Rack как первоклассные сущности с иерархией; Host→Rack+позиция;
  Rack несёт топологические атрибуты (LOC-01…04)
- Семантические доменные события на каждое изменение + envelope `eventId`/`version`/`actor`/`occurredAt`
  с первого дня (EVT-01, EVT-02)
- Usecases (interactor'ы) на write-операции, оркестрирующие через порты `UnitOfWork`/`Outbox`/repo (на фейках)
- Канон-слои `domain/usecases/query/repositories/api/cron` + `app` (SVC-01) — насколько они получают код в этой фазе
- `knowledge/glossary.md` (DOC-07): ubiquitous language, граница «факт существования vs динамическое состояние», `decommission ≠ delete`

**НЕ в скоупе (другие фазы):**
- Реальная persistence/Mongo, реализация `UnitOfWork` (Mongo-txn), transactional Outbox-коллекция,
  gRPC-адаптеры + identity-interceptor, query-сервисы на Mongo, partial unique FQDN-index — **Phase 7**
- protobuf-схемы событий (buf codegen), relay → Kafka, dual-topic, tombstone-эмиссия — **Phase 8**
- Топология `connections` хост↔модуль + read-model зависимостей, внешние HW-модули — **Phase 9**
- Test-consumer / верификация backbone — **Phase 10**
- Полный (нестабильный) advisory-matching движок — отдельный интеграционный сервис (SEED-001)
- VM/VMGroup, sync из внешней инвентори, Audit-домен, Access/права — будущие эпики

</domain>

<decisions>
## Implementation Decisions

### Срез Phase 6 ↔ Phase 7 (глубина вертикали)
- **D-01:** Вертикаль = **домен + usecases**. Строим агрегаты/VO/события/фабрики/инварианты
  **и** interactor-usecases на write-операции (RegisterHost, DecommissionHost, DeleteHost,
  ReassignHost, RelocateHost, ChangeHostHardware, CreateProject/…, CRUD локаций). Это делает
  SC1/SC3/SC5 реально верифицируемыми в Phase 6 («оператор может через usecase…», «FQDN-конфликт
  возвращает доменный конфликт», «каждое изменение рождает событие»).
- **D-02:** Usecases **уже оркестрируют через порты** `UnitOfWork.Do(ctx, fn)` + `Outbox.Append`
  ровно как в [knowledge/architecture.md](knowledge/architecture.md): внутри `fn` — `repo.Save(агрегат)` +
  `PullEvents()`→`outbox.Append`. Порты реализованы **фейками** (in-memory uow = просто зовёт `fn`;
  in-memory outbox = слайс). **Phase 7 = свап фейков на Mongo-impl, usecase НЕ меняется** — эталон
  виден целиком уже сейчас (предотвращает Pitfall 8 dual-write на уровне формы кода).
- **D-03:** Тест-дублёры — **только mockery-моки** для портов (repo/uow/outbox/query-порты): тест
  задаёт expectations на вызовы. Домен/агрегаты/инварианты тестируются **прямыми** unit-тестами без
  моков (чистые функции/фабрики). Mockery + Ginkgo v2 + Gomega уже провязаны в Phase 5.

### Агрегаты, идентичность, локации
- **D-04:** Локации DC/Module/Rack — **три независимых агрегата**, каждый свой корень с CRUD.
  Иерархия — через ссылки по **внутреннему ID**: Module несёт DCID, Rack несёт ModuleID. Host→RackID
  + позиция (юнит). Маленькие агрегаты (канон DDD), ложатся на «первоклассные CRUD» (LOC-01) и на
  Kafka key = ID каждой сущности (Phase 8). НЕ один агрегат-дерево Location.
- **D-05:** Внутренние ID — **типизированные ID-VO на агрегат** (`HostID`/`ProjectID`/`DCID`/`ModuleID`/`RackID`),
  обёртка над `uuid.UUID` (`github.com/google/uuid` уже в go.mod) с фабрикой/парсингом. Компилятор
  ловит перепутанные ссылки; идентичность выражена в типах. Внешние идентификаторы (`Owner`, `INV`,
  serial, MAC и т.п.) — сырой `string` (INV-09/HW-06).
- **D-06:** Ссылки **между** агрегатами — только по внутреннему ID (не вложенные объекты, не
  back-references). Reassign — операция на Host (меняет ProjectID), не на Project; Project не держит
  список своих хостов.
- **D-07:** `HostHardware` — **единый immutable VO** со всеми вложенными компонентами
  (Motherboard/RAM[]/CPU[]/Drives[]/NIC[]/PSU[]/storage-controller/внутренний GPU/chassis, IPMIMac).
  Изменение железа = собрать новый VO и заменить целиком → **одно** событие `HostHardwareChanged`.
  Чистая VO-семантика, простые инварианты. НЕ мутабельные под-компоненты с per-компонент событиями.
  NIC — структурированный компонент (модель/скорость/MAC'и), не плоский `MACs[]` (HW-03).

### ЖЦ, удаление, конфликты
- **D-08:** **`lifecycleState ∈ {shadow, registered, decommissioned}`**. `decommissioned` — смена
  состояния, хост **остаётся видим** (в live-store и будущем `*.state`-снапшоте), это НЕ tombstone.
- **D-09:** **`deleted` = hard-удаление**, НЕ член enum'а `lifecycleState`. Моделируется как агрегатный
  метод `Delete()` → эмитит событие `HostDeleted` (полный payload + actor), а usecase зовёт `repo.Delete`
  (физическое удаление). История/факт живут **только в append-only `*.events`** (Phase 8); в `*.state`
  уходит Kafka-**tombstone** (Phase 8); FQDN освобождается. Никакой строки `state=deleted` не остаётся
  (это и есть «hard, а не soft»; аудит не теряется — он в immutable-логе). Ключевой канон: Pitfall 2.
- **D-10:** Граф переходов — **гибкий вход + терминальный decommission**. Host создаётся либо в `shadow`
  (заготовка/обнаружен), либо сразу `registered` (ручная полная регистрация). Переходы: `shadow→registered`;
  `shadow→decommissioned` и `registered→decommissioned`. `decommissioned` **терминально** (нет
  воскрешения; повторный ввод = новый `ID`, INV-08). `Delete()` — из любого состояния.
- **D-11:** **FQDN-конфликт среди active** — доменный инвариант, проверяется **в usecase** через
  query-порт (напр. `ActiveHostByFQDN`/uniqueness-checker) → возвращает **типизированный доменный
  конфликт** (не сырой DB-error). Partial unique index — defense-in-depth в Phase 7. Канон: Pitfall 7.
- **D-12:** **Advisory-matching (INV-08)** — объявляется как **порт-хук** (интерфейс) под будущую
  интеграцию + заглушка (no-op/пустые кандидаты) в Phase 6. Полный (нестабильный) составной матч
  (INV+FQDN+MAC+локация+окно) — отдельный интеграционный сервис, вне scope (SEED-001). Авто-restore/merge
  запрещён by design.

### Доменные события + envelope
- **D-13:** **Одно семантическое событие на бизнес-операцию** (минимально-достаточный payload, не
  `HostUpdated`-дамп — Pitfall 5). Регистрация → один `HostRegistered` (identity + начальные
  hardware/location в payload). Далее отдельные: `HostHardwareChanged`, `HostReassigned` (смена Project),
  `HostRelocated` (смена Rack/позиции), `HostDecommissioned`, `HostDeleted`. Аналогично для Project
  (`ProjectCreated`/`ProjectDeleted`/…) и локаций (`DCCreated`/`DCUpdated`/`ModuleCreated`/`RackCreated`/…).
  Имена событий = ubiquitous language, фиксируются в DOC-07 glossary.
- **D-14:** **Факты в домене + envelope на границе.** Агрегат эмитит «голые» семантические факты
  (`eventType`/`entityId`/`version` + payload) и копит их; `PullEvents()` сливает их. Envelope-мета
  (`eventId`, `occurredAt`, `actor{id, source}`) **навешивается между `PullEvents()` и `outbox.Append`**
  на границе usecase. Домен-ядро остаётся чистым: **нет** Clock/IDGen/actor внутри агрегата; `actor`
  (транспортная identity) никогда не входит в агрегат. Ложится на «actor из gRPC-interceptor» (Phase 7).
- **D-15:** Envelope несёт `eventId`, `entityId`, `eventType`, `version` (поле агрегата), `occurredAt`,
  `actor{id, source: human|api|integration|system}` — **с первого дня** (forward-compat для Audit,
  Pitfall 6 / SEED-002). На v3.0 source почти всегда `human`/`api`, но поле присутствует и заполняется.
  `version` — поле агрегата, инкрементится при каждом изменении (optimistic-concurrency enforcement — Phase 7).

### Claude's Discretion
- **DOC-07 glossary:** точный состав/формулировки терминов и границы «факт существования vs динамическое
  состояние» — на усмотрение планнера/executor'а в рамках D-08…D-15 и [knowledge/authoring.md](knowledge/authoring.md).
  Обязательно фиксирует: Project/Host/Owner/Module/Connection, идентичность, `decommission ≠ delete`,
  имена семантических событий (D-13).
- **Какие канон-слои получают код в Phase 6 (SVC-01):** ожидаемо `domain` (агрегаты/VO/события/порты),
  `usecases` (interactor'ы), и query-порт(ы) для uniqueness-проверки; `repositories`/`api`/`cron`/`app`
  предположительно остаются скелетом/фейками до Phase 7. Финальную раскладку определяет планнер в рамках D-01/D-02.
- **Project-агрегат — операции/события** (rename, смена Owner, delete) и **инвариант уникальности
  позиции host↔rack** (два хоста в одном юните) — на усмотрение планнера, в рамках D-13 (одно событие
  на операцию) и D-11 (инвариант в usecase, если нужен).
- Имена пакетов, точные сигнатуры портов, форма типизированных ID-VO, структура фабрик — планнеру/executor'у.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Ресёрч (доменные решения, грабли — ядро этой фазы)
- `.planning/research/PITFALLS.md` — **критично**. Pitfall 2 (`decommissioned`=lifecycle vs `deleted`=tombstone,
  decommission≠delete), Pitfall 5 (семантические события vs `HostUpdated`-дамп), Pitfall 6 (`actor/initiator`
  в envelope с 1-го дня), Pitfall 7 (re-add = доменный конфликт + advisory, не DB-error). Чеклист
  «Looks Done But Isn't» — приёмочные точки Phase 6.
- `.planning/research/SUMMARY.md` — executive summary v3.0, dual-topic, rationale идентичности.
- `.planning/research/ARCHITECTURE.md` — целевая архитектура backbone (контекст outbox/relay/dual-topic).
- `.planning/research/FEATURES.md` — нарезка фич Inventory.
- `.planning/research/STACK.md` — версии (`mongo-driver/v2`, `google/uuid`); MongoDB UoW-специфика (контекст для портов Phase 7).

### Каноны репозитория (правила, которым следует домен)
- `knowledge/architecture.md` — **канон слоёв** `domain/usecases/query/repositories/api/cron`+`app`, направление
  импортов внутрь, use case=interactor `Execute`, порт `UnitOfWork`, фабрики + `PullEvents()` + outbox/relay,
  валидация транспорта на edge, запрет CQRS-шины. Ядро для D-01/D-02/D-14.
- `knowledge/patterns.md` — пошаговые рецепты «как добавить use case / query / агрегат / репозиторий».
- `knowledge/style.md` — стиль кода; плейсхолдер `Order`; русские доменные комментарии, английские строки ошибок/логов.
- `knowledge/structure.md` — членство `go.work`, inventory как эталонный компилируемый модуль.
- `knowledge/boundaries.md` — границы «что не трогать», карта владения фактами (один факт = один канон).
- `knowledge/authoring.md` — стандарт MUST/SHOULD/WON'T + pointer-over-copy (для написания DOC-07 glossary).
- `knowledge/testing.md` — Ginkgo v2 + Gomega + mockery конвенции (для D-03; англ. комментарии в тестах).
- `knowledge/glossary.md` — **создаётся в этой фазе (DOC-07)**; ubiquitous language домена Inventory.

### Требования и роадмап
- `.planning/ROADMAP.md` § Phase 6 — Goal + 6 Success Criteria + список требований; § Phase 7/8/9/10 — границы (что НЕ здесь).
- `.planning/REQUIREMENTS.md` — INV-01…09, HW-01…06, LOC-01…04, EVT-01, EVT-02, DOC-07, SVC-01 (тексты требований).
- `.planning/PROJECT.md` — Core Value, milestone v3.0, Out of Scope (VM/sync/Audit/Access).
- `.planning/L2-ARCHITECTURE.md` — L2-карта доменов, инвариант семантических событий, кросс-доменный принцип сущностей.

### Seeds (forward-compat контекст)
- `.planning/seeds/SEED-001-inv-matching-instability.md` — почему advisory-matching — хук, а не движок (D-12); «нет внешнего ключа стабильного И уникального».
- `.planning/seeds/SEED-002-audit-logging.md` — почему `actor/eventId` в envelope с 1-го дня (D-15).
- `.planning/seeds/SEED-003-authorization-on-all-actions.md` — почему identity пробрасывается (stub под будущий Access; контекст для actor).

### Прошлый контекст (carry-forward)
- `.planning/phases/05-dev/05-CONTEXT.md` — **D-10** (список агрегатов data-driven: `project`/`location`/`module` добавляются дёшево),
  **D-12** (Kafka key = внутренний `ID`, НЕ FQDN/INV/MAC), connection-helper как сигнатура подключения (не реализация UoW).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `services/inventory/go.mod`: `github.com/google/uuid v1.6.0` уже есть — основа типизированных ID-VO (D-05);
  `mongo-driver/v2 v2.7.0` (после Phase 5) — для портов Phase 7, в Phase 6 не используется напрямую.
- `services/inventory/cmd/main.go`: точка входа (composition root поднимается в `app` — Phase 7+).
- Phase 5 провязал `.mockery.yaml` + `make generate-mocks` (smoke на example-интерфейсе) — реальные доменные
  порты (repo/uow/outbox/query) появляются в Phase 6 и становятся первыми «настоящими» целями mockery (D-03).
- `pkg/` — общая библиотека; **generic**-элементы (если возникнут) обязаны жить в `pkg/`, не дублироваться (канон проекта).

### Established Patterns
- `services/inventory/internal/` — целевые канон-слои (`domain/usecases/query/repositories/api/cron/app`) ожидаются пустыми
  (чистый лист после v1.0); домен Phase 6 — первый реальный код в `domain` + `usecases`.
- inventory — полноправный член `go.work`, **всегда компилируется** (D-01/D-03 из Phase 5); pre-push гоняет unit-тесты inventory.
- Тесты: Ginkgo v2 + Gomega; интеграционные — за build-tag `integration` (Phase 6 — чистые unit, без контейнеров).
- Комментарии в коде — на русском (доменная терминология), строки ошибок/логов — на английском ([knowledge/style.md]).

### Integration Points
- `domain` (агрегаты/события/порты) ← `usecases` (interactor'ы) — направление импортов внутрь (architecture.md).
- Порты `UnitOfWork`/`Outbox`/repo/uniqueness объявляются в `domain`, реализуются фейками в Phase 6 → Mongo в Phase 7 (D-02).
- Envelope-граница (`PullEvents()` → enrich → `outbox.Append`) — точка, куда в Phase 7 ляжет реальная outbox-коллекция, в Phase 8 — relay→Kafka.
- Имена семантических событий (D-13) ↔ protobuf-схемы Phase 8 ↔ DOC-07 glossary — единый ubiquitous language.

</code_context>

<specifics>
## Specific Ideas

- «hard, а не soft»: пользователь явно подтвердил, что `deleted` — физическое удаление записи, история
  только в append-only `*.events` (D-09). Это не теряет аудит, т.к. он в immutable-логе, не в живой коллекции.
- «Эталон виден целиком уже сейчас»: usecases в Phase 6 имеют **ту же форму** (UoW+Outbox через порты),
  что и в Phase 7 — цель в том, чтобы Phase 7 был чистым свапом реализаций, а не переписыванием (D-02).
- Типизированные ID-VO выбраны ради компайл-тайм защиты от перепутанных ссылок (Host.RackID ≠ ProjectID), D-05.

</specifics>

<deferred>
## Deferred Ideas

- **Полный advisory-matching движок** (составной нестабильный матч INV+FQDN+MAC+локация+окно) — отдельный
  интеграционный сервис, не Inventory (SEED-001). В Phase 6 — только порт-хук + заглушка (D-12).
- **Реальная persistence/UoW(Mongo-txn)/Outbox-коллекция/gRPC/query-на-Mongo/partial FQDN-index** — Phase 7.
- **protobuf-схемы событий + relay→Kafka + dual-topic + tombstone-эмиссия** — Phase 8 (имена/форма событий
  фиксируются сейчас как контракт, но кодоген и публикация — там).
- **Топология `connections` + read-model + внешние HW-модули** — Phase 9.
- **VM/VMGroup, sync из внешней инвентори, Audit-домен, Access/права** — будущие эпики (PROJECT.md Out of Scope).

</deferred>

---

*Phase: 6-inventory*
*Context gathered: 2026-06-30*
