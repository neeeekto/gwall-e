// Direct unit specs for the three independent location aggregates DC/Module/Rack: factory
// invariants (non-empty name; non-zero parent ID for Module/Rack — LOC-02/D-06), hierarchy
// by internal ID (Module holds DCID, Rack holds ModuleID), Rack topology attributes (LOC-04),
// and exactly one semantic event per CRUD operation (D-13/EVT-01). Black-box, no mocks (D-03).
// Comments are English (style.md).
package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	domain "github.com/gwall-e/services/inventory/internal/domain"
)

var _ = Describe("DC aggregate", func() {
	Describe("NewDC factory", func() {
		It("creates a DC and emits one DCCreated", func() {
			dc, err := domain.NewDC("dc-msk-1", "Moscow")

			Expect(err).ToNot(HaveOccurred())
			Expect(dc.ID().IsZero()).To(BeFalse(), "factory generates a non-zero DCID (INV-03)")
			Expect(dc.Name()).To(Equal("dc-msk-1"))
			Expect(dc.Location()).To(Equal("Moscow"))

			events := dc.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("DCCreated"))
			Expect(events[0].EntityID()).To(Equal(dc.ID().String()))
		})

		It("rejects an empty name with ErrInvalidLocation (LOC-01)", func() {
			dc, err := domain.NewDC("", "Moscow")

			Expect(dc).To(BeNil())
			Expect(err).To(MatchError(domain.ErrInvalidLocation))
		})
	})

	Describe("CRUD operations emit exactly one event each (D-13)", func() {
		var dc *domain.DC

		BeforeEach(func() {
			var err error
			dc, err = domain.NewDC("dc-msk-1", "Moscow")
			Expect(err).ToNot(HaveOccurred())
			Expect(dc.PullEvents()).To(HaveLen(1))
		})

		It("Update records DCUpdated", func() {
			Expect(dc.Update("dc-msk-2", "Moscow-2")).To(Succeed())
			Expect(dc.Name()).To(Equal("dc-msk-2"))

			events := dc.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("DCUpdated"))
		})

		It("Update rejects an empty name and records nothing", func() {
			Expect(dc.Update("", "x")).To(MatchError(domain.ErrInvalidLocation))
			Expect(dc.PullEvents()).To(BeEmpty())
		})

		It("Delete records DCDeleted", func() {
			dc.Delete()

			events := dc.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("DCDeleted"))
		})
	})
})

var _ = Describe("Module aggregate", func() {
	Describe("NewModule factory", func() {
		It("carries the parent DCID and emits one ModuleCreated", func() {
			dcID := domain.NewDCID()

			m, err := domain.NewModule(dcID, "hall-A")

			Expect(err).ToNot(HaveOccurred())
			Expect(m.ID().IsZero()).To(BeFalse())
			Expect(m.DCID()).To(Equal(dcID), "Module references its DC by internal ID (LOC-02/D-06)")
			Expect(m.Name()).To(Equal("hall-A"))

			events := m.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ModuleCreated"))
		})

		It("rejects a zero DCID with ErrInvalidLocation (no dangling module — LOC-02/T-06-09)", func() {
			m, err := domain.NewModule(domain.DCID{}, "hall-A")

			Expect(m).To(BeNil())
			Expect(err).To(MatchError(domain.ErrInvalidLocation))
		})

		It("rejects an empty name", func() {
			m, err := domain.NewModule(domain.NewDCID(), "")

			Expect(m).To(BeNil())
			Expect(err).To(MatchError(domain.ErrInvalidLocation))
		})
	})

	Describe("CRUD operations emit exactly one event each (D-13)", func() {
		var m *domain.Module

		BeforeEach(func() {
			var err error
			m, err = domain.NewModule(domain.NewDCID(), "hall-A")
			Expect(err).ToNot(HaveOccurred())
			Expect(m.PullEvents()).To(HaveLen(1))
		})

		It("Update records ModuleUpdated", func() {
			Expect(m.Update("hall-B")).To(Succeed())
			Expect(m.Name()).To(Equal("hall-B"))

			events := m.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ModuleUpdated"))
		})

		It("Delete records ModuleDeleted", func() {
			m.Delete()

			events := m.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ModuleDeleted"))
		})
	})
})

var _ = Describe("Rack aggregate", func() {
	Describe("NewRack factory", func() {
		It("carries the parent ModuleID and topology attributes, emits one RackCreated", func() {
			moduleID := domain.NewModuleID()
			power := domain.PowerTopology{PowerSource: "feed-1", Generator: "gen-A"}

			r, err := domain.NewRack(moduleID, "rack-01", power)

			Expect(err).ToNot(HaveOccurred())
			Expect(r.ID().IsZero()).To(BeFalse())
			Expect(r.ModuleID()).To(Equal(moduleID), "Rack references its Module by internal ID (LOC-02/D-06)")
			Expect(r.Name()).To(Equal("rack-01"))
			Expect(r.Power()).To(Equal(power), "Rack carries power topology attributes (LOC-04)")

			events := r.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("RackCreated"))
		})

		It("rejects a zero ModuleID with ErrInvalidLocation (no dangling rack — LOC-02/T-06-09)", func() {
			r, err := domain.NewRack(domain.ModuleID{}, "rack-01", domain.PowerTopology{})

			Expect(r).To(BeNil())
			Expect(err).To(MatchError(domain.ErrInvalidLocation))
		})

		It("rejects an empty name", func() {
			r, err := domain.NewRack(domain.NewModuleID(), "", domain.PowerTopology{})

			Expect(r).To(BeNil())
			Expect(err).To(MatchError(domain.ErrInvalidLocation))
		})
	})

	Describe("CRUD operations emit exactly one event each (D-13)", func() {
		var r *domain.Rack

		BeforeEach(func() {
			var err error
			r, err = domain.NewRack(domain.NewModuleID(), "rack-01", domain.PowerTopology{PowerSource: "feed-1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(r.PullEvents()).To(HaveLen(1))
		})

		It("Update changes name and topology and records RackUpdated", func() {
			newPower := domain.PowerTopology{PowerSource: "feed-2", Generator: "gen-B"}
			Expect(r.Update("rack-02", newPower)).To(Succeed())
			Expect(r.Name()).To(Equal("rack-02"))
			Expect(r.Power()).To(Equal(newPower))

			events := r.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("RackUpdated"))
		})

		It("Delete records RackDeleted", func() {
			r.Delete()

			events := r.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("RackDeleted"))
		})
	})
})
