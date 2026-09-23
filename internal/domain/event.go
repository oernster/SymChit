package domain

import (
	"cmp"
	"errors"
	"time"
)

// EventID identifies an event in the store.
type EventID int64

// DefinitionID identifies a symptom definition in the store.
type DefinitionID int64

// Event is one observation: a symptom at a point in time, with the user's
// optional severity and note. It holds what the user said and nothing SymDiary
// concluded from it.
type Event struct {
	ID         EventID
	Definition DefinitionID
	// Symptom is the definition's label, exactly as the user typed it.
	Symptom string
	// OccurredAt is when the user says it happened. Editable.
	OccurredAt time.Time
	// RecordedAt is when SymDiary created the event. Set once (FR-007).
	RecordedAt time.Time
	Severity   Severity
	// Note is kept byte for byte as entered (FR-005).
	Note string
}

// ErrFutureOccurrence refuses an occurrence later than the moment it is
// recorded or edited (FR-006).
var ErrFutureOccurrence = errors.New("an occurrence cannot be in the future")

// CheckOccurrence refuses an occurrence time later than now.
func CheckOccurrence(occurred, now time.Time) error {
	if occurred.After(now) {
		return ErrFutureOccurrence
	}
	return nil
}

// ErrNoSuchEvent refuses an operation on an event that is not there.
var ErrNoSuchEvent = errors.New("no such event")

// Observation is what makes two events the same observation: the same symptom,
// the same two instants, the same severity and the same note. Store identity
// and the zone an instant was written in play no part. It is comparable, so an
// import can hold a set of them and skip what the record already has (FR-053).
type Observation struct {
	Symptom    SymptomKey
	OccurredAt int64
	RecordedAt int64
	Severity   Severity
	Note       string
}

// ObservationOf answers an event's observation.
func ObservationOf(e Event) Observation {
	return Observation{
		Symptom:    KeyOf(e.Symptom),
		OccurredAt: e.OccurredAt.UnixNano(),
		RecordedAt: e.RecordedAt.UnixNano(),
		Severity:   e.Severity,
		Note:       e.Note,
	}
}

// SameEvent reports whether two events record the same observation.
func SameEvent(a, b Event) bool { return ObservationOf(a) == ObservationOf(b) }

// chronological orders two events oldest first: by occurrence, then by
// recording, then by identity, so two stored events never tie.
func chronological(a, b Event) int {
	if compared := a.OccurredAt.Compare(b.OccurredAt); compared != 0 {
		return compared
	}
	if compared := a.RecordedAt.Compare(b.RecordedAt); compared != 0 {
		return compared
	}
	return cmp.Compare(a.ID, b.ID)
}
