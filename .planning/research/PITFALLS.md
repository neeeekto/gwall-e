# Pitfalls Research

**Domain:** Hardware inventory (DC asset identity/reconciliation) + Kafka event-backbone via transactional outbox/relay (producer-only)
**Researched:** 2026-06-26
**Confidence:** HIGH (доменные решения уже зафиксированы в SEED-001/002 и L2-ARCHITECTURE; технические грабли Kafka/outbox верифицированы по Confluent docs + независимым практикам, см. Sources)

> Scope-напоминание: v3.0 — **полный продюсер, консьюмеров нет** (PROJECT.md). Грабли «idempotency/inbox консьюмера» формально вне v3.0, но включены как **forward-compat обязательства схемы/контракта**, которые дёшево заложить сейчас и дорого — потом. Помечены `[forward-compat]`.

---

## Critical Pitfalls

### Pitfall 1: Compacted-топик объявлен носителем «истории» — но compaction по определению уничтожает историю

**What goes wrong:**
PROJECT.md фиксирует одновременно две вещи: (1) «История — на event-backbone» (Key Decision про удаление, SEED-001 п.5) и (2) «семантические события + **compacted**-снапшот по `entityID`». Если эти две сущности живут в **одном** compacted-топике (или если «история» = чтение compacted-топика с offset=earliest), то аудит-след физически невосстановим: log-compaction оставляет только **последнее** значение на ключ, а промежуточные `HostUpdated`/`FQDNChanged`/`OwnershipChanged` стираются при первом проходе клинера. После `delete.retention.ms` tombstone удаляется тоже — и **факт, что ключ вообще существовал, исчезает**. Это прямо ломает core value «безопасное и согласованное управление» и SEED-002 (Audit).

**Why it happens:**
Соблазн «один топик решает обе задачи» (фид + снапшот идентичности для онбординга). Kafka-доки действительно рекламируют «feeding both use cases off the same backing topic», но это про **снапшот текущего состояния**, не про **полную историю**. Разработчики путают «replay compacted-топика» (восстановит только current state) с «replay event-лога» (восстановит любой момент времени).

**How to avoid:**
Разделить два потока физически:
- **Append-only (non-compacted) топик семантических событий** = immutable источник истории/аудита. Ретеншн — длинный/`-1` (compliance-окно), партиция by `entityID`. Это то, что слушает будущий Audit/Analytics и с чего делается backfill.
- **Compacted топик «снапшот идентичности» by `entityID`** = derived current-state, для дешёвого онбординга нового домена. Только последнее состояние на сущность.
Relay пишет в оба (dual-topic из одной outbox-записи) ИЛИ снапшот-топик — проекция off event-лога. «История на event-backbone» из PROJECT.md = **append-only топик**, не compacted.

**Warning signs:**
- В дизайне один топик с `cleanup.policy=compact` назван и «фидом», и «историей».
- Тест «восстановить, что происходило с хостом X во времени» проходит на свежих данных, но падает после прогона компакции.
- В схеме топиков нет ни одного `cleanup.policy=delete` (или `compact,delete` с длинным ретеншном) для событийного потока.

**Phase to address:**
Фаза проектирования event-backbone / топология топиков (до первого `relay`-кода). Зафиксировать как инвариант в L2-ARCHITECTURE «Анти-паттерны».

---

### Pitfall 2: Tombstone стирает аудит-след decommission/delete; гонка `delete.retention.ms`

**What goes wrong:**
Модель decommission/delete на событиях легко скатывается в «послать tombstone (key + null) в compacted-топик = удалить». Но tombstone в Kafka — это команда **забыть ключ**: после двух проходов клинера и истечения `delete.retention.ms` исчезает и tombstone, и все прежние значения. То есть «удаление через tombstone» = **уничтожение аудита удаления** — ровно того, что SEED-002 требует сохранить («кто, что, когда удалил»). Плюс гонка: медленный/оффлайн-консьюмер (будущий Audit), читающий с offset=0 дольше `delete.retention.ms`, **никогда не увидит факт удаления**.

