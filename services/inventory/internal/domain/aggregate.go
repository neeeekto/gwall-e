package domain

// aggregateBase — встраиваемая база всех агрегатов Inventory (D-14). Накапливает голые
// DomainEvent во внутреннем буфере и владеет version агрегата. Встраивается всеми пятью
// агрегатами (Host/Project/DC/Module/Rack); НЕ выносится в pkg/ — inventory-специфична
// (RESEARCH A3; MEMORY: shared-code-in-pkg только для truly-generic). Нижний регистр —
// поля и record() не экспортируются: события рождаются только внутри доменных переходов.
type aggregateBase struct {
	version int           // версия агрегата (растёт на каждом record — оптимистичная блокировка)
	events  []DomainEvent // буфер накопленных, ещё не слитых событий
}

// record копит доменное событие и инкрементит version в ОДНОЙ точке (Pitfall 3:
// version++ нигде больше). Вызывается доменными переходами агрегата.
func (b *aggregateBase) record(e DomainEvent) {
	b.version++
	b.events = append(b.events, e)
}

// PullEvents отдаёт накопленные события и очищает буфер (отдаёт-и-очищает, Pitfall 5):
// репозиторий сливает их в outbox в той же tx. Повторный вызов подряд вернёт пустой
// слайс — события не дублируются.
func (b *aggregateBase) PullEvents() []DomainEvent {
	out := b.events
	b.events = nil
	return out
}

// Version возвращает текущую версию агрегата (для оптимистичной блокировки в репозитории).
func (b *aggregateBase) Version() int {
	return b.version
}
