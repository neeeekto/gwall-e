# Feature Research

**Domain:** Hardware-инвентаризация ДЦ (DCIM / inventory source-of-truth) — домен Inventory сервиса gwall-e
**Researched:** 2026-06-26
**Confidence:** HIGH (модель данных зрелых DCIM подтверждена официальной докой NetBox/Nautobot, DMTF Redfish, патентами и практикой Wikimedia); MEDIUM по конкретным «должны/не должны» в нашем solo-режиме (контекст-специфично)

> **Как читать этот документ.** Мы НЕ строим DCIM-продукт. gwall-e — оркестратор, а Inventory —
> источник истины об **идентичности / ЖЦ / железе / локации** хоста для домен-first модели в solo-режиме.
> Поэтому многие «table stakes для DCIM» (мониторинг питания в реальном времени, cooling-телеметрия,
> 3D-визуализация) для нас — **anti-features**: либо чужой домен (Health/Network), либо вне scope.
> Здесь «table stakes» = **то, без чего доменная модель Inventory неполна** (зрелые DCIM это моделируют,
> и отсутствие приведёт к переделке схемы), а не «то, что ждёт конечный пользователь UI».

---

## Проверка нашего черновика по зрелым DCIM (сводка)

| Наш черновик | Что делают зрелые DCIM (NetBox / Nautobot / Redfish) | Вердикт |
|---|---|---|
| `Project` (группа хостов под Owner) | **Tenant** — «ownership of an object, single tenant per object»; Tenant Groups для иерархии; влияет на uniqueness имени | ✓ Совпадает. Owner ≈ tenant. Рассмотреть Tenant Group-аналог позже |
| `Host` (железный сервер) | **Device** (instance от Device Type), смонтирован в Rack, занимает N юнитов на face (front/rear) | ✓ Совпадает, но не хватает **face/depth/position-в-юнитах** (см. gap L2) |
| Внешние HW-модули (полки/GPU) без owner | **Module / Device Bay / Child Device** + plugin `netbox-inventory` (asset в storage без привязки). NetBox: child-device — first-class независимый managed-объект | ✓ Концепция «самостоятельный модуль» легитимна. См. gap T1 (как моделить связь) |
| `HostHardware{ ... RAM/CPU/Drives... }` | Redfish: Systems (logical: CPU/Memory/Storage/NIC) + Chassis (physical: PSU/Fan/Sensors) + UpdateService/FirmwareInventory | ⚠ Покрыто частично — крупные gaps по NIC/PSU/firmware (см. HW-gaps) |
| Lifecycle: shadow/registered/decommissioned/deleted | NetBox: offline/active/planned/staged/failed/inventory/decommissioning (кастомизируемо) | ⚠ Не хватает промежуточных состояний (см. gap LC) |
| Локация DC→Module→Rack + юнит | Site→Location→Rack (height U, face, width 19/23", facility ID); Rack — power/space/weight/cooling capacity | ⚠ Атрибуты стойки бедны (см. gap L) |
| Топология `connections` хост↔модуль | Power Panel→Feed→PDU→Port; directed dependency graph; «impacted vs failed» маркировка; fault domains | ✓ Направление верное; нужно обобщить тип связи (см. gap T) |

---

## Feature Landscape

### Table Stakes (без этого доменная модель Inventory неполна)

| Feature | Why Expected (зрелые DCIM моделируют это) | Complexity | Notes |
|---|---|---|---|
| **Host = инстанс с identity + статусом + локацией** | NetBox Device — ядро DCIM | LOW | У нас есть: постоянный `ID`, статус, ссылка на стойку/позицию |
| **Project/Tenant как контейнер владения** | Tenant — single-owner группировка; влияет на uniqueness | LOW | У нас есть `Project.owner` (непрозрачный внешний ID) |
| **Иерархия локации DC→Module→Rack** | Site→Location→Rack — фундамент | LOW | Есть. См. gap про face/position |
| **Серийники/INV/asset-tag на хосте и компонентах** | manufacturer + part-id + serial на каждом inventory item | LOW | Есть `Inv(string)`+`serial`+`vendor`+`model`+`lot` на компонентах |
| **Hardware-компоненты: CPU / RAM / Drives** | Redfish Systems: Processors/Memory/Storage | MEDIUM | Есть. Не хватает **NIC, PSU, firmware** (HW-gaps) |
| **Сетевые интерфейсы (NIC) с MAC** | Redfish EthernetInterfaces; NIC asset (MAC) — обязательно для матчинга | MEDIUM | **GAP HW1** — есть только `MACs[]` плоско, нет NIC-сущности (slot/model/speed/MAC↔порт) |
| **Lifecycle-статусы с разделением списания и удаления записи** | NetBox: decommissioning ≠ inventory ≠ failed; «человек решает failed vs offline» | MEDIUM | Есть `decommissioned ≠ deleted`. См. gap LC по промежуточным |
| **История изменений (audit trail)** | Серийник/asset-tag — mutable; теряется история при swap (известная критика NetBox) | MEDIUM | У нас сильнее: история на event-backbone, НЕ soft-delete-флаг |
| **Топология зависимостей (что зависит от X)** | Directed dependency graph; «what depends on this PDU/UPS» | HIGH | Есть `connections` + read-model. См. gaps T1–T3 |
| **Uniqueness FQDN среди active** | NetBox: имя уникально в site, если нет tenant | LOW | Есть (SEED-001) |

### Differentiators (наша осознанная позиция, расходящаяся с типовым DCIM)

| Feature | Value Proposition | Complexity | Notes |
|---|---|---|---|
| **Identity как НАШ постоянный ID, без авто-мерджа** | NetBox device — не asset: serial/tag мутабельны → теряется история при swap. Мы решаем это by design (SEED-001) | MEDIUM |差ferentiator: «нельзя надёжно ответить тот же ли хост» — свойство реальности, моделируем явно |
| **Событийная история вместо soft-delete-флага** | Re-add = новый ID; restore-with-merge запрещён → нет ложного матча на рециклинге FQDN/смене материнки | MEDIUM | Решает известную боль NetBox (потеря независимых историй) |
| **Самостоятельные HW-модули без owner (полки/GPU)** | Общее железо инфраструктурно (каскад отказа), но gwall-e им не управляет → отдельная сущность, не child хоста | MEDIUM | Соответствует NetBox child-device-как-first-class, но без tenant |
| **Match как советочный (human-in-the-loop), не авто** | Составной матч INV+FQDN+MAC+локация+окно → кандидаты человеку | HIGH | Это для будущего sync-эпика (SEED-001); в Inventory — только хук |
| **Compacted-снапшот идентичности по entityID** | «Снапшот идентичности» для онбординга нового домена бесплатно (Kafka log-compaction) | MEDIUM | L2-инвариант; differentiator для платформы |

### Anti-Features (выглядят нужными, но это чужой домен / вне scope)

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| **Power-телеметрия реального времени** (kW, draw, утилизация) | DCIM «table stakes»; NetBox считает power utilization рейка | Runtime-метрики — домен **Health**; в Inventory будет dual-source-of-truth | Inventory держит **capacity/паспортные** атрибуты (max power), не measured draw |
| **Cooling/thermal-мониторинг** (BTU, sensors, PUE) | DCIM-фича; sensors в Redfish Chassis | Чужой домен (Health); телеметрия не идентичность | В Inventory — максимум паспортная cooling-capacity стойки (опц., см. gap L) |
| **Cable management / patch-панели / порт-маппинг** | NetBox моделирует кабели как first-class | Сетевая топология — домен **Network** (свитчи/VLAN/IPAM) | Inventory ограничивается power/parent-child топологией, не network-кабелями |
| **Каскадные действия по топологии** (drain/shutdown зависимых) | «Failed PDU → reroute load» в DCIM | Это **операции** → домены Actions/Scenarios/Orchestrator | Inventory отдаёт **read-model** «что зависит от X»; решение действует другой домен |
| **Provisioning / наливка / профилировка как состояния хоста** | NetBox staged/planned ведут к provisioning | «Тяжёлые операции» — внешние провайдеры (Integrations), процесс — Actions | В Inventory только **факт-статусы** ЖЦ, не process-state наливки |
| **Procurement / PO / RMA / warranty-lifecycle** | NetBox Labs Asset Lifecycle (BOM/PO/RMA/EOX) | Не наш Core Value; раздувает домен | Опционально позже отдельным эпиком; в v3.0 — out |
| **3D-визуализация / rack elevation UI** | Siemens/Sunbird фича | Фронтенд после бэкенд-доменов; не доменная модель | Отложено (PROJECT: фронтенд после доменов) |
| **VM / VMGroup** | Парк включает VM | Модель работы с VM не ясна | Отложено в отдельный эпик (PROJECT out-of-scope) |
| **Sync из внешней инвентори** | Реальное наполнение | Отдельный интеграционный сервис; reconciliation сложен | SEED-001; solo-режим в v3.0, хук на match — позже |

---

## Конкретные пробелы нашей hardware-модели (HW-gaps)

> Запрос явно требовал назвать пробелы конкретно. По Redfish-модели (Systems logical + Chassis physical
> + UpdateService/FirmwareInventory) наш черновик `HostHardware{Name,Platform,IPMIMac,Motherboard,MACs[],RAM[],CPU[],Drives[]}` имеет следующие дыры:

| # | Gap | Что именно отсутствует | Почему важно | Complexity | Рекомендация |
|---|---|---|---|---|---|
| **HW1** | **NIC как компонент** | `MACs[]` лежит плоско, без сущности NIC | Redfish моделит NIC отдельно (slot/model/vendor/speed, несколько MAC на одну карту, IPMI-MAC ≠ data-MAC). MAC — ключ советочного матча (SEED-001) | MEDIUM | Ввести `NIC{slot,model,vendor,serial,Inv,speed,MACs[]}`; `MACs[]` верхнего уровня убрать (выводить из NIC) |
| **HW2** | **PSU (блоки питания)** | Полностью отсутствуют | Redfish Chassis: PSU с Manufacturer/Model/Serial; PSU определяет power-зависимость и redundancy (primary/redundant feed) | MEDIUM | `PSU{slot,model,vendor,serial,Inv,maxPowerW,inputType}`; связать с power-топологией (T2) |
| **HW3** | **Firmware-версии** | Нет ни одного firmware-поля | Redfish UpdateService/FirmwareInventory — отдельный first-class слой: версии BMC/BIOS/NIC/storage-controller/drive. Нужно для compliance/EOL | MEDIUM | Добавить `firmwareVersion` на компоненты, у которых есть прошивка (BMC, BIOS/Motherboard, NIC, Drive, RAID); либо общий `Firmware[]` по компонентам |
| **HW4** | **BMC/BMC как сущность отдельно от IPMI MAC** | Есть только `IPMIMac` (плоское поле) | Redfish Managers: BMC = отдельный управляющий компонент (vendor/model AST2520, firmware). `IPMIMac` — лишь его MAC | LOW-MEDIUM | `BMC{model,vendor,firmwareVersion,mac}`; `IPMIMac` поглощается в BMC |
| **HW5** | **Storage controller / RAID** | `Drives[]` есть, контроллера нет | Redfish Storage: StorageControllers (Broadcom/Marvell) с firmware; влияет на видимость дисков, отдельная прошивка | MEDIUM | `StorageController{model,vendor,serial,Inv,firmwareVersion}` + связь drive↔controller |
| **HW6** | **BIOS/UEFI как именованный компонент с версией** | `Motherboard` есть, BIOS-версии нет | Redfish: BIOS-firmware отдельно от материнки | LOW | Поле `biosVersion` на Motherboard или отдельный компонент |
| **HW7** | **GPU как внутренний компонент хоста** | Внешние GPU = модуль (ок), но **внутренние GPU** не описаны | Redfish Systems: Processors включает GPU/accelerator; для AI-парка критично | MEDIUM | `GPU{slot,model,vendor,serial,Inv,memoryGB}` (внутр.) — отличать от внешнего GPU-модуля |
| **HW8** | **Платформенный паспорт / chassis-уровень** | `Platform` — строка; нет chassis serial/model отдельно | Redfish Chassis: PartNumber/SKU/AssetTag всего шасси; отличается от материнки | LOW | `chassisModel/chassisSerial/chassisInv` (паспорт шасси отдельно от Motherboard) |

**Принцип для всех HW-полей:** наш существующий паттерн компонента `{slot, model, vendor, lot, serial, Inv}`
— хороший базовый VO. Не хватает (а) **`firmwareVersion`** там, где есть прошивка, (б) **типизированных компонентов NIC/PSU/BMC/StorageController/GPU** вместо плоских полей.
Решение «hardware = VO внутри Host» (Key Decision v3.0) остаётся в силе — это структура VO, а не новые агрегаты.

---

## Конкретные пробелы lifecycle (LC-gaps)

> Наш черновик: `shadow → registered(active) → decommissioned → deleted`. NetBox даёт
> offline/active/planned/staged/failed/inventory/decommissioning. Сопоставление:

| Наш статус | NetBox-аналог | Комментарий |
|---|---|---|
| `shadow` | planned / staged | Хост заведён, но не «боевой» — ок |
| `registered (active)` | active | ✓ |
| `decommissioned` | inventory / decommissioning | Списан, но запись жива — ✓ (differentiator: ≠ deleted) |
| `deleted` | (нет; в NetBox это физ. удаление записи) | ✓ Запись убрана; история — в событиях |

| # | LC-gap | Чего не хватает | Почему | Complexity | Рекомендация |
|---|---|---|---|---|---|
| **LC1** | **Нет состояния «сломан/в ремонте» (failed)** | NetBox явно различает offline vs failed; «человек решает» | Парк 150k — хосты ломаются; «active» с отказом ≠ «списан». Нужно для read-model зависимостей | MEDIUM | Добавить `failed`/`maintenance` как факт-статус (не process). Решить: статус или отдельный атрибут health-flag (граница с Health!) |
| **LC2** | **`decommissioning` как переходное** vs мгновенное `decommissioned` | NetBox: decommissioning = «в процессе», т.к. снос — saga с вето (L2) | Decommission — гарантированная saga (Orchestrator); Inventory должен отражать «в процессе сноса» | MEDIUM | Промежуточный `decommissioning` (in-progress) перед `decommissioned`; согласовать с decommission-saga |
| **LC3** | **`shadow` слишком обобщён** | NetBox разделяет planned (ещё нет железа) vs staged (железо стоит, не активирован) | Разные операционные смыслы; влияет на «что зависит» и FQDN-uniqueness | LOW-MEDIUM | Уточнить: достаточно ли одного `shadow` или нужны planned/staged. Domain-first: решить при глоссарии (DOC-07) |
| **LC4** | **Граница «факт-статус ЖЦ» vs «process/provisioning-state»** | NetBox смешивает; мы не должны | provisioning/наливка — Actions (anti-feature выше). Inventory держит только факт-статусы | — | **Решение для модели:** статусы Inventory = факты идентичности/ЖЦ; всё process-state — наружу. Зафиксировать в DOC-07 |

---

## Конкретные пробелы локации (L-gaps)

> Наш черновик: DC→Module→Rack + позиция(юнит); атрибуты стойки (питание / дизель-генератор).

| # | L-gap | Чего не хватает | Почему (DCIM) | Complexity | Рекомендация |
|---|---|---|---|---|---|
| **L1** | **PDU / power feed как объекты локации** | Питание стойки — атрибут, не сущность | NetBox: Power Panel→Feed→PDU(device)→Outlet. Без PDU топология «что зависит от питания» неполна (нужно для T) | MEDIUM | Ввести PDU/feed как узлы локации/топологии (минимально: feed с primary/redundant) |
| **L2** | **Position: face (front/rear) + depth + диапазон юнитов** | Есть «позиция (юнит)», но не face/depth/высота в U | NetBox: device занимает N U на face; 0U-устройства (верт. PDU); half/full depth | LOW | `position{startUnit, heightU, face, depth}`; поддержать 0U |
| **L3** | **Capacity-атрибуты стойки** | Только питание/генератор | DCIM рейк: space(U), power capacity(kW), weight/max-load, cooling capacity | LOW-MEDIUM | Паспортные (не measured!): `powerCapacityW`, `maxWeightKg`, `heightU`, опц. `coolingCapacity`. Measured — anti-feature (Health) |
| **L4** | **Redundancy питания (A/B feed, fault domain)** | Дизель-генератор есть, но не A/B-резерв | DCIM: каждый feed primary/redundant; rack across fault domains; PDU от 2 UPS | MEDIUM | Моделить feed-leg/redundancy и принадлежность fault-domain для read-model «blast radius» |
| **L5** | **Facility ID / внешний ID стойки** | Нет | NetBox: facility ID (ДЦ присваивает свой ID арендуемой стойке) ≠ внутреннее имя | LOW | `facilityId(string)` на стойке (как внешние ID — string, по Key Decision) |
| **L6** | **UPS/ATS/PDU-цепочка над генератором** | Генератор есть, цепочка нет | DCIM power chain: utility→ATS→UPS→PDU→rack. Генератор — один узел из многих | MEDIUM | Решить глубину: минимум generator+UPS+feed как узлы топологии (для T) |

---

## Конкретные пробелы топологии зависимостей (T-gaps)

> Наш черновик: `connections` хост↔модуль (тип подключения) + read-model «что зависит от X».

| # | T-gap | Чего не хватает | Почему (DCIM) | Complexity | Рекомендация |
|---|---|---|---|---|---|
| **T1** | **Типизация connection (power / data / storage / parent-child)** | «тип подключения» не перечислен | DCIM моделит power-topology ≠ data ≠ storage отдельно. Полка подключается storage-линком, GPU-box — PCIe/power | MEDIUM | Enum типов: `power`, `storage`, `data`, `parent-child`(chassis-bay), `pcie`. Хост↔модуль уточнить тип |
| **T2** | **Power-зависимость хоста (через PSU→feed→PDU→UPS→generator)** | connections только хост↔модуль | «What depends on PDU/UPS/generator» требует power-цепочку как граф | HIGH | Power-топология как directed graph: PSU→feed→PDU→UPS→generator; read-model snizu-vverh и sverhu-vniz |
| **T3** | **`impacted` vs `failed` в read-model** | read-model «что зависит» бинарна | Патенты DCIM: даже редундантный потомок помечается `impacted` (повышенный риск), не `failed` | MEDIUM | Read-model различает «прямо зависит / зависит через резерв (impacted)». Полезно для blast-radius |
| **T4** | **Направление обхода: dependency-of vs dependency-on** | read-model «что зависит от X» — одно направление | DCIM: per-asset dependency-on chart И dependency-of chart | LOW-MEDIUM | Двунаправленный read-model: «от чего зависит X» и «что зависит от X» |
| **T5** | **Связь модуля без owner с несколькими хостами** | Полка/GPU = самостоятельный модуль | Общий модуль (дисковая полка на 2 хоста) → каскад отказа на оба | MEDIUM | Connection many-to-many: один модуль ↔ N хостов (shared-resource fault domain) |

---

## Feature Dependencies

```
[Project/Owner identity]
    └──requires──> [постоянный ID + event-backbone]   (single identity owner, L2)

[Host hardware-модель (NIC/PSU/BMC/firmware)]
    └──requires──> [Host identity]
    └──enhances──> [советочный матч SEED-001]          (MAC/serial — ключи матча)

[Топология connections]
    └──requires──> [Host identity] + [Module identity] + [Location]
    └──requires──> [PSU + PDU/feed]                    (T2: power-цепочка)
         └──requires──> [Локация: PDU/UPS/generator как узлы]  (L1/L6)

[Read-model «что зависит от X» (impacted/failed, оба направления)]
    └──requires──> [Топология connections с типами]    (T1/T3/T4)

[Lifecycle decommissioning (in-progress)]
    └──requires──> [decommission-saga / Orchestrator]  (L2-видение; в Inventory — статус-хук)

[Event-backbone: семантические события + compacted snapshot]
    └──requires──> [outbox→relay→Kafka] (канон v1.0)
    └──enhances──> [онбординг будущих доменов] (backfill)

[Sync / reconciliation]  ──conflicts(by design)──> [авто-мердж идентичности]
    (SEED-001: match только советочный; restore-with-merge запрещён)
```

### Dependency Notes

- **Топология требует PSU и PDU/feed:** без них (HW2, L1) нельзя построить power-зависимость (T2) — «что зависит от генератора» обрывается на стойке. Это связывает HW-gaps и L-gaps в одну фазу с топологией.
- **Hardware-матч усиливает SEED-001:** NIC.MAC и serial — ключи советочного матча. Если NIC моделим плоско (HW1), матчинг ослаблен. Но сам матч — будущий sync-эпик; в Inventory нужен только корректный hardware-VO.
- **decommissioning(in-progress) зависит от Orchestrator:** Inventory не владеет saga; статус отражает внешний процесс. В solo-режиме v3.0 переход может быть ручным; модель должна допускать промежуточное состояние.
- **Конфликт by design:** авто-мердж/restore-with-merge запрещены — это не «недоделка», а инвариант (SEED-001). Любая будущая sync-фича строится вокруг советочного матча.

---

## MVP Definition (для REQUIREMENTS.md, solo + domain-first)

### Launch With (v3.0 — доменная модель)

- [ ] **Project + Owner (непрозрачный внешний ID)** — контейнер владения; `ProjectCreated`
- [ ] **Host identity + lifecycle** — постоянный ID; статусы shadow/active/decommissioning/decommissioned/deleted; **+failed/maintenance (LC1)**; FQDN uniqueness среди active
- [ ] **HostHardware VO (расширенный)** — базовый паттерн компонента + **NIC(HW1), PSU(HW2), BMC(HW4), firmwareVersion(HW3), StorageController(HW5), внутр. GPU(HW7), chassis-паспорт(HW8)**
- [ ] **Самостоятельные HW-модули без owner** — дисковые полки / внешние GPU; внешние ID = string
- [ ] **Локация DC→Module→Rack + position{startUnit,heightU,face,depth}(L2)** + паспортные capacity(L3) + facilityId(L5)
- [ ] **Топология connections с типами (T1)** + many-to-many модуль↔хосты (T5)
- [ ] **Read-model зависимостей: оба направления (T4)** — «что зависит от X» / «от чего зависит X»
- [ ] **Event-backbone** — outbox→relay→Kafka, семантические события + compacted snapshot по entityID, actor/initiator
- [ ] **DOC-07 glossary** — ubiquitous language (зафиксировать факт-статусы vs process-state, LC4)

### Add After Validation (v3.x)

- [ ] **Power-топология PSU→feed→PDU→UPS→generator (T2, L1, L6)** — триггер: когда нужен реальный blast-radius «что зависит от генератора»
- [ ] **impacted vs failed в read-model (T3)** — триггер: появление потребителя read-model (Health/Scenarios)
- [ ] **Promejutochnye lifecycle planned/staged (LC3)** — триггер: если solo-наполнение покажет нужду различать
- [ ] **Fault-domain / A-B feed redundancy (L4)** — триггер: запрос на анализ резервирования

### Future Consideration (отдельные эпики)

- [ ] **Sync + советочный reconciliation (SEED-001)** — отдельный интеграционный сервис
- [ ] **VM / VMGroup** — модель работы не ясна
- [ ] **Procurement / warranty / EOX lifecycle** — не Core Value
- [ ] **Cooling/thermal паспорт стойки** — если понадобится capacity-планирование
- [ ] **Audit-домен** — consumer событий (SEED-002)

## Feature Prioritization Matrix

| Feature | User/Domain Value | Implementation Cost | Priority |
|---|---|---|---|
| Host identity + lifecycle (incl. failed LC1) | HIGH | MEDIUM | P1 |
| HostHardware VO (NIC/PSU/BMC/firmware HW1-8) | HIGH | MEDIUM | P1 |
| Project + Owner ref | HIGH | LOW | P1 |
| Локация DC→Module→Rack + position (L2/L3/L5) | HIGH | MEDIUM | P1 |
| Самостоятельные HW-модули без owner | HIGH | MEDIUM | P1 |
| Топология connections типизированная (T1/T5) | HIGH | MEDIUM | P1 |
| Read-model зависимостей оба направления (T4) | HIGH | MEDIUM | P1 |
| Event-backbone (semantic + compacted) | HIGH | MEDIUM | P1 |
| Power-топология PSU→PDU→UPS→generator (T2) | MEDIUM | HIGH | P2 |
| impacted vs failed (T3) | MEDIUM | MEDIUM | P2 |
| planned/staged уточнение (LC3) | LOW | LOW | P3 |
| Fault-domain/A-B redundancy (L4) | MEDIUM | MEDIUM | P3 |
| Cooling-capacity паспорт | LOW | LOW | P3 |

## Competitor Feature Analysis

| Feature | NetBox / Nautobot | Redfish (DMTF) | Our Approach |
|---|---|---|---|
| Identity | Device — serial/tag мутабельны, **не asset**; теряет историю при swap | n/a (live BMC) | **Наш постоянный ID**, история на событиях, без авто-мерджа (SEED-001) — сильнее |
| Hardware components | Inventory Item (deprecated v4.3) → Module; firmware = custom field | First-class: Systems+Chassis+FirmwareInventory | Hardware = VO в Host; добавить firmware/NIC/PSU/BMC по образцу Redfish |
| Lifecycle | offline/active/planned/staged/failed/inventory/decommissioning (кастом) | n/a | shadow/active/decommissioning/decommissioned/deleted **+failed** (факт-статусы, без process-state) |
| Owner | Tenant (single-owner) + Tenant Groups | n/a | Project.owner = непрозрачный внешний ID; owner-роли → домен Access (2 слоя) |
| Локация | Site→Location→Rack (U/face/width/facilityId) + power/space/weight/cooling capacity | Chassis | DC→Module→Rack + position + паспортные capacity (measured → Health) |
| Топология | Power Panel→Feed→PDU; кабели; dependency graph; impacted/failed | n/a | connections типизированные + read-model; power-цепочка в P2; каскадные действия → чужие домены |
| Asset lifecycle (PO/RMA/warranty) | NetBox Labs commercial | n/a | Anti-feature (вне Core Value) |

## Sources

- [NetBox DCIM Devices](https://netboxlabs.com/docs/netbox/models/dcim/device/) — Device/face/U/child-device модель (HIGH)
- [NetBox Inventory Items](https://netboxlabs.com/docs/netbox/models/dcim/inventoryitem/) — компоненты, deprecation v4.3 → Modules (HIGH)
- [NetBox Power Feed](https://netboxlabs.com/docs/netbox/models/dcim/powerfeed/) — Panel→Feed→PDU→Outlet, primary/redundant (HIGH)
- [NetBox device status discussion #12855 / issue #3070](https://github.com/netbox-community/netbox/issues/3070) — lifecycle статусы, decommissioning (HIGH)
- [Wikimedia Server Lifecycle](https://wikitech.wikimedia.org/wiki/Server_Lifecycle) — практический decommission-workflow (MEDIUM)
- [Redfish (DMTF) — Wikipedia](https://en.wikipedia.org/wiki/Redfish_(specification)) — модели BIOS/Memory/Storage/NIC/firmware (HIGH)
- [Supermicro Redfish Firmware Inventory](https://www.supermicro.com/manuals/other/redfish-ref-guide-html/Content/general-content/firmware-inventory-update-service.htm) — FirmwareInventory, BMC/BIOS/NIC/storage версии (HIGH)
- [Graphcore BMC inventory monitoring](https://docs.graphcore.ai/projects/bmc-user-guide/en/latest/inventory-monitoring.html) — Systems vs Chassis, PSU/CPU serial примеры (HIGH)
- [Device42 Capacity Planning](https://www.device42.com/data-center-infrastructure-management-guide/data-center-capacity-planning/) — rack space/power/cooling/weight capacity (MEDIUM)
- [Sunbird Data Center Capacity](https://www.sunbirddcim.com/blog/data-center-capacity-how-measure-how-plan-and-how-much-left) — capacity KPIs, used vs available (MEDIUM)
- [USPTO 7711980 — topology-based failure impact](https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/7711980) — impacted vs failed, dependency-on/of, fault domains (MEDIUM)
- [arXiv 1610.04872 — DCIM Fault Detection Engine](https://arxiv.org/pdf/1610.04872) — power chain как directed graph, cascade (MEDIUM)
- [Nautobot Tenancy](https://docs.nautobot.com/projects/core/en/stable/core-functionality/tenancy/) — tenant = single-owner, uniqueness (HIGH)
- [ArnesSI/netbox-inventory](https://github.com/ArnesSI/netbox-inventory) — asset в storage, warranty, serial/tag sync, статусы (MEDIUM)

---
*Feature research for: Hardware-инвентаризация ДЦ (домен Inventory, gwall-e v3.0)*
*Researched: 2026-06-26*
