// White-box specs for aggregateBase: record() increments version in a single point
// and PullEvents() drains the buffer (Pitfall 3/5). White-box (package domain) because
// aggregateBase, record() and its fields are unexported by design — events are born only
// inside domain transitions. No mocks (D-03). Comments are English (style.md).
package domain

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeEvent is a minimal DomainEvent used to exercise the base without a real aggregate.
type fakeEvent struct {
	typ string
	eid string
}

func (e fakeEvent) EventType() string { return e.typ }
func (e fakeEvent) EntityID() string  { return e.eid }

// testAggregate embeds aggregateBase so the spec can drive record() the way a real
// aggregate would.
type testAggregate struct {
	aggregateBase
}

var _ = Describe("aggregateBase", func() {
	var agg *testAggregate

	BeforeEach(func() {
		agg = &testAggregate{}
	})

	Context("when recording several events", func() {
		It("increments version once per record and buffers every event", func() {
			agg.record(fakeEvent{typ: "A", eid: "1"})
			agg.record(fakeEvent{typ: "B", eid: "1"})
			agg.record(fakeEvent{typ: "C", eid: "1"})

			Expect(agg.Version()).To(Equal(3), "version must equal the record count (Pitfall 3)")

			pulled := agg.PullEvents()
			Expect(pulled).To(HaveLen(3), "every recorded event must be pulled")
		})
	})

	Context("when pulling events twice in a row", func() {
		It("returns the buffer once and empties it (Pitfall 5)", func() {
			agg.record(fakeEvent{typ: "A", eid: "1"})

			first := agg.PullEvents()
			Expect(first).To(HaveLen(1))

			second := agg.PullEvents()
			Expect(second).To(BeEmpty(), "second pull must be empty — events are not duplicated")
		})
	})

	Context("on a fresh aggregate", func() {
		It("has zero version and no events", func() {
			Expect(agg.Version()).To(Equal(0))
			Expect(agg.PullEvents()).To(BeEmpty())
		})
	})
})
