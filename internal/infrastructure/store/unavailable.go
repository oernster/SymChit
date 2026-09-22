package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oernster/symchit/internal/application"
	"github.com/oernster/symchit/internal/domain"
	"github.com/oernster/symchit/internal/product"
)

// DefaultPath answers where the record lives: %APPDATA%\SymChit\symchit.db
// (FR-060).
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding the folder for the record: %w", err)
	}
	return filepath.Join(base, product.Name, product.RecordFileName), nil
}

// Unavailable stands in for a record that could not be opened. Every operation
// refuses with the reason, so the window still opens and every action says why
// it cannot act, while the file itself is never touched (FR-062).
type Unavailable struct {
	Reason error
}

// Definitions refuses.
func (u Unavailable) Definitions() ([]domain.Definition, error) { return nil, u.Reason }

// Events refuses.
func (u Unavailable) Events() ([]domain.Event, error) { return nil, u.Reason }

// Event refuses.
func (u Unavailable) Event(domain.EventID) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// AddEvent refuses.
func (u Unavailable) AddEvent(application.EventInput) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// UpdateEvent refuses.
func (u Unavailable) UpdateEvent(domain.EventID, application.EventInput) (domain.Event, error) {
	return domain.Event{}, u.Reason
}

// DeleteEvents refuses.
func (u Unavailable) DeleteEvents([]domain.EventID) error { return u.Reason }

// RenameDefinition refuses.
func (u Unavailable) RenameDefinition(domain.DefinitionID, string, domain.SymptomKey) error {
	return u.Reason
}

// Import refuses.
func (u Unavailable) Import([]application.DefinitionInput, []application.EventInput) error {
	return u.Reason
}
