package application

import (
	"testing"

	"github.com/oernster/symdiary/internal/domain"
)

const exportPath = `C:\exports\symdiary.json`

func TestExportThenImportRoundTrip(t *testing.T) {
	t.Parallel()
	source, _ := filled(t)
	source.definitions = append(source.definitions, domain.Definition{ID: 50, Label: "Nausea"})
	file := &fakeFile{written: map[string]Record{}}
	if err := NewTransfer(source, file).Export(exportPath); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if got := len(file.written[exportPath].Events); got != 3 {
		t.Fatalf("export holds %d events, want 3", got)
	}

	target := newFakeStore()
	result, err := NewTransfer(target, file).Import(exportPath)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if result.Added != 3 || result.Skipped != 0 {
		t.Errorf("first import = %+v, want 3 added", result)
	}
	if _, ok := domain.FindDefinition(target.definitions, "Nausea"); !ok {
		t.Error("a definition with no events was lost")
	}
	again, err := NewTransfer(target, file).Import(exportPath)
	if err != nil || again.Added != 0 || again.Skipped != 3 {
		t.Errorf("second import = %+v, %v; want everything skipped", again, err)
	}
}

func TestImportSkipsRepeatsWithinTheFile(t *testing.T) {
	t.Parallel()
	when := at(t, 9, 0)
	repeated := domain.Event{Symptom: "Tired", OccurredAt: when, RecordedAt: when}
	file := &fakeFile{written: map[string]Record{
		exportPath: {Events: []domain.Event{repeated, repeated}},
	}}
	result, err := NewTransfer(newFakeStore(), file).Import(exportPath)
	if err != nil || result.Added != 1 || result.Skipped != 1 {
		t.Errorf("import = %+v, %v; want 1 added and 1 skipped", result, err)
	}
}

func TestTransferRefusals(t *testing.T) {
	t.Parallel()
	store, _ := filled(t)
	file := &fakeFile{written: map[string]Record{}}
	transfer := NewTransfer(store, file)

	file.failWrite = true
	wantIs(t, "failed write", transfer.Export(exportPath), ErrNotExported)
	file.failWrite = false
	if err := transfer.Export(exportPath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	file.failRead = true
	_, err := transfer.Import(exportPath)
	wantIs(t, "failed read", err, ErrNotImported)
	file.failRead = false

	store.failImport = true
	_, err = transfer.Import(exportPath)
	wantIs(t, "failed store import", err, ErrNotImported)

	store.failReads = true
	wantIs(t, "export with the store failing", transfer.Export(exportPath), ErrNotExported)
	_, err = transfer.Import(exportPath)
	wantIs(t, "import with the store failing", err, ErrNotImported)
}

func TestExportFailsWhenEventsCannotBeRead(t *testing.T) {
	t.Parallel()
	store := &eventsFailStore{fakeStore: newFakeStore()}
	err := NewTransfer(store, &fakeFile{written: map[string]Record{}}).Export(exportPath)
	wantIs(t, "events unreadable", err, ErrNotExported)
}

// eventsFailStore reads definitions but not events, reaching the second read's
// failure path in Export.
type eventsFailStore struct{ *fakeStore }

func (eventsFailStore) Events() ([]domain.Event, error) { return nil, errBroken }