**Why it happens:**
Семантическое смешение бизнес-понятий с механикой брокера: `decommissioned`/`deleted` (доменные ЖЦ-состояния) ≠ Kafka-tombstone (физическое забывание ключа). PROJECT.md прямо разводит `decommissioned ≠ deleted`, но при compacted-снапшоте оба соблазнительно реализовать одним tombstone.

**How to avoid:**
- `decommissioned` и `deleted` — это **доменные события** (`HostDecommissioned`, `HostDeleted`) с полным payload + `actor/initiator`, едут в **append-only** топик. Это и есть аудит-след — он immutable.
- Tombstone посылать **только в compacted снапшот-топик** и **только** на терминальный `deleted` (убрать запись из current-state-проекции). `decommissioned` — это смена `lifecycleState`, НЕ tombstone (списанный хост остаётся видим).
- `delete.retention.ms` снапшот-топика ≥ worst-case lag будущего консьюмера (правило: ≥24ч, лучше больше под backfill оффлайн-домена). Никогда `delete.retention.ms=0`.
- Никогда не использовать compacted снапшот как единственный носитель факта удаления.

**Warning signs:**
- В коде relay/usecase `decommission` приводит к публикации записи с null-value.
- `delete.retention.ms` не задан явно или мал.
- Нет отдельного доменного события `HostDeleted` с actor — только tombstone.

**Phase to address:**
Фаза «Идентичность/удаление» (доменные события ЖЦ) + фаза «event-backbone» (конфиг топиков). Связать с SEED-002 forward-compat.

---

### Pitfall 3: Дизайн ключа compacted-топика на нестабильном/переиспользуемом идентификаторе

**What goes wrong:**
Compaction работает **по ключу**. Если ключом снапшот-топика взять что-то «внешнее и понятное» (FQDN, INV, IPMI MAC) — модель немедленно протекает по SEED-001: FQDN рециклится (новый хост с тем же FQDN схлопнет/перезатрёт снапшот старого), INV/MAC меняются при замене материнки (один логический хост раздваивается на два ключа). Получаем либо ложный мердж (потеря данных под compaction), либо дубликаты-«призраки».

**Why it happens:**
Желание сделать ключ человекочитаемым/матчабельным. Прямое нарушение SEED-001 («нет внешнего ключа стабильного И уникального во времени») на уровне инфраструктуры Kafka, а не только домена.

**How to avoid:**
Ключ Kafka = **наш внутренний постоянный `ID`** (тот же, что single identity owner в Inventory). Он стабилен, уникален, не переиспользуется — единственный валидный ключ и для партиционирования (порядок by entity), и для compaction. FQDN/INV/MAC — это **атрибуты в payload**, по которым НИКОГДА не строится ключ/идентичность. Валидировать в relay: null-ключ на compacted-топике запрещён (иначе unbounded growth).

**Warning signs:**
- В коде продюсера `kafkaKey = host.FQDN` / `host.INV`.
- Обсуждение «давайте матчить по FQDN на стороне топика».
- Дубликаты снапшота для одного физического хоста после замены материнки.

**Phase to address:**
Фаза «event-backbone» (выбор ключа) — но решение вытекает из фазы «Идентичность». Зафиксировать: Kafka key == internal ID, hard rule.

---

### Pitfall 4: Relay нарушает per-entity порядок (параллелизм / сортировка по времени / retries)

**What goes wrong:**
Партиция by `entityID` даёт порядок **только если** relay реально публикует события одной сущности в порядке коммита. Три классические течи ломают это даже при правильном ключе:
1. **Параллельный relay** без стратегии — `HostUpdated` обгоняет `HostRegistered`.
2. **`ORDER BY created_at`** в polling-relay — при параллельных транзакциях/clock-skew порядок timestamp ≠ порядок коммита.
3. **Producer retries при `max.in.flight.requests > 1`** без идемпотентного продюсера — сетевой ретрай переставляет два сообщения.
Для inventory это означает: консьюмер (Analytics/Search/Audit) увидит «обновление хоста раньше его создания» или «decommission раньше последнего апдейта» — рассогласование current-state и аудита.

