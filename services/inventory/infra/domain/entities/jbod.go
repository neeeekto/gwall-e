package entities

type JBodHardware struct {
	HDD []HardwareDrive
	SSD []HardwareDrive
}

type JBodID string

type JBod struct {
	ID       JBodID
	Inv      int
	Location Location
	Hardware JBodHardware
}
