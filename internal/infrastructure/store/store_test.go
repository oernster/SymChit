package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/application"
	"github.com/oernster/symdiary/internal/domain"
)

// open answers a store over a fresh file in a folder whose name has a space in
// it, as %APPDATA% under a user name with a space would.
func open(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "App Data", "SymDiary", "symdiary.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, path
}

// instant answers a fixed instant with a one-hour offset, as British Summer
// Time writes it.
func instant(hour, minute int) time.Time {
	return time.Date(2026, time.September, 22, hour, minute, 0, 0, time.FixedZone("BST", 3600))
}

// input answers an event input for a label, occurring and recorded at a time.
func input(label string, when time.Time) application.EventInput {
	return application.EventInput{
		Label: label, Key: domain.KeyOf(label), OccurredAt: when, RecordedAt: when,
	}
}

func TestMissingFileGivesEmptyRecord(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the record file was not created: %v", err)
	}
	events, err := store.Events()
	if err != nil || len(events) != 0 {
		t.Errorf("a new record holds %v, %v", events, err)
	}
	definitions, err := store.Definitions()
	if err != nil || len(definitions) != 0 {
		t.Errorf("a new record holds definitions %v, %v", definitions, err)
	}
}

func TestNoteAndSymptomRoundTripByteForByte(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	note := "  Much worse than yesterday.  \n\tAnd a second line; ümlaut, emoji \U0001F635."
	given := input(" Kopfschmerzen ", instant(12, 30))
	given.Note, given.Severity = note, domain.SeverityModerate
	saved, err := store.AddEvent(given)
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	read, err := store.Event(saved.ID)
	if err != nil {
		t.Fatalf("Event: %v", err)
	}
	if read.Note != note || read.Symptom != " Kopfschmerzen " || read.Severity != domain.SeverityModerate {
		t.Errorf("read back %+v", read)
	}
	if !read.OccurredAt.Equal(instant(12, 30)) || !read.RecordedAt.Equal(instant(12, 30)) {
		t.Errorf("times read back as %v and %v", read.OccurredAt, read.RecordedAt)
	}
}

func TestCaseVariantReusesTheHeldLabel(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	first, _ := store.AddEvent(input("Tired", instant(9, 0)))
	second, err := store.AddEvent(input("  TIRED", instant(17, 12)))
	if err != nil {
		t.Fatalf("AddEvent: %v", err)
	}
	if second.Definition != first.Definition || second.Symptom != "Tired" {
		t.Errorf("the variant made %+v, want the held Tired", second)
	}
	definitions, _ := store.Definitions()
	if len(definitions) != 1 || !definitions[0].LastUsed.Equal(instant(17, 12)) {
		t.Errorf("definitions = %+v, want one, last used 17:12", definitions)
	}
}

func TestUpdateCannotMoveRecordedAt(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	saved, _ := store.AddEvent(input("Tired", instant(9, 30)))
	change := input("Headache", instant(8, 0))
	change.RecordedAt = instant(23, 59)
	change.Note = "edited"
	updated, err := store.UpdateEvent(saved.ID, change)
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if !updated.RecordedAt.Equal(instant(9, 30)) {
		t.Errorf("recorded-at moved to %v", updated.RecordedAt)
	}
	if updated.Symptom != "Headache" || !updated.OccurredAt.Equal(instant(8, 0)) || updated.Note != "edited" {
		t.Errorf("update did not take: %+v", updated)
	}
	if _, err := store.UpdateEvent(999, change); !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("updating a missing event = %v, want ErrNoSuchEvent", err)
	}
}

