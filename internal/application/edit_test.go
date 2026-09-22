package application

import (
	"strings"
	"testing"

	"github.com/oernster/symchit/internal/domain"
)

// recorded answers a store holding one Tired event at 09:30 plus the editor over it.
func recorded(t *testing.T) (*fakeStore, Editor, domain.Event) {
	t.Helper()
	store := newFakeStore()
	morning := at(t, 9, 30)
	saved, err := NewRecorder(store, fixedClock{morning}).Record(RecordForm{Symptom: "Tired"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	return store, NewEditor(store, fixedClock{at(t, 14, 0)}, london(t)), saved
}

func TestEditKeepsOccurredAt(t *testing.T) {
	t.Parallel()
	_, editor, saved := recorded(t)
	edited, err := editor.Edit(saved.ID, EditForm{Symptom: "Tired", Note: "Went away after an hour."})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if !edited.OccurredAt.Equal(at(t, 9, 30)) {
		t.Errorf("occurred-at moved to %v; want 09:30", edited.OccurredAt)
	}
	if edited.Note != "Went away after an hour." {
		t.Errorf("note = %q", edited.Note)
	}
}

func TestEditLeavesRecordedAtAlone(t *testing.T) {
	t.Parallel()
	_, editor, saved := recorded(t)
	earlier := at(t, 8, 0)
	edited, err := editor.Edit(saved.ID, EditForm{
		Symptom: "Headache", OccurredAt: &earlier, Severity: domain.SeveritySevere,
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if !edited.RecordedAt.Equal(saved.RecordedAt) {
		t.Errorf("recorded-at changed from %v to %v", saved.RecordedAt, edited.RecordedAt)
	}
	if !edited.OccurredAt.Equal(earlier) || edited.Symptom != "Headache" ||
		edited.Severity != domain.SeveritySevere {
		t.Errorf("edit did not take: %+v", edited)
	}
}

func TestEditRefusals(t *testing.T) {
	t.Parallel()
	store, editor, saved := recorded(t)
	_, err := editor.Edit(saved.ID, EditForm{Symptom: ""})
	wantIs(t, "blank symptom", err, domain.ErrBlankSymptom)
	future := at(t, 14, 1)
	_, err = editor.Edit(saved.ID, EditForm{Symptom: "Tired", OccurredAt: &future})
	wantIs(t, "future occurrence", err, domain.ErrFutureOccurrence)
	_, err = editor.Edit(999, EditForm{Symptom: "Tired"})
	wantIs(t, "unknown event", err, domain.ErrNoSuchEvent)
	store.failUpdate = true
	_, err = editor.Edit(saved.ID, EditForm{Symptom: "Tired"})
	wantIs(t, "failed write", err, ErrNotChanged)
}

func TestFailedEditChangesNothing(t *testing.T) {
	t.Parallel()
	store, editor, saved := recorded(t)
	store.failUpdate = true
	if _, err := editor.Edit(saved.ID, EditForm{Symptom: "Headache"}); err == nil {
		t.Fatal("the edit should have failed")
	}
	held, _ := store.Event(saved.ID)
	if held.Symptom != "Tired" {
		t.Errorf("a failed edit changed the event to %+v", held)
	}
}

func TestDeleteNeedsConfirmation(t *testing.T) {
	t.Parallel()
	store, editor, saved := recorded(t)
	prompt, err := editor.DeletionPrompt([]domain.EventID{saved.ID})
	if err != nil {
		t.Fatalf("DeletionPrompt: %v", err)
	}
	if !strings.Contains(prompt, "Tired") || !strings.Contains(prompt, "22 Sep 2026 09:30") {
		t.Errorf("the prompt should name the symptom and time: %q", prompt)
	}
	many, err := editor.DeletionPrompt([]domain.EventID{1, 2, 3})
	if err != nil || !strings.Contains(many, "3 events") {
		t.Errorf("the prompt for several should count them: %q, %v", many, err)
	}
	if err := editor.Delete([]domain.EventID{saved.ID}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(store.events) != 0 {
		t.Errorf("the event is still there: %+v", store.events)
	}
}

func TestDeleteRefusals(t *testing.T) {
	t.Parallel()
	store, editor, saved := recorded(t)
	_, err := editor.DeletionPrompt(nil)
	wantIs(t, "prompt for nothing", err, ErrNothingChosen)
	wantIs(t, "delete nothing", editor.Delete(nil), ErrNothingChosen)
	store.failEvent = true
	_, err = editor.DeletionPrompt([]domain.EventID{saved.ID})
	wantIs(t, "prompt with the store failing", err, ErrNotRead)
	store.failDelete = true
	wantIs(t, "failed delete", editor.Delete([]domain.EventID{saved.ID}), ErrNotDeleted)
	if len(store.events) != 1 {
		t.Error("a failed delete removed the event")
	}
}
