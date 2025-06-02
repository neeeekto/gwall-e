package common

// UnitType represents the type of infrastructure unit (server, VM, etc.)
type UnitType string

const (
	TypeServer       UnitType = "server"
	TypeVM           UnitType = "vm"
	TypeMac          UnitType = "mac"
	TypeShadowServer UnitType = "shadow-server"
)