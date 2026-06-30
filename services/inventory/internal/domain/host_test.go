// White-box specs for the Host aggregate: factory invariants, the three-member lifecycle
// state-machine (terminal decommission), and one semantic event per operation. White-box
// (package domain) because lifecycleState and its members (stateShadow/...) are unexported
// — the transition matrix must drive them directly. No mocks (D-03). Comments are English
// (style.md).
package domain

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newValidHost builds a registered host with minimal valid hardware for operation specs.
func newValidHost() *Host {
	hw, _ := NewHostHardware(HardwareSpec{Name: "cfg-a"})
	h, err := NewHost(NewProjectID(), "host-a.dc.example", hw, NewRackID(), "U10", stateRegistered)
	Expect(err).ToNot(HaveOccurred())
	// Drop the HostRegistered emitted by the factory so per-operation specs start clean.
	h.PullEvents()
	return h
}

var _ = Describe("Host aggregate factory (NewHost)", func() {
	var (
		hw HostHardware
	)

	BeforeEach(func() {
		hw, _ = NewHostHardware(HardwareSpec{Name: "cfg-a"})
	})

	Context("when the project binding is missing (INV-02)", func() {
		It("rejects a zero-value ProjectID", func() {
			_, err := NewHost(ProjectID{}, "h.example", hw, NewRackID(), "U1", stateRegistered)
			Expect(err).To(MatchError(ErrMissingProject))
		})
	})

	Context("when the initial state is not a valid start (D-10)", func() {
		It("rejects starting directly as decommissioned", func() {
			_, err := NewHost(NewProjectID(), "h.example", hw, NewRackID(), "U1", stateDecommissioned)
			Expect(err).To(MatchError(ErrInvalidTransition))
		})
	})

	Context("when the input is valid", func() {
		It("generates a non-zero HostID (INV-03) and carries rack+position (LOC-03)", func() {
			h, err := NewHost(NewProjectID(), "h.example", hw, NewRackID(), "U7", stateShadow)
			Expect(err).ToNot(HaveOccurred())
			Expect(h.ID().IsZero()).To(BeFalse())
			Expect(h.Position()).To(Equal("U7"))
		})

		It("emits exactly one HostRegistered (EVT-01)", func() {
			h, _ := NewHost(NewProjectID(), "h.example", hw, NewRackID(), "U7", stateRegistered)
			events := h.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0]).To(BeAssignableToTypeOf(HostRegistered{}))
			Expect(h.Version()).To(Equal(1))
		})
	})
})

var _ = Describe("Host lifecycle state-machine", func() {
	// Decommission transition matrix: valid from shadow/registered, terminal afterwards.
	DescribeTable("Decommission outcome by starting state",
		func(start lifecycleState, wantErr bool) {
			hw, _ := NewHostHardware(HardwareSpec{Name: "cfg"})
			h, err := NewHost(NewProjectID(), "h.example", hw, NewRackID(), "U1", start)
			Expect(err).ToNot(HaveOccurred())

			err = h.Decommission("end of life")
			if wantErr {
				Expect(err).To(MatchError(ErrAlreadyDecommissioned))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("shadow -> decommissioned (ok)", stateShadow, false),
		Entry("registered -> decommissioned (ok)", stateRegistered, false),
	)

	Context("when decommissioning twice (terminal, D-10)", func() {
		It("rejects the second call with ErrAlreadyDecommissioned", func() {
			h := newValidHost()
			Expect(h.Decommission("first")).ToNot(HaveOccurred())
			Expect(h.Decommission("second")).To(MatchError(ErrAlreadyDecommissioned))
		})

		It("emits exactly one HostDecommissioned on the first call", func() {
			h := newValidHost()
			Expect(h.Decommission("eol")).ToNot(HaveOccurred())
			events := h.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0]).To(BeAssignableToTypeOf(HostDecommissioned{}))
		})
	})

	Context("when operating on a decommissioned host", func() {
		It("rejects Reassign/Relocate/ChangeHardware (no resurrection)", func() {
			h := newValidHost()
			Expect(h.Decommission("eol")).ToNot(HaveOccurred())

			Expect(h.Reassign(NewProjectID())).To(MatchError(ErrAlreadyDecommissioned))
			Expect(h.Relocate(NewRackID(), "U2")).To(MatchError(ErrAlreadyDecommissioned))
			newHW, _ := NewHostHardware(HardwareSpec{Name: "cfg-b"})
			Expect(h.ChangeHardware(newHW)).To(MatchError(ErrAlreadyDecommissioned))
		})
	})
})

