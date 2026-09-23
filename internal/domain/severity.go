package domain

import (
	"errors"
	"fmt"
)

// Severity is the user's own optional rating of an event (FR-004). SymDiary never
// computes one.
type Severity int

// The severities. SeverityNone means the user gave none; it is the default.
const (
	SeverityNone Severity = iota
	SeverityMild
	SeverityModerate
	SeveritySevere
)

// severityNames holds each severity's name, indexed by value. The name is what
// the user sees, what the store keeps and what the export carries, so it has
// this one home.
var severityNames = [...]string{
	SeverityNone:     "",
	SeverityMild:     "Mild",
	SeverityModerate: "Moderate",
	SeveritySevere:   "Severe",
}

// ErrUnknownSeverity refuses a severity name outside the list.
var ErrUnknownSeverity = errors.New("unknown severity")

// Severities answers the choices a user can make, in the order they are offered.
// SeverityNone is not among them: it is the absence of a choice.
func Severities() []Severity {
	return []Severity{SeverityMild, SeverityModerate, SeveritySevere}
}

// String answers the severity's name; the empty string for SeverityNone.
func (s Severity) String() string {
	if s < SeverityNone || int(s) >= len(severityNames) {
		return fmt.Sprintf("Severity(%d)", int(s))
	}
	return severityNames[s]
}

// ParseSeverity answers the severity with the given name. The empty name is
// SeverityNone.
func ParseSeverity(name string) (Severity, error) {
	for value, known := range severityNames {
		if known == name {
			return Severity(value), nil
		}
	}
	return SeverityNone, fmt.Errorf("%w: %q", ErrUnknownSeverity, name)
}
