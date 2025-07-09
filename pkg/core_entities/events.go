package core_entities

import "time"

type Event interface {
	SetOccurredOn(time int64)
}

type EventBase struct {
	OccurredOn int64 `bson:"occurred_on"`
	
}

func (e *EventBase) SetOccurredOn(t int64) {
	e.OccurredOn = t
}

type Events struct {
	events []Event `bson:"_"`
}

func (p *Events) List() []Event {
	return p.events
}

func (p *Events) Add(event Event) {
	event.SetOccurredOn(time.Now().Unix())
	p.events = append(p.events, event)

}
