package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/symchit/internal/domain"
)

// aTiredEvent puts one event in the record so there is something to write.
func aTiredEvent(t *testing.T, app *App) {
	t.Helper()
	record(t, app, RecordFormDTO{Symptom: "Tired", Note: "Only been awake for about 10 minutes."})
}

func TestSavePDFWritesTheRecordWhereTheReaderChose(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = filepath.Join(t.TempDir(), "record.pdf")

	path, err := app.SavePDF("2026-08-23", "2026-09-22")
	if err != nil {
		t.Fatalf("SavePDF: %v", err)
	}

	if path != chooser.save {
		t.Errorf("SavePDF answered %q, want the path the reader chose, %q", path, chooser.save)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading what was written: %v", err)
	}
	// A reader takes this to an appointment. It has to be a document a machine
	// they have never met will open, so the check is on the file itself.
	if !strings.HasPrefix(string(written), "%PDF-") {
		t.Errorf("what was written begins %q, which is not a PDF", string(written[:min(8, len(written))]))
	}
}

func TestSavePDFWritesNothingWhenTheReaderCancels(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = ""

	path, err := app.SavePDF("2026-08-23", "2026-09-22")

	if err != nil {
		t.Errorf("a cancelled dialog answered %v, want no error: cancelling is not a fault", err)
	}
	if path != "" {
		t.Errorf("a cancelled dialog answered %q, want nothing written", path)
	}
}

func TestSavePDFRefusesAnEmptyRangeBeforeAskingWhereToSave(t *testing.T) {
	t.Parallel()
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	unreachable := filepath.Join(t.TempDir(), "should-not-exist.pdf")
	chooser.save = unreachable

	_, err := app.SavePDF("2026-01-01", "2026-01-31")

	if !errors.Is(err, domain.ErrEmptyRange) {
		t.Errorf("SavePDF over an empty range = %v, want ErrEmptyRange", err)
	}
	// Asking where to save and then saying there is nothing to save wastes the
	// reader's answer, so the refusal comes first.
	if _, statErr := os.Stat(unreachable); statErr == nil {
		t.Error("a file was written for a range holding no events")
	}
}

func TestSavePDFRefusesADateItCannotRead(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	aTiredEvent(t, app)

	if _, err := app.SavePDF("not-a-date", "2026-09-22"); err == nil {
		t.Error("SavePDF with an unreadable start date answered no error")
	}
	if _, err := app.SavePDF("2026-08-23", "not-a-date"); err == nil {
		t.Error("SavePDF with an unreadable end date answered no error")
	}
}

func TestSavePDFWithNoWriterIsAnInternalFault(t *testing.T) {
	t.Parallel()
	// A facade built without a sheet is a wiring mistake rather than a state the
	// reader can reach, so it answers the same internal fault every other bound
	// method answers rather than ending the window.
	app, _, chooser := facade(t)
	aTiredEvent(t, app)
	chooser.save = filepath.Join(t.TempDir(), "record.pdf")
	app.sheet = nil

	if _, err := app.SavePDF("2026-08-23", "2026-09-22"); err == nil {
		t.Error("SavePDF with no sheet answered no error, want the internal fault")
	}
}
