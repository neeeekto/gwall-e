package entities

type JBogHardware struct {
}

type JBogID string

type JBog struct {
	ID       JBogID
	Inc      int
	Location Location
	Hardware JBogHardware
}
