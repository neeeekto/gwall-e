// Direct unit specs for the typed ID-VO factories, parsing and zero-value guards
// (INV-03/D-05). Pure functions, no mocks (D-03). Comments are English (style.md).
package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// idType bundles the factory/parse/zero behaviour of one typed ID-VO so the five
// concrete types can be driven through the same table without reflection.
type idCase struct {
	name string
	// new returns a freshly generated ID and whether it reports zero.
	newID func() (str string, isZero bool)
	// parseRoundTrip parses the given canonical string and reports the round-tripped
	// String() value plus any error.
	parseRoundTrip func(s string) (out string, err error)
	// parseBad parses a malformed string and returns the resulting error.
	parseBad func(s string) error
	// zeroValueIsZero reports IsZero() on the type's zero value.
	zeroValueIsZero func() bool
}

var idCases = []idCase{
	{
		name: "HostID",
		newID: func() (string, bool) {
			id := domain.NewHostID()
			return id.String(), id.IsZero()
		},
		parseRoundTrip: func(s string) (string, error) {
			id, err := domain.ParseHostID(s)
			return id.String(), err
		},
		parseBad:        func(s string) error { _, err := domain.ParseHostID(s); return err },
		zeroValueIsZero: func() bool { return domain.HostID{}.IsZero() },
	},
	{
		name: "ProjectID",
		newID: func() (string, bool) {
			id := domain.NewProjectID()
			return id.String(), id.IsZero()
		},
		parseRoundTrip: func(s string) (string, error) {
			id, err := domain.ParseProjectID(s)
			return id.String(), err
		},
		parseBad:        func(s string) error { _, err := domain.ParseProjectID(s); return err },
		zeroValueIsZero: func() bool { return domain.ProjectID{}.IsZero() },
	},
	{
		name: "DCID",
		newID: func() (string, bool) {
			id := domain.NewDCID()
			return id.String(), id.IsZero()
		},
		parseRoundTrip: func(s string) (string, error) {
			id, err := domain.ParseDCID(s)
			return id.String(), err
		},
		parseBad:        func(s string) error { _, err := domain.ParseDCID(s); return err },
		zeroValueIsZero: func() bool { return domain.DCID{}.IsZero() },
	},
	{
		name: "ModuleID",
		newID: func() (string, bool) {
			id := domain.NewModuleID()
			return id.String(), id.IsZero()
		},
		parseRoundTrip: func(s string) (string, error) {
			id, err := domain.ParseModuleID(s)
			return id.String(), err
		},
		parseBad:        func(s string) error { _, err := domain.ParseModuleID(s); return err },
		zeroValueIsZero: func() bool { return domain.ModuleID{}.IsZero() },
	},
	{
		name: "RackID",
		newID: func() (string, bool) {
			id := domain.NewRackID()
			return id.String(), id.IsZero()
		},
		parseRoundTrip: func(s string) (string, error) {
			id, err := domain.ParseRackID(s)
			return id.String(), err
		},
		parseBad:        func(s string) error { _, err := domain.ParseRackID(s); return err },
		zeroValueIsZero: func() bool { return domain.RackID{}.IsZero() },
	},
}

var _ = Describe("Typed ID-VO", func() {
	DescribeTable("factory produces a non-zero, unique identifier",
		func(c idCase) {
			first, firstZero := c.newID()
			second, _ := c.newID()

			Expect(first).ToNot(BeEmpty(), "generated id must have a string form")
			Expect(firstZero).To(BeFalse(), "generated id must not report zero")
			Expect(first).ToNot(Equal(second), "two generations must differ (INV-03)")
		},
		entriesFor(),
	)

	DescribeTable("parse round-trips a canonical string",
		func(c idCase) {
			canonical, _ := c.newID()
			out, err := c.parseRoundTrip(canonical)

			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(canonical), "parse then String() must round-trip")
		},
		entriesFor(),
	)

	DescribeTable("parse of a malformed string wraps ErrInvalidID",
		func(c idCase) {
			err := c.parseBad("not-a-uuid")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(domain.ErrInvalidID), "errors.Is must see the sentinel")
		},
		entriesFor(),
	)

	DescribeTable("zero value reports IsZero() true",
		func(c idCase) {
			Expect(c.zeroValueIsZero()).To(BeTrue())
		},
		entriesFor(),
	)
})

// entriesFor builds the shared table entries from idCases. Centralised so each
// DescribeTable drives all five ID types without copy-paste.
func entriesFor() []TableEntry {
	entries := make([]TableEntry, 0, len(idCases))
	for _, c := range idCases {
		entries = append(entries, Entry(c.name, c))
	}
	return entries
}

// Compile-time type-distinctness note (documented, not executed): the compiler rejects
// cross-type assignment such as `var p domain.ProjectID = domain.NewHostID()` because
// HostID and ProjectID are distinct struct types (D-05/T-06-01). Enabling that line
// would fail `go build`, which is exactly the guarantee under test.
