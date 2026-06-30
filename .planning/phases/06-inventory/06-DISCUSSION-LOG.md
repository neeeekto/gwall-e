# Phase 6: Доменная модель Inventory - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-30
**Phase:** 6-inventory
**Areas discussed:** Срез Phase 6 ↔ Phase 7, Агрегаты и локации, ЖЦ/идентичность/конфликты, Таксономия событий + envelope

---

## Срез Phase 6 ↔ Phase 7

### Глубина вертикали
| Option | Description | Selected |
|--------|-------------|----------|
| Домен + usecases на фейк-портах | Агрегаты/VO/события/фабрики + interactor-usecases на портах, тест на in-memory/mock | ✓ |
| Только доменное ядро | Агрегаты/VO/события/порты без usecases; interactor'ы — в Phase 7 | |

**User's choice:** Домен + usecases на фейк-портах
**Notes:** SC1/SC3/SC5 формулируются через usecase → нужны для верифицируемости фазы.

### Форма usecase
| Option | Description | Selected |
|--------|-------------|----------|
| Уже через UoW + Outbox (фейки) | usecase оборачивает в uow.Do + outbox.Append как в architecture.md; Phase 7 = свап реализаций | ✓ |
| Persistence-agnostic, UoW — Phase 7 | usecase зовёт repo.Save + PullEvents; UoW/outbox добавляются позже (переструктурирование) | |

**User's choice:** Уже через UoW + Outbox (фейки)
**Notes:** Цель — Phase 7 чистый свап, не переписывание; эталон виден целиком сейчас.

### Тест-дублёры
| Option | Description | Selected |
|--------|-------------|----------|
| Гибрид: фейки + моки точечно | In-memory фейки с состоянием + mockery для узких портов | |
| Только mockery-моки | Все порты — сгенерированные моки, expectations на вызовы | ✓ |

**User's choice:** Только mockery-моки
**Notes:** Домен/инварианты тестируются прямыми unit-тестами без моков.

---

## Агрегаты и локации

### Локации DC/Module/Rack
| Option | Description | Selected |
|--------|-------------|----------|
| Три агрегата, parent по ID | Каждый свой корень с CRUD; Module→DCID, Rack→ModuleID; маленькие агрегаты | ✓ |
| Один агрегат-дерево Location | DC владеет Module'ами/Rack'ами вложенно; крупный агрегат | |

**User's choice:** Три агрегата, parent по ID

### Типизация внутренних ID
| Option | Description | Selected |
|--------|-------------|----------|
| Типизированные ID-VO на агрегат | HostID/ProjectID/… как обёртки над uuid.UUID; компайл-тайм защита | ✓ |
| Сырой uuid.UUID везде | Меньше типов, но легко перепутать ссылки | |

**User's choice:** Типизированные ID-VO на агрегат

### Изменяемость HostHardware
| Option | Description | Selected |
|--------|-------------|----------|
| Единый immutable VO, замена целиком | Неизменяемый VO со всеми компонентами; изменение = новый VO → одно HostHardwareChanged | ✓ |
| Мутабельные под-компоненты | Отдельные компоненты меняются с per-компонент событиями | |

**User's choice:** Единый immutable VO, замена целиком

---

## ЖЦ, идентичность, конфликты

### decommissioned vs deleted
| Option | Description | Selected |
|--------|-------------|----------|
| Оба — терминальные состояния lifecycleState | shadow/registered/decommissioned/deleted в одном поле (soft-delete) | |
| deleted — ортогональный tombstone-флаг | lifecycleState без deleted; deleted как отдельный флаг/deletedAt | |
| **Hard-удаление (free-text)** | lifecycleState ∈ {shadow,registered,decommissioned}; deleted = HostDeleted-событие + hard-removal | ✓ |

**User's choice:** (free-text) «Мы кажется решили что у нас будет hard удаление, а не soft» → подтверждено
после сверки с PITFALLS.md: `lifecycleState ∈ {shadow, registered, decommissioned}`, `deleted` = `HostDeleted` + hard-removal.
**Notes:** decommissioned = виден со статусом (НЕ tombstone); deleted = записи нет, история только в append-only `*.events`,
tombstone в `*.state` (Phase 8). Канон: Pitfall 2 + чеклист «Looks Done But Isn't».

### Граф переходов и точка входа
| Option | Description | Selected |
|--------|-------------|----------|
| Гибкий вход + терминальный decommission | вход shadow или registered; shadow→registered, *→decommissioned (терминально); delete из любого | ✓ |
| Строго линейный | только shadow→registered→decommissioned; вход всегда shadow | |

**User's choice:** Гибкий вход + терминальный decommission

### FQDN-конфликт и advisory-matching
| Option | Description | Selected |
|--------|-------------|----------|
| Инвариант в usecase + advisory-порт-хук | FQDN-конфликт = доменный инвариант (типизированный конфликт); advisory = порт-хук + заглушка | ✓ |
| Считать кандидатов сразу в usecase | Составной матч (SEED-001) прямо в usecase | |

**User's choice:** Инвариант в usecase + advisory-порт-хук
**Notes:** Полный матчинг → отдельный сервис (SEED-001); авто-restore/merge запрещён.

---

## Таксономия событий + envelope

### Гранулярность
| Option | Description | Selected |
|--------|-------------|----------|
| Одно событие на операцию | HostRegistered (с начальными hardware/location) + отдельные на отдельные операции | ✓ |
| Дробить операцию на атомарные факты | HostRegistered + HostHardwareSet + HostPlacedInRack… | |

**User's choice:** Одно событие на операцию
**Notes:** Семантические имена-факты, фиксируются в DOC-07.

### Сборка envelope
| Option | Description | Selected |
|--------|-------------|----------|
| Факты в домене + envelope на границе | Агрегат эмитит голые факты + version; eventId/occurredAt/actor навешиваются между PullEvents и Append | ✓ |
| Envelope целиком в домене | Агрегат имеет Clock+IDGen порты, actor передаётся в агрегат | |

**User's choice:** Факты в домене + envelope на границе
**Notes:** Домен-ядро чистое; actor (транспортная identity) не входит в агрегат; ложится на «actor из interceptor» (Phase 7).

---

## Claude's Discretion

- Точный состав/формулировки DOC-07 glossary (в рамках D-08…D-15).
- Какие канон-слои SVC-01 получают код в Phase 6 (ожидаемо domain + usecases + query-порт; repositories/api/cron/app — фейки/скелет до Phase 7).
- Project-агрегат: операции/события (rename, смена Owner, delete) и инвариант уникальности позиции host↔rack.
- Имена пакетов, сигнатуры портов, форма ID-VO, структура фабрик.

## Deferred Ideas

- Полный advisory-matching движок → отдельный интеграционный сервис (SEED-001).
- Persistence/UoW(Mongo)/Outbox-коллекция/gRPC/query-на-Mongo/partial FQDN-index → Phase 7.
- protobuf-схемы + relay→Kafka + dual-topic + tombstone → Phase 8.
- Топология connections + read-model + внешние HW-модули → Phase 9.
- VM/VMGroup, sync, Audit, Access → будущие эпики.
