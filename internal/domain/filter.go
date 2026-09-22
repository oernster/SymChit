package domain

import (
	"slices"
	"time"
)

// Filter narrows the history (FR-021). An unset date leaves that end open; an
// empty list places no limit. Every set part narrows the others.
type Filter struct {
	From        Date
	To          Date
	Definitions []DefinitionID
	// Severities may include SeverityNone, which matches events given none.
	Severities []Severity
}

// Matches reports whether an event passes every part of the filter, reading its
// occurrence date in the given zone.
func (f Filter) Matches(event Event, zone *time.Location) bool {
	day := DateOf(event.OccurredAt, zone)
	if !f.From.IsZero() && day.Compare(f.From) < 0 {
		return false
	}
	if !f.To.IsZero() && day.Compare(f.To) > 0 {
		return false
	}
	if len(f.Definitions) > 0 && !slices.Contains(f.Definitions, event.Definition) {
		return false
	}
	if len(f.Severities) > 0 && !slices.Contains(f.Severities, event.Severity) {
		return false
	}
	return true
}

// Select answers the events that pass the filter, newest first (FR-020).
func Select(events []Event, f Filter, zone *time.Location) []Event {
	kept := make([]Event, 0, len(events))
	for _, event := range events {
		if f.Matches(event, zone) {
			kept = append(kept, event)
		}
	}
	slices.SortFunc(kept, func(a, b Event) int { return chronological(b, a) })
	return kept
}
