//go:build integration

// Integration smoke test for the dev-infra stack: ephemeral Kafka (KRaft) and Mongo
// single-node replica set spun up via testcontainers. Isolated behind the `integration`
// build tag (D-15) so `go test ./...` and pre-push compile/run unit tests only, without
// Docker. Run with `make test-integration` (= `go test -tags=integration ./...`).
//
// Both the CLI and this test call the SAME topology.Bootstrap and mongoconn.Connect
// (single source of truth, D-06): the test never redefines topic names or cleanup policies.
package topology_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/gwall-e/pkg/mongoconn"
	"github.com/gwall-e/services/inventory/internal/kafka/topology"
)

func TestTopologyIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Topology Integration Suite")
}

var _ = Describe("dev infra smoke", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("provisions inventory topics on ephemeral Kafka (KRaft)", func() {
		// ephemeral single-broker KRaft cluster; image tag pinned for parity with compose (D-07).
		kc, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0",
			kafka.WithClusterID("test-cluster"))
		DeferCleanup(func() { _ = testcontainers.TerminateContainer(kc) })
		Expect(err).ToNot(HaveOccurred())

		// trust the container's advertised brokers — never hardcode host (Pitfall 1).
		brokers, err := kc.Brokers(ctx)
		Expect(err).ToNot(HaveOccurred())

		cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(cl.Close)

		// smoke connect to the broker.
		Expect(cl.Ping(ctx)).To(Succeed())

		adm := kadm.NewClient(cl)

		// call the single-source Bootstrap (D-06) with the dev/test partition default (D-11).
		const partitions = 6
		Expect(topology.Bootstrap(ctx, adm, partitions)).To(Succeed())

		// assert both topics were provisioned with the expected partition count (SC3).
		const (
			eventsTopic = "inventory.host.events"
			stateTopic  = "inventory.host.state"
		)
		details, err := adm.ListTopics(ctx, eventsTopic, stateTopic)
		Expect(err).ToNot(HaveOccurred())
		Expect(details.Has(eventsTopic)).To(BeTrue())
		Expect(details.Has(stateTopic)).To(BeTrue())
		Expect(details[eventsTopic].Partitions).To(HaveLen(partitions))
		Expect(details[stateTopic].Partitions).To(HaveLen(partitions))

		// assert cleanup policies match the topology contract (D-12): events=delete, state=compact (SC3).
		configs, err := adm.DescribeTopicConfigs(ctx, eventsTopic, stateTopic)
		Expect(err).ToNot(HaveOccurred())

		Expect(cleanupPolicy(configs, eventsTopic)).To(Equal("delete"))
		Expect(cleanupPolicy(configs, stateTopic)).To(Equal("compact"))
	})

	It("connects to ephemeral Mongo single-node replica set", func() {
		// single-node RS; the module configures rs.initiate and ConnectionString returns
		// a correct URI — trust it, do not hardcode the host (Pitfall 1).
		mc, err := mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))
		DeferCleanup(func() { _ = testcontainers.TerminateContainer(mc) })
		Expect(err).ToNot(HaveOccurred())

		uri, err := mc.ConnectionString(ctx)
		Expect(err).ToNot(HaveOccurred())

		// the same connection-helper used by Phase 6/7 repositories (D-14).
		client, err := mongoconn.Connect(ctx, uri)
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() { _ = client.Disconnect(ctx) })
	})
})

// cleanupPolicy extracts the cleanup.policy config value for the named topic.
func cleanupPolicy(configs kadm.ResourceConfigs, topic string) string {
	rc, err := configs.On(topic, nil)
	Expect(err).ToNot(HaveOccurred())
	for _, c := range rc.Configs {
		if c.Key == "cleanup.policy" {
			return c.MaybeValue()
		}
	}
	return ""
}
