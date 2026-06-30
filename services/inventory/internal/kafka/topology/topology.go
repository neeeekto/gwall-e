// Package topology — единственный источник истины ДОМЕННОЙ топологии Kafka-топиков
// inventory (D-06).
//
// Здесь живут имена топиков, cleanup-политики и набор агрегатов. Механика провижна
// (admin-клиент, CreateTopics) вынесена в общую библиотеку pkg/kafka — этот пакет лишь
// описывает домен и делегирует. И тонкий CLI-провижн (cmd/), и интеграционные тесты зовут
// Bootstrap из этого пакета напрямую — дублировать конфиг топиков где-либо ещё запрещено
// (anti-pattern D-06).
package topology

import (
	"context"

	"github.com/twmb/franz-go/pkg/kadm"

	"github.com/gwall-e/pkg/kafka"
)

// aggregates — список агрегатов, под которые провижнятся топики (data-driven, D-10).
// Добавление нового агрегата (project/module/location в Phase 6+) = одна строка здесь.
var aggregates = []string{"host"}

const (
	// topicPrefix — общий префикс всех топиков сервиса inventory.
	topicPrefix = "inventory."
	// eventsSuffix — суффикс топика immutable-истории фактов (cleanup=delete, D-12).
	eventsSuffix = ".events"
	// stateSuffix — суффикс топика снапшота состояния (cleanup=compact, D-12).
	stateSuffix = ".state"

	// replicationFactor — single broker в дев/тест-стенде (D-07).
	replicationFactor = int16(1)

	// stateRetentionMs — окно жизни tombstone'ов в compacted-топике: 24h (≥24h по D-12).
	stateRetentionMs = "86400000"
)

// eventsTopic возвращает имя топика истории фактов для агрегата (inventory.<agg>.events).
func eventsTopic(aggregate string) string {
	return topicPrefix + aggregate + eventsSuffix
}

// stateTopic возвращает имя топика снапшота состояния для агрегата (inventory.<agg>.state).
func stateTopic(aggregate string) string {
	return topicPrefix + aggregate + stateSuffix
}

// eventsConfig — cleanup-конфиг топика *.events: только delete-политика (immutable-история, D-12).
func eventsConfig() map[string]*string {
	return map[string]*string{
		"cleanup.policy": kafka.StringPtr("delete"),
	}
}

// stateConfig — cleanup-конфиг топика *.state: compact + delete.retention.ms≥24h (снапшот, D-12).
func stateConfig() map[string]*string {
	return map[string]*string{
		"cleanup.policy":      kafka.StringPtr("compact"),
		"delete.retention.ms": kafka.StringPtr(stateRetentionMs),
	}
}

// specs строит декларативные спецификации топиков по списку агрегатов: на каждый агрегат —
// пара inventory.<agg>.events / .state с соответствующими cleanup-политиками.
func specs() []kafka.TopicSpec {
	out := make([]kafka.TopicSpec, 0, len(aggregates)*2)
	for _, aggregate := range aggregates {
		out = append(out,
			kafka.TopicSpec{Name: eventsTopic(aggregate), Configs: eventsConfig()},
			kafka.TopicSpec{Name: stateTopic(aggregate), Configs: stateConfig()},
		)
	}
	return out
}

// Bootstrap провижнит для каждого агрегата пару топиков inventory.<agg>.events / .state с
// соответствующими cleanup-политиками, делегируя механику в pkg/kafka. Число партиций —
// параметр (дев-дефолт задаёт вызывающий, D-11); replication factor — 1 (single broker,
// D-07). Идемпотентность поверх уже существующих топиков обеспечивает вызывающий (kadm
// возвращает ошибку на дубликат — её обработка делается на уровне CLI/теста по контексту).
func Bootstrap(ctx context.Context, adm *kadm.Client, partitions int32) error {
	return kafka.EnsureTopics(ctx, adm, partitions, replicationFactor, specs()...)
}
