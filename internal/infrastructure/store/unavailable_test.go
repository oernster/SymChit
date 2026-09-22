package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/symchit/internal/application"
	"github.com/oernster/symchit/internal/product"
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

// TestTheRecordSitsInTheFolderThePlatformNames pins the one fact every
// packaged claim about the record's whereabouts rests on: the record is the
// product's own folder inside whatever the platform calls the place for a
// user's configuration; nothing else decides it.
//
// That is what makes the Flatpak's promise true without a second code path.
// Flatpak redirects XDG_CONFIG_HOME into the sandbox, os.UserConfigDir reads
// it on Linux, so the record lands at
// ~/.var/app/uk.codecrafter.SymChit/config/SymChit/symchit.db, which is the
// path build_flatpak.sh prints when it finishes. A change here that reached
// for the home directory instead would move the record on Linux and macOS
// while leaving Windows looking correct, so it is asserted rather than
// assumed.
func TestTheRecordSitsInTheFolderThePlatformNames(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("this machine names no configuration folder: %v", err)
	}
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if want := filepath.Join(base, product.Name, product.RecordFileName); path != want {
		t.Errorf("DefaultPath = %q, want %q: the record belongs in the folder "+
			"the platform names, which is what the Flatpak's redirect relies on", path, want)
	}
}
