package broadcast

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMemoryEventLog_NewEvent(t *testing.T) {
	log := NewMemoryEventLog(10)
	newEvent := log.NewEvent("First")

	if newEvent.Message != "First" {
		t.Errorf("New Event should have specified message :\n got: %v\nwant: %v", newEvent.Message, "Test")
	}
	if time.Now().Sub(newEvent.Timestamp) > time.Second {
		t.Errorf("New Event timestamp should be current time :\n got: %v\nwant: %v", newEvent.Timestamp, time.Now())
	}
	if newEvent.Occurrence != 1 {
		t.Errorf("New Event Occurrence should be one :\n got: %v", newEvent.Occurrence)
	}
	if newEvent.Source != "" {
		t.Errorf("New Event Source should be empty :\n got: %v", newEvent.Source)
	}

	if newEvent != log.Events()[0] {
		t.Errorf("New Event should be stored in Log :\n got: %v", log.Events())
	}

	for i := 1; i <= log.size; i++ {
		log.NewEvent(fmt.Sprintf("Next %d", i))
	}
	if log.Events()[0].Message != "Next 1" {
		t.Errorf("First Event should be lost when log is full :\n got: %v\nwant: %v", log.Events()[0].Message, "Next 1")
	}
}

func TestLocalEventLog_NewEvent(t *testing.T) {
	parent := NewMemoryEventLog(10)
	localEventLog := LocalEventLog{
		Parent: parent,
		Source: "test",
	}
	newEvent := localEventLog.NewEvent("dummy")
	if newEvent.Source != localEventLog.Source {
		t.Errorf("New Event Source should be LocalEventLog source :\n got: %v\nwant: %v", newEvent.Source, localEventLog.Source)
	}

	if newEvent != parent.Events()[0] {
		t.Errorf("New Event should be stored in parent Log :\n got: %v", parent.Events)
	}
}

func TestLocalEventLog_Events(t *testing.T) {
	parent := NewMemoryEventLog(10)
	parent.NewEvent("Other")

	localEventLog := LocalEventLog{
		Parent: parent,
		Source: "test",
	}
	newEvent := localEventLog.NewEvent("dummy")

	if !reflect.DeepEqual(localEventLog.Events(), []*Event{newEvent}) {
		t.Errorf("Events() should only return local events :\n got: %v", localEventLog.Events())
	}
}