**Why it happens:**
Outbox-канон v1.0 гарантирует at-least-once и отсутствие dual-write, но **порядок** — отдельная забота relay, её часто упускают, фокусируясь на доставке.

**How to avoid:**
- Outbox-таблица: монотонный **`sequence` (BIGSERIAL)**, relay читает строго `ORDER BY sequence` — НЕ по `created_at`.
- Relay публикует **последовательно в рамках партиции** (можно параллелить **между** entityID, но не внутри); проще — последовательный продьюсер на старте v3.0 (150k хостов, producer-only — нагрузка скромная).
- Kafka producer: `enable.idempotence=true` (и тогда `max.in.flight` безопасно до 5) ИЛИ `max.in.flight.requests.per.connection=1`. `acks=all`.
- Cross-aggregate out-of-order — **это норма by design**, не баг; не пытаться чинить глобальный порядок (один партишн убьёт масштаб).

**Warning signs:**
- В relay `SELECT ... ORDER BY created_at`.
- `max.in.flight` не сконфигурирован + `enable.idempotence` не выставлен.
- Тест «создание→апдейт→decommission одного хоста» иногда приходит в консьюмер в перепутанном порядке.

**Phase to address:**
Фаза «event-backbone / relay». Покрыть интеграционным тестом порядка на одном `entityID`.

---

### Pitfall 5: «Жирные `EntityUpdated`-дампы» vs тощие notification — неверный объём event-carried state

**What goes wrong:**
Два симметричных анти-паттерна (L2 «Анти-паттерны»):
- **Жирный `HostUpdated` со всем снапшотом хоста** на каждое чихание → консьюмеры не понимают, *что* изменилось; высокий трафик; связность по полной схеме; compaction снапшот-топика раздувается.
- **Тощий `HostChanged{id}`** без полей → каждый консьюмер вынужден синхронно дёргать Inventory (ACL query) → distributed monolith, гонки, нагрузка на Inventory.

**Why it happens:**
Нет осознанного контракта «семантическое событие». Соблазн «отправим всё на всякий случай» или «отправим только ID, остальное дочитают».

**How to avoid:**
**Семантические доменные события** (L2 инвариант 5): `HostRegistered`, `HostDecommissioned`, `FQDNReassigned`, `OwnershipChanged`, `HostMovedToRack` — с **минимально-достаточным** payload именно этого факта (что изменилось + ключевые поля для проекции) + `actor/initiator`. Отдельно — **compacted снапшот-топик** для «полного текущего состояния» (онбординг/backfill). Не путать: семантическое событие ≠ снапшот. `HostChanged`-обобщения избегать (инвариант: осмысленные факты, не `HostChanged`).

**Warning signs:**
- В коде один тип события `HostUpdated` на все мутации.
- Консьюмеры (будущие) вынуждены звать Inventory сразу после события.
- payload события = полная Mongo-документ-проекция Host.

**Phase to address:**
Фаза «доменные события» (дизайн схемы событий). Glossary DOC-07 фиксирует имена событий как ubiquitous language.

---

### Pitfall 6: `actor/initiator` не заложен в схему событий с самого начала (forward-compat для Audit)

**What goes wrong:** `[forward-compat]`
Консьюмеров нет в v3.0, поэтому соблазн «actor добавим, когда будет Audit». Но если события Inventory с самого начала идут без `actor/initiator`, то когда Audit (E11) станет consumer'ом, **исторические события в append-only/compacted топиках не атрибутируемы** — придётся переэмитить/обогащать события задним числом, ломая immutability аудит-лога. SEED-002 прямо помечает это «дешёвой страховкой сейчас vs дорого потом».

**Why it happens:**
YAGNI применён к тому, что **нельзя дозаложить ретроактивно** в immutable-лог.

