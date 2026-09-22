package application

import (
	"testing"

	"github.com/oernster/symchit/internal/domain"
)

func TestRecordNowUsesTheClock(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	recorder := NewRecorder(store, fixedClock{at(t, 17, 12)})
	saved, err := recorder.Record(RecordForm{
		Symptom: "Tired", Note: "Only been awake for about 10 minutes.",
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !saved.OccurredAt.Equal(at(t, 17, 12)) || !saved.RecordedAt.Equal(at(t, 17, 12)) {
		t.Errorf("times = %v, %v; want 17:12 for both", saved.OccurredAt, saved.RecordedAt)
	}
	if saved.Symptom != "Tired" || saved.Severity != domain.SeverityNone ||
		saved.Note != "Only been awake for about 10 minutes." {
		t.Errorf("saved = %+v", saved)
	}
}

func TestRetrospectiveTimeKept(t *testing.T) {
	t.Parallel()
	recorder := NewRecorder(newFakeStore(), fixedClock{at(t, 15, 0)})
	earlier := at(t, 12, 30)
	saved, err := recorder.Record(RecordForm{Symptom: "Headache", OccurredAt: &earlier})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !saved.OccurredAt.Equal(earlier) || !saved.RecordedAt.Equal(at(t, 15, 0)) {
		t.Errorf("occurred %v recorded %v; want 12:30 and 15:00", saved.OccurredAt, saved.RecordedAt)
	}
}

func TestRecordRefusals(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	recorder := NewRecorder(store, fixedClock{at(t, 15, 0)})
	_, err := recorder.Record(RecordForm{Symptom: "  "})
	wantIs(t, "blank symptom", err, domain.ErrBlankSymptom)
	later := at(t, 15, 1)
	_, err = recorder.Record(RecordForm{Symptom: "Tired", OccurredAt: &later})
	wantIs(t, "future occurrence", err, domain.ErrFutureOccurrence)
	if len(store.events) != 0 {
		t.Errorf("a refused event was stored: %+v", store.events)
	}
}

func TestWriteFailureLosesNothing(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	store.failAdd = true
	_, err := NewRecorder(store, fixedClock{at(t, 15, 0)}).Record(RecordForm{Symptom: "Tired"})
	wantIs(t, "failed write", err, ErrNotSaved)
	wantIs(t, "failed write", err, errBroken)
}

func TestNewSymptomBecomesDefinition(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	recorder := NewRecorder(store, fixedClock{at(t, 15, 0)})
	for _, typed := range []string{"Tired", " tired", "Headache"} {
		if _, err := recorder.Record(RecordForm{Symptom: typed}); err != nil {
			t.Fatalf("Record(%q): %v", typed, err)
		}
	}
	history := NewHistory(store, london(t))
	symptoms, err := history.Symptoms()
	if err != nil {
		t.Fatalf("Symptoms: %v", err)
	}
	if len(symptoms) != 2 || symptoms[0].Label != "Headache" || symptoms[1].Label != "Tired" {
		t.Errorf("definitions = %+v, want Headache and Tired as first typed", symptoms)
	}
}
