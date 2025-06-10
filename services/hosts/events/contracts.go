package events

type Event interface {
	SetOccurredOn(time int64)
}