**How to avoid:**
В базовый envelope события (наряду с `eventId`, `entityId`, `eventType`, `version`, `occurredAt`) **сразу** класть `actor`/`initiator`: `{ id, source: human|api|integration|system }`. На v3.0 источник почти всегда `human`/`api` (solo-режим), но поле присутствует и заполняется. Описать envelope в DOC-07 glossary как контракт.

**Warning signs:**
- Envelope события не содержит actor.
- В дискуссии звучит «actor добавим в Audit-эпике».

**Phase to address:**
Фаза «доменные события / envelope» (самая ранняя event-фаза v3.0). Hard requirement, не «по ходу».

---

### Pitfall 7: Re-add без авто-мерджа реализован как «упадём на FQDN-уникальности» вместо явного конфликта

**What goes wrong:**
SEED-001: re-add = новый ID, без авто-мерджа, FQDN уникален только среди `active`. Наивная реализация: при добавлении хоста с уже занятым (среди active) FQDN — просто кинуть `duplicate key`-ошибку из Mongo. Это (а) протекает реализацией БД наружу, (б) не даёт человеку контекст «возможно это вернувшийся ранее списанный хост — кандидаты вот», (в) при возврате decommissioned-хоста с тем же FQDN молча либо упадёт, либо (хуже) кто-то добавит авто-restore-with-merge (запрещён by design, Out of Scope).

**Why it happens:**
Уникальность FQDN среди active выглядит как «просто unique index». Но семантика — **доменный конфликт для human-in-the-loop**, а не инфраструктурная ошибка.

**How to avoid:**
- FQDN-уникальность среди `active` — доменный инвариант, проверяется в usecase (а не только partial unique index в Mongo как defense-in-depth).
- При коллизии — возвращать **доменный конфликт** с advisory-кандидатами (составной советочный матч из SEED-001: INV+FQDN+MAC+локация+окно), а не сырой DB-error.
- Никакого авто-restore/merge: re-add всегда = новый ID (Out of Scope запрещает merge by design).
- Partial unique index `{fqdn: 1}` только для `lifecycleState: active` — освобождение FQDN при decommission/delete должно реально снимать индексную блокировку.

**Warning signs:**
- В usecase нет явной проверки FQDN-конфликта, полагается на Mongo `E11000`.
- Уникальный индекс на FQDN без `partialFilterExpression`.
- Появился код «найти прошлый хост и восстановить».

**Phase to address:**
Фаза «Идентичность/удаление» (инварианты записи) — провязать с Out of Scope «авто-мердж запрещён».

---

### Pitfall 8: Outbox-запись и доменная мутация не в одной UoW-транзакции (dual-write возвращается)

**What goes wrong:**
Канон v1.0: события публикуются через transactional outbox **внутри UoW-транзакции** (`PullEvents` → запись в outbox в той же Mongo-txn). Грабля при первой реальной реализации: usecase делает `repo.Save(host)` в txn, а событие пишет в outbox **после** коммита (или через отдельный клиент/сессию) — это и есть dual-write, который outbox должен был убить. При краше между коммитом домена и записью outbox событие теряется → консьюмеры/аудит навсегда рассинхронены.

**Why it happens:**
Mongo-транзакции требуют, чтобы **все** записи шли через одну `session`/`ctx`. Легко «потерять» сессию для outbox-коллекции. Особенно — на 150k нагрузке соблазн «вынести outbox за транзакцию ради скорости».

**How to avoid:**
- UoW оборачивает запись; outbox-репозиторий берёт транзакцию из `ctx` (как все репозитории — Key Decision v1.0).
- `PullEvents()` из агрегата → вставка в outbox-коллекцию **той же** Mongo-сессией, в **той же** txn, что и `Save`.
- Интеграционный тест: инъецировать панику между save и outbox — проверить, что либо оба, либо ничего (atomicity).
- Relay — отдельный процесс, читает outbox после коммита (at-least-once). Не «inline-публикация» из usecase.

