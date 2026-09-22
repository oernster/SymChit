package domain

import (
	"errors"
	"testing"
	"time"
)

func TestBlankSymptomRefused(t *testing.T) {
	t.Parallel()
	for _, label := range []string{"", " ", "\t\n "} {
		if err := CheckLabel(label); !errors.Is(err, ErrBlankSymptom) {
			t.Errorf("CheckLabel(%q) = %v, want ErrBlankSymptom", label, err)
		}
	}
	if err := CheckLabel(" Tired "); err != nil {
		t.Errorf("CheckLabel with surrounding space = %v, want nil", err)
	}
}

func TestKeyIgnoresCaseAndSurroundingSpace(t *testing.T) {
	t.Parallel()
	if KeyOf("  Back Pain ") != KeyOf("back pain") {
		t.Error("case and surrounding space should not change the key")
	}
	if KeyOf("Back pain") == KeyOf("Backpain") {
		t.Error("inner space is part of the label")
	}
}

func TestFutureOccurrenceRefused(t *testing.T) {
	t.Parallel()
	now := at(t, 2026, time.September, 22, 15, 0)
	if err := CheckOccurrence(now.Add(time.Minute), now); !errors.Is(err, ErrFutureOccurrence) {
		t.Errorf("a minute ahead = %v, want ErrFutureOccurrence", err)
	}
	if err := CheckOccurrence(now, now); err != nil {
		t.Errorf("now = %v, want nil", err)
	}
	if err := CheckOccurrence(at(t, 2026, time.September, 22, 12, 30), now); err != nil {
		t.Errorf("earlier today = %v, want nil", err)
	}
}

func TestSameEventComparesEveryRecordedField(t *testing.T) {
	t.Parallel()
	when := at(t, 2026, time.September, 22, 17, 12)
	base := Event{
		ID: 1, Definition: 1, Symptom: "Tired", OccurredAt: when, RecordedAt: when,
		Severity: SeverityMild, Note: "Only been awake for about 10 minutes.",
	}
	copied := base
	copied.ID, copied.Definition, copied.Symptom = 9, 7, "tired"
	copied.OccurredAt = when.UTC()
	if !SameEvent(base, copied) {
		t.Error("identity, label case and zone of the same instant are not the observation")
	}
	changes := map[string]func(*Event){
		"symptom":  func(e *Event) { e.Symptom = "Headache" },
		"occurred": func(e *Event) { e.OccurredAt = when.Add(time.Minute) },
		"recorded": func(e *Event) { e.RecordedAt = when.Add(time.Minute) },
		"severity": func(e *Event) { e.Severity = SeveritySevere },
		"note":     func(e *Event) { e.Note += " " },
	}
	for field, change := range changes {
		changed := base
		change(&changed)
		if SameEvent(base, changed) {
			t.Errorf("a different %s should be a different observation", field)
		}
	}
}

func TestChronologicalBreaksEveryTie(t *testing.T) {
	t.Parallel()
	when := at(t, 2026, time.September, 22, 9, 0)
	first := event(1, 1, "Tired", when)
	laterRecorded := first
	laterRecorded.ID, laterRecorded.RecordedAt = 2, when.Add(time.Hour)
	sameTimes := first
	sameTimes.ID = 3
	cases := []struct {
		name string
		a, b Event
		want int
	}{
		{"by occurrence", first, event(4, 1, "Tired", when.Add(time.Minute)), -1},
		{"by recording", laterRecorded, first, 1},
		{"by identity", first, sameTimes, -1},
		{"itself", first, first, 0},
	}
	for _, c := range cases {
		if got := chronological(c.a, c.b); got != c.want {
			t.Errorf("%s: chronological = %d, want %d", c.name, got, c.want)
		}
	}
}
