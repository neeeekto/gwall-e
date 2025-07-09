package events

import . "github.com/gwall-e/pkg/core_entities"

type ProjectAddedEvent struct {
	EventBase
	ID   string   `bson:"id"`
	Name string   `bson:"name"`
	Type UnitType `bson:"type"`
}

// nil - no updates for this field
type ProjectInfoChangedEvent struct {
	EventBase
	ID          string    `bson:"id"`
	Name        *string   `bson:"name"`
	Description *string   `bson:"description"`
	Tags        *[]string `bson:"tags"`
	Owners      *[]string `bson:"tags"`
}