**Warning signs:**
- Outbox-вставка вне `UnitOfWork.Execute`/без txn-`ctx`.
- Kafka-producer вызывается прямо из usecase.
- Нет теста на atomicity save+outbox.

**Phase to address:**
Фаза «эталон UoW + outbox» (фундамент). Это reference-реализация для всех будущих сервисов — критично сделать образцово.

---

### Pitfall 9: Топология `connections` (хост↔модуль) и read-model рассинхронены / без целостности ссылок

**What goes wrong:**
`connections` — Mongo cross-refs (хост↔дисковая полка/GPU) + read-model «что зависит от X». Внешние HW-модули — **без owner**, их ID — `string`. Грабли: (а) висячие ссылки (модуль decommissioned, connection осталась → read-model «зависит от X» врёт); (б) read-model «что зависит от X» построен как live-join на запросе (N+1 на 150k) или, наоборот, материализован событиями, но без идемпотентного апдейта → дрейф; (в) каскад отказа (полка тянет несколько хостов) не отражён, т.к. направление связи перепутано.

**Why it happens:**
Inventory владеет **физической топологией**, но не операциями (каскадные действия — другие домены, Out of Scope). Легко начать тащить операционную логику или строить наивный live-граф.

**How to avoid:**
- `connections` — явные доменные сущности/VO со ссылками на **внутренние ID** (для хостов) и `string` внешние ID (для модулей); направление связи зафиксировать в glossary (что от чего зависит).
- Read-model «что зависит от X» — отдельный query-сервис, читает Mongo напрямую в DTO (CQRS-lite канон); при росте — материализовать, но обновлять идемпотентно по событиям топологии.
- При decommission/delete хоста или модуля — событие топологии (`ConnectionRemoved`), read-model реагирует; не оставлять висячих connection.
- НЕ реализовывать каскадные **действия** (это Actions/Orchestrator, Out of Scope) — только read-model зависимостей.

**Warning signs:**
- Read-model «зависит от X» делает live multi-join на каждый запрос.
- Connection ссылается на FQDN/INV вместо внутреннего ID.
- Decommission хоста не трогает его connections.

