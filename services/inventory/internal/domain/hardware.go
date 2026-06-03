package domain

type HostHardware struct {
	Name        string
	Platform    string
	IPMIMac     string
	Motherboard string
	MACs        []string
	RAM         []HardwareRAM
	CPU         []HardwareCPU
	Drives      []HardwareDrive
}

type HardwareRAM struct {
	Slot          string
	Model         string
	Vendor        string
	LotNumber     string
	SerialNumber  string
	RamsizeGB     uint32
	ModuleType    string
	FrequencyMHZ  uint32
	EccCapability string
	Inv           int
}

type HardwareCPU struct {
	Slot          string
	Model         string
	Vendor        string
	LotNumber     string
	SerialNumber  string
	Threads       uint32
	PhysicalCores uint32
	FrequencyMHz  uint32
}

type HardwareDrive struct {
	Slot         string
	Model        string
	Vendor       string
	LotNumber    string
	SerialNumber string
	Interface    string
	CapacityGB   uint32
}
