package domain

import (
	"cmp"
	"errors"
	"slices"
	"strings"
	"time"
)

// Definition is a symptom kept for reuse, so it is offered the next time
// (FR-010, FR-011).
type Definition struct {
	ID DefinitionID
	// Label is the symptom exactly as first typed, else as last renamed.
	Label string
	// LastUsed is the latest recording instant of an event using it; it orders
	// the suggestions.
	LastUsed time.Time
}

// ErrDefinitionExists refuses a rename onto a label another definition already
// holds (FR-029).
var ErrDefinitionExists = errors.New("another symptom already has that name")

// ErrNoSuchDefinition refuses an operation on a definition that is not there.
var ErrNoSuchDefinition = errors.New("no such symptom")

// FindDefinition answers the definition whose label matches the given one
// ignoring case and surrounding space (FR-014).
func FindDefinition(definitions []Definition, label string) (Definition, bool) {
	key := KeyOf(label)
	for _, definition := range definitions {
		if KeyOf(definition.Label) == key {
			return definition, true
		}
	}
	return Definition{}, false
}

// Suggest answers every definition whose label contains the typed text, ignoring
// case, most recently used first (FR-011). Nothing typed suggests everything.
func Suggest(definitions []Definition, typed string) []Definition {
	wanted := string(KeyOf(typed))
	found := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		if strings.Contains(string(KeyOf(definition.Label)), wanted) {
			found = append(found, definition)
		}
	}
	slices.SortStableFunc(found, func(a, b Definition) int {
		if compared := b.LastUsed.Compare(a.LastUsed); compared != 0 {
			return compared
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return found
}

// CheckRename refuses three renames: to a blank label, of a definition that is
// not held, onto a label another definition holds. Renaming a definition to a
// case variant of its own label is allowed.
func CheckRename(definitions []Definition, id DefinitionID, label string) error {
	if err := CheckLabel(label); err != nil {
		return err
	}
	if !slices.ContainsFunc(definitions, func(d Definition) bool { return d.ID == id }) {
		return ErrNoSuchDefinition
	}
	if other, found := FindDefinition(definitions, label); found && other.ID != id {
		return ErrDefinitionExists
	}
	return nil
}
