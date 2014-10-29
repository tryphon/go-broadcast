package broadcast

import (
	"time"
)

type Event struct {
	Timestamp  time.Time
	Message    string
	Source     string `json:",omitempty"`
	Occurrence int
}

func (event *Event) Equal(other *Event) bool {
	return other != nil &&
		event.Message == other.Message &&
		event.Source == other.Source
	// && math.Abs(float64(event.Timestamp.Unix()-other.Timestamp.Unix())) <= 10
}

type EventLogger interface {
	NewEvent(message string) *Event
}

type EventCollection interface {
	EventLogger

	Create(message string) *Event
	Append(event *Event)
	Events() []*Event
}

var EventLog EventCollection = &LoggerEventLog{Parent: NewMemoryEventLog(256)}

type MemoryEventLog struct {
	events []*Event
	size   int
}

func NewMemoryEventLog(size int) *MemoryEventLog {
	return &MemoryEventLog{
		events: make([]*Event, 0),
		size:   size,
	}
}

func (log *MemoryEventLog) Create(message string) *Event {
	return &Event{Timestamp: time.Now(), Message: message, Occurrence: 1}
}

func (log *MemoryEventLog) Append(event *Event) {
	tail := log.events
	if len(log.events) >= log.size {
		tail = log.events[1:]
	}

	log.events = append(tail, event)
}

func (log *MemoryEventLog) NewEvent(message string) *Event {
	event := log.Create(message)
	log.Append(event)
	return event
}

func (log *MemoryEventLog) Events() []*Event {
	return log.events
}

type LocalEventLog struct {
	Parent    EventCollection
	Source    string
	lastEvent *Event
}

func (log *LocalEventLog) parent() EventCollection {
	if log.Parent == nil {
		log.Parent = EventLog
	}
	return log.Parent
}

func (log *LocalEventLog) NewEvent(message string) *Event {
	event := log.Create(message)
	log.Append(event)
	return event
}

func (log *LocalEventLog) Create(message string) *Event {
	event := log.parent().Create(message)
	event.Source = log.Source
	return event
}

func (log *LocalEventLog) Append(event *Event) {
	if event.Source == log.Source {
		if log.lastEvent != nil && event.Equal(log.lastEvent) {
			log.lastEvent.Occurrence += 1
		} else {
			log.parent().Append(event)
			log.lastEvent = event
		}
	}
}

func (log *LocalEventLog) Events() []*Event {
	events := make([]*Event, 0)
	for _, event := range log.parent().Events() {
		if event.Source == log.Source {
			events = append(events, event)
		}
	}
	return events
}

type LoggerEventLog struct {
	Parent EventCollection
}

func (log *LoggerEventLog) NewEvent(message string) *Event {
	return log.Parent.NewEvent(message)
}

func (log *LoggerEventLog) Create(message string) *Event {
	return log.Parent.Create(message)
}

func (log *LoggerEventLog) Append(event *Event) {
	if event.Source != "" {
		Log.Printf("%s > %s", event.Source, event.Message)
	} else {
		Log.Printf(event.Message)
	}

	log.Parent.Append(event)
}

func (log *LoggerEventLog) Events() []*Event {
	return log.Parent.Events()
}
