package store

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/oernster/symchit/internal/application"
)

func TestUnavailableRefusesEverythingWithTheReason(t *testing.T) {
	t.Parallel()
	reason := errors.New("the record could not be read")
	var store application.Store = Unavailable{Reason: reason}
	_, e1 := store.Definitions()
	_, e2 := store.Events()
	_, e3 := store.Event(1)
	_, e4 := store.AddEvent(application.EventInput{})
	_, e5 := store.UpdateEvent(1, application.EventInput{})
	e6 := store.DeleteEvents(nil)
	e7 := store.RenameDefinition(1, "", "")
	e8 := store.Import(nil, nil)
	for i, err := range []error{e1, e2, e3, e4, e5, e6, e7, e8} {
		if !errors.Is(err, reason) {
			t.Errorf("operation %d answered %v, want the reason", i+1, err)
		}
	}
}

func TestOpensAtTheGivenPath(t *testing.T) {
	base := t.TempDir()
	t.Setenv("APPDATA", base)
	path, err := DefaultPath()
	if want := filepath.Join(base, "SymChit", "symchit.db"); err != nil || path != want {
		t.Errorf("DefaultPath = %q, %v; want %q", path, err, want)
	}
	t.Setenv("APPDATA", "")
	if _, err := DefaultPath(); err == nil {
		t.Error("with no APPDATA there is no folder for the record")
	}
}
