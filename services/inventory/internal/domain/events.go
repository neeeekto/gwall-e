package domain

import "time"

// DomainEvent — голый семантический факт доменного перехода (D-14). Несёт ТОЛЬКО
// предметную семантику (тип события + идентификатор сущности), но НЕ envelope-мету
// (eventId/occurredAt/actor): та заполняется на границе usecase (06-04 envelope.go),
// а не в агрегате. Конкретные события (HostRegistered, ProjectCreated, ...) реализуют
// этот интерфейс.
type DomainEvent interface {
	// EventType возвращает стабильное имя семантического события (D-13), напр.
	// "HostRegistered" — оно же ложится в EventEnvelope.EventType и далее в Kafka.
	EventType() string
	// EntityID возвращает строковый идентификатор сущности-источника события
	// (будущий Kafka-ключ = внутренний ID, не FQDN/INV/MAC).
	EntityID() string
}

// Actor — атрибуция инициатора события (EVT-02/D-15/SEED-002): кто и через какой канал
// вызвал переход. Source принимает значения human|api|integration|system — это форма,
// а не точка валидации (валидация — на границе usecase).
type Actor struct {
	ID     string // идентификатор инициатора (пользователь, сервис, интеграция)
	Source string // канал: human | api | integration | system
}

// EventEnvelope — транспортная форма доменного события (D-14/D-15/EVT-02). Оборачивает
// голый DomainEvent envelope-метой для outbox/relay/Kafka. Forward-compatible с первого
// дня: eventId/version/actor живут в форме до relay-кода (Phase 8). Envelope-мета здесь —
// ТОЛЬКО форма; её заполняет граница usecase (06-04), НЕ агрегат (D-14).
type EventEnvelope struct {
	EventID    string      // уникальный id события (заполняет IDGenerator на границе)
	EntityID   string      // id сущности-источника (Kafka-ключ)
	EventType  string      // имя семантического события (D-13)
	Version    int         // версия агрегата на момент события (оптимистичная блокировка)
	OccurredAt time.Time   // момент возникновения (заполняет Clock на границе)
	Actor      Actor       // атрибуция инициатора (EVT-02)
	Payload    DomainEvent // голый доменный факт
}
