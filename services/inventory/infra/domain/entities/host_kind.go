package entities

//go:generate go-enum -f=$GOFILE --marshal --bson --sql
/*
Action
ENUM(
	unknown
	server
	mac
)
*/
type HostKind int