var _ = Describe("Host operations emit one semantic event each (D-13)", func() {
	It("Reassign emits exactly one HostReassigned and changes ProjectID (INV-05)", func() {
		h := newValidHost()
		newPID := NewProjectID()
		Expect(h.Reassign(newPID)).ToNot(HaveOccurred())
		Expect(h.ProjectID()).To(Equal(newPID))

		events := h.PullEvents()
		Expect(events).To(HaveLen(1))
		Expect(events[0]).To(BeAssignableToTypeOf(HostReassigned{}))
	})

	It("Reassign rejects a zero-value ProjectID (INV-02)", func() {
		h := newValidHost()
		Expect(h.Reassign(ProjectID{})).To(MatchError(ErrMissingProject))
	})

	It("Relocate emits exactly one HostRelocated", func() {
		h := newValidHost()
		Expect(h.Relocate(NewRackID(), "U42")).ToNot(HaveOccurred())
		Expect(h.Position()).To(Equal("U42"))

		events := h.PullEvents()
		Expect(events).To(HaveLen(1))
		Expect(events[0]).To(BeAssignableToTypeOf(HostRelocated{}))
	})

	It("ChangeHardware replaces the VO whole and emits one HostHardwareChanged (D-07)", func() {
		h := newValidHost()
		newHW, _ := NewHostHardware(HardwareSpec{Name: "cfg-b"})
		Expect(h.ChangeHardware(newHW)).ToNot(HaveOccurred())
		Expect(h.Hardware().Name()).To(Equal("cfg-b"))

		events := h.PullEvents()
		Expect(events).To(HaveLen(1))
		Expect(events[0]).To(BeAssignableToTypeOf(HostHardwareChanged{}))
	})
})

var _ = Describe("Host Delete (hard-delete, Pitfall 1/D-09)", func() {
	It("emits HostDeleted from any state without setting a deleted lifecycle-state", func() {
		h := newValidHost()
		before := h.state
		h.Delete()

		Expect(h.state).To(Equal(before), "Delete must NOT mutate lifecycleState (deleted is not an enum member)")
		events := h.PullEvents()
		Expect(events).To(HaveLen(1))
		Expect(events[0]).To(BeAssignableToTypeOf(HostDeleted{}))
	})

	It("works even on a decommissioned host", func() {
		h := newValidHost()
		Expect(h.Decommission("eol")).ToNot(HaveOccurred())
		h.PullEvents()

		h.Delete()
		Expect(h.state).To(Equal(stateDecommissioned), "state stays decommissioned; deleted is not a state")
		Expect(h.PullEvents()).To(HaveLen(1))
	})
})

var _ = Describe("Host version tracking (Pitfall 3)", func() {
	It("has Version() equal to the number of operations and matching event count", func() {
		hw, _ := NewHostHardware(HardwareSpec{Name: "cfg"})
		h, _ := NewHost(NewProjectID(), "h.example", hw, NewRackID(), "U1", stateRegistered) // op 1

		Expect(h.Reassign(NewProjectID())).ToNot(HaveOccurred())    // op 2
		Expect(h.Relocate(NewRackID(), "U2")).ToNot(HaveOccurred()) // op 3
		Expect(h.Decommission("eol")).ToNot(HaveOccurred())         // op 4

		Expect(h.Version()).To(Equal(4))
		Expect(h.PullEvents()).To(HaveLen(4), "every recorded operation must be pulled (Version == event count)")
	})
})