**Phase to address:**
Фаза «Топология connections + read-model». Provязать scope-границу с Out of Scope (каскадные действия — не здесь).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Один compacted-топик и для фида, и для «истории» | Меньше топиков/кода | Аудит-след невосстановим; SEED-002 ломается; переезд = переэмит всего | **Never** — append-only лог истории заложить сразу |
| Tombstone как реализация decommission | «Удаление работает» одним механизмом | Стирание аудита удаления; гонка delete.retention.ms | **Never** для decommission; tombstone только для терминального delete снапшота |
| Relay `ORDER BY created_at` | Просто писать | Перестановка событий внутри сущности при clock-skew/parallel txn | Never — использовать монотонный `sequence` |
| Events без `actor/initiator` до Audit-эпика | Меньше полей сейчас | Исторические события не атрибутируемы; переэмит immutable-лога | Never (SEED-002) — поле дешёвое |
| Sequential relay (без параллелизма между entity) | Простой порядок, мало кода | Пропускная способность ниже | OK на v3.0 (producer-only, нагрузка скромная); ревизия при росте |
| FQDN unique index без partialFilter | Один индекс | Не освобождает FQDN при decommission; ломает re-add | Never — partial по `active` |
| Read-model зависимостей как live-join | Нет материализации | N+1 на 150k хостов | OK для MVP малого парка; материализовать до prod-масштаба |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Outbox → Mongo txn | Outbox-вставка вне UoW-сессии (dual-write) | Та же `session`/`ctx`, что и доменный `Save`; тест atomicity |
| Relay → Kafka | `max.in.flight>1` без идемпотентного продюсера | `enable.idempotence=true`, `acks=all`; читать outbox `ORDER BY sequence` |
| Kafka compacted-топик | Активный сегмент не компактится на низком трафике (стейл-снапшот) | Тюнить `segment.ms`/`segment.bytes`; задать `max.compaction.lag.ms` |
| Kafka compacted-топик | Null-ключ / нестабильный ключ (FQDN) | Ключ = внутренний `ID`; запрет null-ключа в relay |
| Kafka tombstone | `delete.retention.ms` мал/0 → консьюмер пропускает удаление | `delete.retention.ms` ≥ worst-case lag (≥24ч); никогда 0 |
| Mongo cross-refs (connections) | Висячие ссылки после decommission | Событие `ConnectionRemoved`; read-model реагирует идемпотентно |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Read-model «зависит от X» live-join | Рост latency запроса топологии | Материализовать проекцию по событиям | ~десятки тыс. хостов / глубокая топология |
| Sequential relay как единственная стратегия | Лаг outbox растёт | Параллелить **между** entityID (порядок внутри сохранён ключом) | Когда producer-throughput станет узким местом |
| Compacted-топик не чистится (низкий трафик per key) | Снапшот-топик растёт, стейл-значения | `max.compaction.lag.ms` + сегмент-тюнинг | Редкие апдейты на ключ + большой park |
| Cleaner не успевает (>5M уник. ключей/проход) | Высокий dirty-ratio не падает | Доп. cleaner-потоки; следить `uncleanable.partitions.count` | ~150k+ хостов в одном compacted-топике |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Событие без `actor/initiator` | Невозможно установить «кто сделал» (нарушение core value безопасности) | Envelope с actor сразу (SEED-002) |
| Owner = «понятный» внешний ID в открытом виде без проверки источника | Подмена владельца проекта | Owner = непрозрачный внешний ID группы; резолв — интеграция (Key Decision) |
| Авто-restore-with-merge вернувшегося хоста | Ложная атрибуция данных/прав чужому хосту (рециклинг FQDN) | Запрещён by design (Out of Scope); только advisory-match для человека |
| Compacted-снапшот как источник аудита удаления | Аудит удаления исчезает после delete.retention.ms | Аудит — в append-only immutable логе |

## "Looks Done But Isn't" Checklist

- [ ] **Event-backbone:** часто отсутствует **append-only** топик истории — есть только compacted-снапшот → проверить, что история восстановима после прогона компакции.
- [ ] **Decommission/delete:** часто `decommissioned` и `deleted` слиты в один tombstone — проверить, что `decommissioned` = смена `lifecycleState` (хост видим), tombstone только на терминальный `deleted`.
- [ ] **Outbox:** часто событие пишется после коммита домена — проверить atomicity тестом с инъекцией паники между `Save` и outbox.
- [ ] **Relay ordering:** часто `ORDER BY created_at` — проверить монотонный `sequence` + тест порядка на одном entityID.
- [ ] **Envelope:** часто нет `actor/initiator` — проверить, что поле в каждом событии (да, на v3.0 заполняется human/api).
- [ ] **Kafka key:** часто ключ = FQDN/INV — проверить, что ключ == внутренний `ID` во всех продюсерах.
- [ ] **FQDN-uniqueness:** часто полный unique index — проверить `partialFilterExpression: {lifecycleState: active}` и освобождение FQDN при decommission.
- [ ] **Re-add:** часто молчаливый restore/merge или сырой DB-error — проверить доменный конфликт с advisory-кандидатами.
- [ ] **Connections:** часто остаются висячие ссылки — проверить очистку при decommission и отсутствие каскадных **действий** (вне scope).

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| История потеряна (только compacted-топик) | **HIGH** | Завести append-only топик, переэмитить из Mongo current-state (история до этого момента невосстановима); зафиксировать дату-границу |
| Tombstone стёр аудит удаления | **HIGH** | Невосстановимо из Kafka; восстановление из Mongo/бэкапов если велись; впредь — append-only лог |
| События без actor в проде | **MEDIUM/HIGH** | Добавить поле в envelope (опционально на старом потоке); старые события атрибутировать нельзя |
| Relay переставляет порядок | **MEDIUM** | Перейти на `ORDER BY sequence` + идемпотентный продюсер; консьюмеры с version/seq переживут (last-writer-by-version) |
| Dual-write потерял событие | **MEDIUM** | Перевести outbox в UoW-txn; reconciliation Mongo↔consumer для расхождений |
| Ключ Kafka = FQDN (дубликаты) | **MEDIUM** | Сменить ключ на ID → новый топик + replay из источника истины (нельзя re-key in place) |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. Compacted ≠ история | Event-backbone (топология топиков) | Тест: восстановить историю хоста после прогона компакции |
| 2. Tombstone стирает аудит | Идентичность/удаление + Event-backbone | `decommissioned`=lifecycle, tombstone только на delete; `delete.retention.ms`≥24ч |
| 3. Ключ на нестабильном ID | Event-backbone (выбор ключа) | grep продюсеров: ключ == internal ID |
| 4. Relay ломает порядок | Event-backbone / relay | Интеграционный тест порядка на одном entityID + idempotent producer config |
| 5. Жирные/тощие события | Доменные события (схема) | Review: семантические события, нет `HostUpdated`-дампа; нет sync-ACL у консьюмера |
| 6. Нет actor/initiator | Доменные события / envelope (ранняя) | Каждое событие несёт actor; зафиксировано в DOC-07 |
| 7. Re-add без явного конфликта | Идентичность/удаление | Тест: re-add active-FQDN → доменный конфликт + кандидаты, не DB-error |
| 8. Dual-write вне UoW | Эталон UoW + outbox (фундамент) | Тест atomicity save+outbox (инъекция паники) |
| 9. Connections дрейф/висячие | Топология connections + read-model | Тест: decommission снимает connections; нет каскадных действий |