func TestDeleteIsAllOrNothing(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	one, _ := store.AddEvent(input("Tired", instant(9, 0)))
	two, _ := store.AddEvent(input("Tired", instant(10, 0)))
	err := store.DeleteEvents([]domain.EventID{one.ID, 999})
	if !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("deleting with a missing id = %v, want ErrNoSuchEvent", err)
	}
	if events, _ := store.Events(); len(events) != 2 {
		t.Fatalf("a refused deletion removed events: %d left", len(events))
	}
	if err := store.DeleteEvents([]domain.EventID{one.ID, two.ID}); err != nil {
		t.Fatalf("DeleteEvents: %v", err)
	}
	if events, _ := store.Events(); len(events) != 0 {
		t.Errorf("%d events left after deleting both", len(events))
	}
	if _, err := store.Event(one.ID); !errors.Is(err, domain.ErrNoSuchEvent) {
		t.Errorf("a deleted event reads %v", err)
	}
}

func TestRename(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	tired, _ := store.AddEvent(input("Tired", instant(9, 0)))
	_, _ = store.AddEvent(input("Headache", instant(10, 0)))
	if err := store.RenameDefinition(tired.Definition, "Exhausted", domain.KeyOf("Exhausted")); err != nil {
		t.Fatalf("RenameDefinition: %v", err)
	}
	if read, _ := store.Event(tired.ID); read.Symptom != "Exhausted" {
		t.Errorf("the event reads %q after the rename", read.Symptom)
	}
	if err := store.RenameDefinition(tired.Definition, "HEADACHE", domain.KeyOf("HEADACHE")); err == nil {
		t.Error("renaming onto a held key should be refused by the unique key")
	}
	if err := store.RenameDefinition(999, "Nausea", domain.KeyOf("Nausea")); !errors.Is(err, domain.ErrNoSuchDefinition) {
		t.Errorf("renaming a missing definition = %v", err)
	}
}

func TestImportIsOneTransaction(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	definitions := []application.DefinitionInput{
		{Label: "Nausea", Key: domain.KeyOf("Nausea"), LastUsed: instant(8, 0)},
	}
	events := []application.EventInput{input("Tired", instant(9, 0)), input("tired", instant(10, 0))}
	if err := store.Import(definitions, events); err != nil {
		t.Fatalf("Import: %v", err)
	}
	held, _ := store.Definitions()
	if len(held) != 2 {
		t.Errorf("definitions after import = %+v, want Nausea and Tired", held)
	}

}

// refuseNote plants a trigger that makes any insert of an event with the given
// note fail, so a failure can be forced part way through a write.
func refuseNote(t *testing.T, store *Store, note string) {
	t.Helper()
	if _, err := store.db.Exec(`CREATE TRIGGER refuse BEFORE INSERT ON events
		WHEN NEW.note = '` + note + `' BEGIN SELECT RAISE(ABORT, 'refused'); END`); err != nil {
		t.Fatal(err)
	}
}

func TestFailedImportLeavesNothing(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	refuseNote(t, store, "boom")
	definitions := []application.DefinitionInput{
		{Label: "Nausea", Key: domain.KeyOf("Nausea"), LastUsed: instant(8, 0)},
	}
	good := input("Dizziness", instant(11, 0))
	bad := input("Dizziness", instant(12, 0))
	bad.Note = "boom"
	if err := store.Import(definitions, []application.EventInput{good, bad}); err == nil {
		t.Fatal("the planted refusal did not fail the import")
	}
	events, _ := store.Events()
	held, _ := store.Definitions()
	if len(events) != 0 || len(held) != 0 {
		t.Errorf("a failed import left %d events and %d definitions", len(events), len(held))
	}
}

func TestFailedAddLeavesNothing(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	refuseNote(t, store, "boom")
	bad := input("Tired", instant(9, 0))
	bad.Note = "boom"
	if _, err := store.AddEvent(bad); err == nil {
		t.Fatal("the planted refusal did not fail the add")
	}
	if held, _ := store.Definitions(); len(held) != 0 {
		t.Errorf("a failed add left the definition it created: %+v", held)
	}
}

