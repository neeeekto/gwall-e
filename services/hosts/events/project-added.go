package events

type ProjectAddedEvent struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`
	Type string `bson:"type"`
}
