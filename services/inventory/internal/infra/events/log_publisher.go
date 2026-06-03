package events

import (
	"context"
	"log"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// LogEventPublisher — простая реализация EventPublisher для разработки/тестирования.
// Логирует события вместо отправки в брокер.
// В production заменяется на KafkaEventPublisher или NATSEventPublisher.
type LogEventPublisher struct{}

func NewLogEventPublisher() *LogEventPublisher {
	return &LogEventPublisher{}
}

var _ domain.EventPublisher = (*LogEventPublisher)(nil)

func (p *LogEventPublisher) Publish(_ context.Context, events ...domain.DomainEvent) error {
	for _, e := range events {
		log.Printf("[EVENT] %s at %s", e.EventName(), e.OccurredAt().Format("2006-01-02T15:04:05Z"))
	}
	return nil
}

// ChannelEventPublisher — реализация через Go-канал.
// Удобна для интеграционных тестов.
type ChannelEventPublisher struct {
	ch chan<- domain.DomainEvent
}

func NewChannelEventPublisher(ch chan<- domain.DomainEvent) *ChannelEventPublisher {
	return &ChannelEventPublisher{ch: ch}
}

var _ domain.EventPublisher = (*ChannelEventPublisher)(nil)

func (p *ChannelEventPublisher) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, e := range events {
		select {
		case p.ch <- e:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
