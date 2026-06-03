package entities

type Project struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Kind        HostKind
}
