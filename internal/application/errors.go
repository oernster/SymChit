package application

import (
	"errors"
	"fmt"
)

// The refusals the application answers when the store or a file fails. Each is
// wrapped round the underlying reason, so the message names what was not done
// and why (FR-009, FR-027, FR-052).
var (
	ErrNotSaved      = errors.New("the event was not saved")
	ErrNotChanged    = errors.New("the change was not saved")
	ErrNotDeleted    = errors.New("nothing was deleted")
	ErrNotRenamed    = errors.New("the symptom was not renamed")
	ErrNotRead       = errors.New("the record could not be read")
	ErrNotExported   = errors.New("the export was not written")
	ErrNotImported   = errors.New("nothing was imported")
	ErrNothingChosen = errors.New("no events were chosen")
)

// because wraps a failure in the refusal it caused, keeping both inspectable
// with errors.Is.
func because(refusal, reason error) error {
	return fmt.Errorf("%w: %w", refusal, reason)
}
