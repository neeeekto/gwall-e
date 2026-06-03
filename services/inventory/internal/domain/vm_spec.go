package domain

import "fmt"

// VMSpec — спецификация ресурсов виртуальной машины.
// Value Object: неизменяемый, сравнивается по значению.
type VMSpec struct {
	CPUCores uint32
	RAMmb    uint32
	DiskGB   uint32
	DiskType string // "ssd" | "hdd" | "nvme"
}

// Validate проверяет корректность спецификации.
func (s VMSpec) Validate() error {
	if s.CPUCores == 0 {
		return fmt.Errorf("%w: cpu cores must be > 0", ErrInvalidVMSpec)
	}
	if s.RAMmb == 0 {
		return fmt.Errorf("%w: ram must be > 0", ErrInvalidVMSpec)
	}
	if s.DiskGB == 0 {
		return fmt.Errorf("%w: disk size must be > 0", ErrInvalidVMSpec)
	}
	if s.DiskType == "" {
		return fmt.Errorf("%w: disk type cannot be empty", ErrInvalidVMSpec)
	}
	return nil
}
