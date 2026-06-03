package domain

type HostKind int

const (
	HostKindUnknown HostKind = iota
	HostKindServer
	HostKindMac
)

func (k HostKind) String() string {
	switch k {
	case HostKindServer:
		return "server"
	case HostKindMac:
		return "mac"
	default:
		return "unknown"
	}
}

func ParseHostKind(s string) (HostKind, error) {
	switch s {
	case "server":
		return HostKindServer, nil
	case "mac":
		return HostKindMac, nil
	default:
		return HostKindUnknown, ErrInvalidHostKind
	}
}
