// Direct unit specs for the immutable HostHardware VO: structured sub-components,
// raw-string external IDs (HW-06), defensive-copy immutability (Pitfall 2), and
// constructor validation. Black-box (package domain_test) — VO is exercised through its
// public constructor and getters. No mocks (D-03). Comments are English (style.md).
package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// validSpec builds a fully-populated HardwareSpec covering every sub-component, so specs
// can assert structure and immutability against a realistic VO.
func validSpec() domain.HardwareSpec {
	return domain.HardwareSpec{
		Name:        "compute-node-a",
		Platform:    "gen5",
		Motherboard: domain.Motherboard{Model: "X12", Vendor: "Supermicro", Serial: "MB-001", Inv: "INV-MB-1"},
		IPMIMac:     "aa:bb:cc:dd:ee:ff",
		RAM:         []domain.RAMModule{{Slot: "A1", Model: "DDR4", Vendor: "Hynix", Lot: "L1", Serial: "R-1", Inv: "INV-R-1", SizeGB: 32, SpeedMHz: 3200}},
		CPU:         []domain.CPU{{Slot: "P1", Model: "Xeon", Vendor: "Intel", Lot: "L2", Serial: "C-1", Inv: "INV-C-1", Cores: 24}},
		Drives:      []domain.Drive{{Slot: "D1", Model: "NVMe-2T", Vendor: "Samsung", Lot: "L3", Serial: "D-1", Inv: "INV-D-1", SizeGB: 2000, Type: "NVMe"}},
		NICs:        []domain.NIC{{Model: "X710", SpeedGbE: 10, MACs: []string{"11:22:33:44:55:66"}}},
		PSUs:        []domain.PSU{{Model: "PWS-1", Vendor: "Supermicro", Serial: "P-1", Inv: "INV-P-1", PowerW: 800}},
		StorageCtl:  []domain.StorageController{{Model: "MegaRAID", Vendor: "Broadcom", Serial: "S-1", Inv: "INV-S-1", RAIDLvl: "10"}},
		GPUs:        []domain.GPU{{Model: "A100", Vendor: "NVIDIA", Serial: "G-1", Inv: "INV-G-1", MemoryGB: 80}},
		Chassis:     domain.Chassis{Model: "CSE-829", Vendor: "Supermicro", Serial: "CH-1", Inv: "INV-CH-1", UnitsU: 2},
	}
}

var _ = Describe("HostHardware VO", func() {
	Context("when built from a valid spec", func() {
		It("exposes every structured sub-component (HW-01..05)", func() {
			hw, err := domain.NewHostHardware(validSpec())
			Expect(err).ToNot(HaveOccurred())

			Expect(hw.Name()).To(Equal("compute-node-a"))
			Expect(hw.Platform()).To(Equal("gen5"))
			Expect(hw.IPMIMac()).To(Equal("aa:bb:cc:dd:ee:ff"))
			Expect(hw.Motherboard().Model).To(Equal("X12"))
			Expect(hw.Chassis().Model).To(Equal("CSE-829"))
			Expect(hw.RAM()).To(HaveLen(1))
			Expect(hw.CPU()).To(HaveLen(1))
			Expect(hw.Drives()).To(HaveLen(1))
			Expect(hw.NICs()).To(HaveLen(1))
			Expect(hw.PSUs()).To(HaveLen(1))
			Expect(hw.StorageControllers()).To(HaveLen(1))
			Expect(hw.GPUs()).To(HaveLen(1))
		})

		It("models NIC as a structured component, not a flat host-level MACs[] (HW-03)", func() {
			hw, _ := domain.NewHostHardware(validSpec())
			nics := hw.NICs()
			Expect(nics[0].Model).To(Equal("X710"))
			Expect(nics[0].SpeedGbE).To(Equal(10))
			Expect(nics[0].MACs).To(ConsistOf("11:22:33:44:55:66"))
		})

		It("stores all external identifiers as raw strings (HW-06)", func() {
			hw, _ := domain.NewHostHardware(validSpec())
			// Serials / inventory numbers / MACs are plain strings — domain does not parse them.
			Expect(hw.Motherboard().Serial).To(Equal("MB-001"))
			Expect(hw.Drives()[0].Inv).To(Equal("INV-D-1"))
			Expect(hw.NICs()[0].MACs[0]).To(Equal("11:22:33:44:55:66"))
		})
	})

	Context("when a caller mutates a slice returned from the VO (Pitfall 2)", func() {
		It("does not leak the mutation back into the immutable VO", func() {
			hw, _ := domain.NewHostHardware(validSpec())

			// Mutate the returned NIC slice and its nested MACs.
			nics := hw.NICs()
			nics[0].Model = "TAMPERED"
			nics[0].MACs[0] = "00:00:00:00:00:00"

			// Mutate another returned slice.
			cpus := hw.CPU()
			cpus[0].Model = "TAMPERED"

			fresh := hw.NICs()
			Expect(fresh[0].Model).To(Equal("X710"), "NIC model must be unchanged (defensive copy)")
			Expect(fresh[0].MACs[0]).To(Equal("11:22:33:44:55:66"), "nested MACs must be unchanged (deep copy)")
			Expect(hw.CPU()[0].Model).To(Equal("Xeon"), "CPU model must be unchanged (defensive copy)")
		})
	})

	Context("when the input spec slice is mutated after construction (Pitfall 2)", func() {
		It("does not affect the already-built VO", func() {
			spec := validSpec()
			hw, _ := domain.NewHostHardware(spec)

			spec.NICs[0].MACs[0] = "ff:ff:ff:ff:ff:ff"
			spec.CPU[0].Model = "TAMPERED"

			Expect(hw.NICs()[0].MACs[0]).To(Equal("11:22:33:44:55:66"), "constructor must defensively copy input slices")
			Expect(hw.CPU()[0].Model).To(Equal("Xeon"))
		})
	})

	Context("when a required field is empty", func() {
		It("rejects an empty name with a domain error (V5 input validation)", func() {
			spec := validSpec()
			spec.Name = ""
			_, err := domain.NewHostHardware(spec)
			Expect(err).To(MatchError(domain.ErrInvalidHardware))
		})
	})
})
