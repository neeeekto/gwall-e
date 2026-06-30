// Direct unit specs for the Project aggregate: factory invariants (non-empty name, INV-01),
// Owner kept as a raw opaque string (INV-09), and exactly one semantic event per operation
// (D-13/EVT-01). Black-box (package domain_test) — only the exported surface is exercised.
// No mocks (D-03). Comments are English (style.md).
package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	domain "github.com/gwall-e/services/inventory/internal/domain"
)

var _ = Describe("Project aggregate", func() {
	Describe("NewProject factory", func() {
		Context("with a valid name", func() {
			It("creates a project, stores the owner raw, and emits one ProjectCreated", func() {
				p, err := domain.NewProject("infra", "core platform", "team-platform")

				Expect(err).ToNot(HaveOccurred())
				Expect(p.ID().IsZero()).To(BeFalse(), "factory must generate a non-zero ProjectID (INV-03)")
				Expect(p.Name()).To(Equal("infra"))
				Expect(p.Description()).To(Equal("core platform"))
				Expect(p.Owner()).To(Equal("team-platform"), "owner is kept as a raw opaque string (INV-09)")
				Expect(p.Version()).To(Equal(1), "exactly one operation recorded")

				events := p.PullEvents()
				Expect(events).To(HaveLen(1), "exactly one semantic event (EVT-01)")
				Expect(events[0].EventType()).To(Equal("ProjectCreated"))
				Expect(events[0].EntityID()).To(Equal(p.ID().String()), "EntityID is the ProjectID (Kafka key)")
			})

			It("allows an empty owner — ownership can be assigned later", func() {
				p, err := domain.NewProject("infra", "", "")

				Expect(err).ToNot(HaveOccurred())
				Expect(p.Owner()).To(BeEmpty())
			})
		})

		Context("with an empty name", func() {
			It("returns ErrInvalidProject and no aggregate (V5/INV-01)", func() {
				p, err := domain.NewProject("", "desc", "owner")

				Expect(p).To(BeNil())
				Expect(err).To(MatchError(domain.ErrInvalidProject))
			})
		})
	})

	Describe("operations emit exactly one semantic event each (D-13)", func() {
		var p *domain.Project

		BeforeEach(func() {
			var err error
			p, err = domain.NewProject("infra", "core", "team-a")
			Expect(err).ToNot(HaveOccurred())
			// Drain the creation event so each operation spec starts from a clean buffer.
			Expect(p.PullEvents()).To(HaveLen(1))
		})

		It("Rename changes the name and records ProjectRenamed", func() {
			err := p.Rename("platform")

			Expect(err).ToNot(HaveOccurred())
			Expect(p.Name()).To(Equal("platform"))

			events := p.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ProjectRenamed"))
		})

		It("Rename rejects an empty name and records nothing", func() {
			err := p.Rename("")

			Expect(err).To(MatchError(domain.ErrInvalidProject))
			Expect(p.Name()).To(Equal("infra"), "name is unchanged on rejection")
			Expect(p.PullEvents()).To(BeEmpty(), "no event on a rejected operation")
		})

		It("ChangeOwner updates the owner and records ProjectOwnerChanged", func() {
			err := p.ChangeOwner("team-b")

			Expect(err).ToNot(HaveOccurred())
			Expect(p.Owner()).To(Equal("team-b"))

			events := p.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ProjectOwnerChanged"))
		})

		It("Delete records ProjectDeleted", func() {
			p.Delete()

			events := p.PullEvents()
			Expect(events).To(HaveLen(1))
			Expect(events[0].EventType()).To(Equal("ProjectDeleted"))
		})
	})

	Describe("version tracking across N operations", func() {
		It("has Version()==N and PullEvents()==N after N operations", func() {
			p, err := domain.NewProject("infra", "core", "team-a") // op 1
			Expect(err).ToNot(HaveOccurred())
			Expect(p.Rename("platform")).To(Succeed())    // op 2
			Expect(p.ChangeOwner("team-b")).To(Succeed()) // op 3
			p.Delete()                                    // op 4

			Expect(p.Version()).To(Equal(4))
			Expect(p.PullEvents()).To(HaveLen(4))
		})
	})
})
