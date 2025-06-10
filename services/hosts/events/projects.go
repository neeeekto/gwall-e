package events

import "github.com/gwall-e/pkg/core_entities"

type ProjectAddedEvent struct {
	ID         string                 `bson:"id"`
	Name       string                 `bson:"name"`
	Type       core_entities.UnitType `bson:"type"`
	OccurredOn int64                  `bson:"occurred_on"`
}

func (e *ProjectAddedEvent) SetOccurredOn(t int64) {
	e.OccurredOn = t
}
