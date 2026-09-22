package domain

import (
	"testing"
	"time"
)

// london is the zone every fixture reads its dates in. It is loaded from the tz
// data embedded in the test binary rather than from the machine.
func london(t *testing.T) *time.Location {
	t.Helper()
	zone, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("loading Europe/London: %v", err)
	}
	return zone
}

// at answers the instant of a London wall-clock time.
func at(t *testing.T, year int, month time.Month, day, hour, minute int) time.Time {
	t.Helper()
	return time.Date(year, month, day, hour, minute, 0, 0, london(t))
}

// event builds an event recorded at the moment it occurred.
func event(id EventID, definition DefinitionID, symptom string, when time.Time) Event {
	return Event{
		ID: id, Definition: definition, Symptom: symptom,
		OccurredAt: when, RecordedAt: when,
	}
}
