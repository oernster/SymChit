// Package domain holds SymChit's rules: what an event is, how symptoms are
// matched, how the history is filtered and what a receipt contains.
//
// It performs no IO and never reads the clock. Every instant and every time zone
// arrives as an argument, so each rule can be tested against fixed values.
package domain

import (
	"errors"
	"strings"
)

// ErrBlankSymptom refuses an event or a definition whose symptom is empty or
// holds nothing but white space (FR-002).
var ErrBlankSymptom = errors.New("a symptom is needed")

// SymptomKey is the form in which two symptom labels are compared: surrounding
// white space removed and letter case folded. It decides whether a typed label
// is a definition already held (FR-014). It is never shown; the label is.
type SymptomKey string

// KeyOf answers the comparison key of a label.
func KeyOf(label string) SymptomKey {
	return SymptomKey(strings.ToLower(strings.TrimSpace(label)))
}

// CheckLabel refuses a label with nothing in it. The label itself is never
// changed: SymChit keeps what the user typed (FR-013).
func CheckLabel(label string) error {
	if KeyOf(label) == "" {
		return ErrBlankSymptom
	}
	return nil
}
