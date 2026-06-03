package domain

// AggregateRoot — базовая структура для накопления доменных событий.
// Агрегат не публикует события сам — он их накапливает.
// Handler забирает события после сохранения и публикует через EventPublisher.
// Это гарантирует: событие публикуется только если транзакция прошла успешно.
type AggregateRoot struct {
	events []DomainEvent
}

// RecordEvent добавляет событие в очередь агрегата.
func (a *AggregateRoot) RecordEvent(e DomainEvent) {
	a.events = append(a.events, e)
}

// PullEvents возвращает накопленные события и очищает очередь.
// Вызывается один раз — в command handler после сохранения.
func (a *AggregateRoot) PullEvents() []DomainEvent {
	events := make([]DomainEvent, len(a.events))
	copy(events, a.events)
	a.events = nil
	return events
}
