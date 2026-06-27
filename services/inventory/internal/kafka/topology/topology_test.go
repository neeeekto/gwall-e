package topology

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTopologySuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kafka Topology Suite")
}

var _ = Describe("topic name derivation", func() {
	It("derives the events topic for the host aggregate", func() {
		Expect(eventsTopic("host")).To(Equal("inventory.host.events"))
	})

	It("derives the state topic for the host aggregate", func() {
		Expect(stateTopic("host")).To(Equal("inventory.host.state"))
	})
})

var _ = Describe("cleanup policy configuration", func() {
	It("configures *.events with delete and never compact", func() {
		cfg := eventsConfig()
		Expect(cfg).To(HaveKey("cleanup.policy"))
		Expect(*cfg["cleanup.policy"]).To(Equal("delete"))
		Expect(*cfg["cleanup.policy"]).ToNot(ContainSubstring("compact"))
	})

	It("configures *.state with compact and a 24h tombstone retention", func() {
		cfg := stateConfig()
		Expect(cfg).To(HaveKey("cleanup.policy"))
		Expect(*cfg["cleanup.policy"]).To(Equal("compact"))
		Expect(cfg).To(HaveKey("delete.retention.ms"))
		// 86400000 ms == 24h, satisfying the >= 24h invariant (D-12).
		Expect(*cfg["delete.retention.ms"]).To(Equal("86400000"))
	})
})

var _ = Describe("aggregate list", func() {
	It("contains exactly the host aggregate in Phase 5", func() {
		Expect(aggregates).To(Equal([]string{"host"}))
	})
})