## Sources

- [Kafka Log Compaction — Confluent Documentation](https://docs.confluent.io/kafka/design/log_compaction.html) — HIGH (официальная): активный сегмент не компактится, tombstone+`delete.retention.ms`, ключ-based.
- [Understanding Kafka Compaction — Ted Naleid](https://www.naleid.com/2023/07/30/understanding-kafka-compaction.html) — MEDIUM: tombstone требует двух проходов; гонка delete.retention.
- [Kafka quirks: tombstones that refuse to disappear — Javier Holguera](https://javierholguera.com/2020/02/17/kafka-quirks-tombstones-that-refuse-to-disappear/) — MEDIUM: `delete.retention.ms=0` опасен при restore.
- [How to Handle Kafka Topic Compaction — OneUptime](https://oneuptime.com/blog/post/2026-01-24-handle-kafka-topic-compaction/view) — MEDIUM: null/нестабильный ключ ломает компакцию; dual-write append-only + compacted.
- [Transactional Outbox: Database-Kafka Consistency — Conduktor](https://www.conduktor.io/blog/transactional-outbox-pattern-database-kafka) — MEDIUM: aggregate_id как ключ, порядок per-partition.
- [Revisiting the Outbox Pattern — Decodable](https://www.decodable.co/blog/revisiting-the-outbox-pattern) — MEDIUM: poor ordering у polling по timestamp; sequence_number.
- [Kafka ordering in the real world — DEV](https://dev.to/amitjkamble/kafka-ordering-in-the-real-world-how-to-scale-without-killing-performance-37fo) — MEDIUM: per-aggregate vs global ordering, `max.in.flight`, идемпотентный продюсер.
- [Event Sourcing Pattern — Microsoft Learn](https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing) — HIGH: append-only лог = источник истории/аудита; снапшот ≠ история.
- Внутренние источники проекта: SEED-001 (идентичность/reconciliation), SEED-002 (audit/actor forward-compat), L2-ARCHITECTURE «Анти-паттерны» + «Инварианты для записи», PROJECT.md Key Decisions v3.0 — HIGH (канон проекта).

---
*Pitfalls research for: DC hardware inventory identity/reconciliation + Kafka event-backbone (producer-only) via transactional outbox/relay*
*Researched: 2026-06-26*
