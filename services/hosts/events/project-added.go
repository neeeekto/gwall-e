package events

import "github.com/gwall-e/pkg/core_entities"

type ProjectAddedEvent struct {
	ID   string                 `bson:"id"`
	Name string                 `bson:"name"`
	Type core_entities.UnitType `bson:"type"`
}
