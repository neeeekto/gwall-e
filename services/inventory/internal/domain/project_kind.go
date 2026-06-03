package domain

// ProjectKind — тип проекта.
// Определяет, какие ресурсы могут быть в проекте.
type ProjectKind int

const (
	ProjectKindBareMetal ProjectKind = iota + 1
	ProjectKindVM
)

func (k ProjectKind) String() string {
	switch k {
	case ProjectKindBareMetal:
		return "baremetal"
	case ProjectKindVM:
		return "vm"
	default:
		return "unknown"
	}
}

func ParseProjectKind(s string) (ProjectKind, error) {
	switch s {
	case "baremetal", "bare-metal", "bare_metal":
		return ProjectKindBareMetal, nil
	case "vm", "virtual_machine":
		return ProjectKindVM, nil
	default:
		return 0, ErrInvalidProjectKind
	}
}
