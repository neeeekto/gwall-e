// Package kafka — generic-плумбинг администрирования Kafka для всех сервисов gwall-e.
//
// Здесь живёт ТОЛЬКО механика (admin-клиент + декларативный провижн топиков); имена
// топиков, cleanup-политики и прочая доменная топология — забота вызывающего сервиса
// (например, services/inventory/internal/kafka/topology). pkg/kafka их не знает.
package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// TopicSpec — декларативное описание одного топика: имя и набор cleanup-конфигов.
// Имена и политики — домен вызывающего; pkg/kafka лишь применяет спецификацию.
type TopicSpec struct {
	Name    string
	Configs map[string]*string
}

// NewAdminClient поднимает kgo-клиент со seed-брокерами и оборачивает его в kadm для
// админ-операций. Возвращает admin-клиент и close-функцию (закрытие нижележащего
// kgo-клиента) — вызывающий обязан вызвать close через defer.
func NewAdminClient(brokers []string) (adm *kadm.Client, close func(), err error) {
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return nil, nil, fmt.Errorf("create kafka client: %w", err)
	}
	return kadm.NewClient(cl), cl.Close, nil
}

// EnsureTopics провижнит каждый топик из specs с заданными числом партиций и replication
// factor. Идемпотентность поверх уже существующих топиков обеспечивает вызывающий (kadm
// возвращает ошибку на дубликат).
func EnsureTopics(ctx context.Context, adm *kadm.Client, partitions int32,
	replicationFactor int16, specs ...TopicSpec,
) error {
	for _, s := range specs {
		if _, err := adm.CreateTopics(ctx, partitions, replicationFactor, s.Configs, s.Name); err != nil {
			return fmt.Errorf("create topic %s: %w", s.Name, err)
		}
	}
	return nil
}

// StringPtr — хелпер для cleanup-конфигов, чтобы сервисам не тащить kadm ради одной обёртки.
func StringPtr(s string) *string { return kadm.StringPtr(s) }
