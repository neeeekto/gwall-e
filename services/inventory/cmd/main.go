// Command inventory — тонкий bootstrap-CLI провижна Kafka-топологии inventory (D-09).
//
// CLI НЕ дублирует топологию: имена топиков, cleanup-политики и число партиций живут в
// пакете internal/kafka/topology (D-06). Здесь — только склейка: читаем env, поднимаем
// kgo/kadm-клиент и зовём общую topology.Bootstrap. Дёргается make-таргетом `topics`
// после `make dev-up`.
//
// ВНИМАНИЕ: cmd/ — это package main; `go build ./cmd` / `go build ./...` падает
// `build output "cmd" already exists` (Pitfall 2). Сборка валидируется через `go vet ./...`,
// запуск — через `go run ./cmd`.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gwall-e/services/inventory/internal/kafka/topology"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	// defaultBrokers — дев-дефолт адреса брокера (single broker compose-стенда, D-07).
	defaultBrokers = "localhost:9092"
	// defaultPartitions — дев/тест-дефолт числа партиций (D-11; >1 гоняет sticky-key partitioner).
	defaultPartitions = 6
)

func main() {
	// one-shot CLI: контекст всего процесса — Background.
	if err := run(context.Background()); err != nil {
		log.Fatalf("bootstrap topology: %v", err)
	}
}

// run выполняет провижн: парсит env → поднимает kgo/kadm-клиент → зовёт общую topology.Bootstrap.
func run(ctx context.Context) error {
	brokers := parseBrokers(os.Getenv("KAFKA_BROKERS"))

	partitions, err := parsePartitions(os.Getenv("KAFKA_PARTITIONS"))
	if err != nil {
		return err
	}

	// kgo-клиент со seed-брокерами; kadm оборачивает его для админ-операций (CreateTopics).
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return fmt.Errorf("create kafka client: %w", err)
	}
	defer cl.Close()

	adm := kadm.NewClient(cl)

	// вся топология (имена/политики/агрегаты) — в пакете topology (D-06), CLI её не дублирует.
	if err := topology.Bootstrap(ctx, adm, int32(partitions)); err != nil {
		return fmt.Errorf("provision topics (brokers=%v, partitions=%d): %w", brokers, partitions, err)
	}

	log.Printf("topology provisioned: brokers=%v, partitions=%d", brokers, partitions)
	return nil
}

// parseBrokers разбирает CSV-список брокеров из env; пустое значение → дев-дефолт.
func parseBrokers(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		raw = defaultBrokers
	}
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, p := range parts {
		if b := strings.TrimSpace(p); b != "" {
			brokers = append(brokers, b)
		}
	}
	return brokers
}

// parsePartitions разбирает число партиций из env; пустое значение → дев-дефолт (D-11).
func parsePartitions(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultPartitions, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("parse KAFKA_PARTITIONS=%q: %w", raw, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("KAFKA_PARTITIONS must be > 0, got %d", n)
	}
	return n, nil
}