func TestUnreadableValuesAreRefusedNotGuessed(t *testing.T) {
	t.Parallel()
	for _, column := range []string{"occurred_at", "recorded_at", "severity"} {
		store, _ := open(t)
		saved, _ := store.AddEvent(input("Tired", instant(9, 0)))
		if _, err := store.db.Exec(`UPDATE events SET ` + column + ` = 'garbage'`); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Events(); err == nil {
			t.Errorf("a garbage %s read back without complaint", column)
		}
		if _, err := store.Event(saved.ID); err == nil {
			t.Errorf("a garbage %s read back one event without complaint", column)
		}
	}
	store, _ := open(t)
	_, _ = store.AddEvent(input("Tired", instant(9, 0)))
	if _, err := store.db.Exec(`UPDATE definitions SET last_used = 'garbage'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Definitions(); err == nil {
		t.Error("a garbage last_used read back without complaint")
	}
	if _, err := store.AddEvent(input("Tired", instant(10, 0))); err == nil {
		t.Error("resolving onto a garbage last_used should fail")
	}
}

func TestClosedStoreRefusesEverything(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	_ = store.Close()
	if _, err := store.Definitions(); err == nil {
		t.Error("Definitions on a closed store")
	}
	if _, err := store.Events(); err == nil {
		t.Error("Events on a closed store")
	}
	if _, err := store.AddEvent(input("Tired", instant(9, 0))); err == nil {
		t.Error("AddEvent on a closed store")
	}
	if err := store.DeleteEvents([]domain.EventID{1}); err == nil {
		t.Error("DeleteEvents on a closed store")
	}
	if err := store.RenameDefinition(1, "X", "x"); err == nil {
		t.Error("RenameDefinition on a closed store")
	}
	if err := store.Import(nil, nil); err == nil {
		t.Error("Import on a closed store")
	}
}

func TestOpenRefusesAFolderItCannotMake(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(filepath.Join(blocker, "symdiary.db")); err == nil {
		t.Error("a record under a file opened")
	}
}

func TestReopenKeepsEverything(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	saved, _ := store.AddEvent(input("Tired", instant(9, 0)))
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	again, err := Open(path)
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	defer again.Close()
	if read, err := again.Event(saved.ID); err != nil || read.Symptom != "Tired" {
		t.Errorf("after reopening: %+v, %v", read, err)
	}
}

func TestCorruptFileIsNeverReplaced(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "symdiary.db")
	garbage := bytes.Repeat([]byte("not a database "), 512)
	if err := os.WriteFile(path, garbage, 0o600); err != nil {
		t.Fatal(err)
	}
	if store, err := Open(path); err == nil {
		_ = store.Close()
		t.Fatal("a garbage file opened as a record")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, garbage) {
		t.Errorf("the unreadable file was changed (%v)", err)
	}
}

func TestNewerSchemaRefused(t *testing.T) {
	t.Parallel()
	store, path := open(t)
	if _, err := store.db.Exec(`PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	_ = store.Close()
	if _, err := Open(path); !errors.Is(err, ErrNewerSchema) {
		t.Errorf("a newer schema = %v, want ErrNewerSchema", err)
	}
}

func TestDurabilityPragmas(t *testing.T) {
	t.Parallel()
	store, _ := open(t)
	// synchronous FULL is 2; see https://sqlite.org/pragma.html#pragma_synchronous.
	const synchronousFull = 2
	var synchronous, foreignKeys, version int
	var journal string
	_ = store.db.QueryRow(`PRAGMA synchronous`).Scan(&synchronous)
	_ = store.db.QueryRow(`PRAGMA journal_mode`).Scan(&journal)
	_ = store.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys)
	_ = store.db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if synchronous != synchronousFull || journal != "wal" || foreignKeys != 1 {
		t.Errorf("synchronous=%d journal=%s foreign_keys=%d", synchronous, journal, foreignKeys)
	}
	if version != len(migrations) {
		t.Errorf("schema version %d, want %d", version, len(migrations))
	}
}
