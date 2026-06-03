package entities

type HostHardware struct {
	Name        string
	Platform    string
	IPMIMac     string
	Motherboard string
	MACs        []string
	RAM         []HardwareRAM
	CPU         []HardwareCPU
	Drivers     []HardwareDrive
}

type HostConnections struct {
	JBoD []JBodID
	JBoG []JBogID
}

type HostID string

type Host struct {
	ID          HostID
	FQDN        string
	Inv         int
	Type        HostKind
	Tags        []string
	Location    Location
	Hardware    HostHardware
	Connections HostConnections
}
