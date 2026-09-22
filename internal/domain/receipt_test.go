package domain

import (
	"errors"
	"testing"
	"time"
)

// sample answers the FR-041 acceptance events: Headache on 25 Aug and 2 Sep,
// Tired on 28 Aug and 22 Sep, handed over out of order.
func sample(t *testing.T) []Event {
	t.Helper()
	tired := event(4, 1, "Tired", at(t, 2026, time.September, 22, 17, 12))
	tired.Note = "Only been awake for about 10 minutes."
	headache := event(2, 2, "Headache", at(t, 2026, time.September, 2, 8, 10))
	headache.Severity = SeverityModerate
	return []Event{
		tired,
		event(3, 1, "Tired", at(t, 2026, time.August, 28, 7, 45)),
		headache,
		event(1, 2, "Headache", at(t, 2026, time.August, 25, 12, 30)),
		event(5, 2, "Headache", at(t, 2026, time.August, 22, 23, 59)),
	}
}

// lineTexts answers the text of each line.
func lineTexts(lines []Line) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line.Text)
	}
	return out
}

// assertLines fails unless got holds exactly the wanted texts.
func assertLines(t *testing.T, got []Line, want []string) {
	t.Helper()
	texts := lineTexts(got)
	if len(texts) != len(want) {
		t.Fatalf("receipt =\n%q\nwant\n%q", texts, want)
	}
	for i := range want {
		if texts[i] != want[i] {
			t.Fatalf("line %d = %q, want %q\nwhole receipt: %q", i, texts[i], want[i], texts)
		}
	}
}

func TestReceiptMatchesTheAcceptanceExample(t *testing.T) {
	t.Parallel()
	receipt, err := BuildReceipt(sample(t),
		Date{2026, time.August, 23}, Date{2026, time.September, 22}, london(t))
	if err != nil {
		t.Fatalf("BuildReceipt: %v", err)
	}
	assertLines(t, receipt.Lines(), []string{
		"SYMPTOM RECORD",
		"23 August - 22 September 2026",
		"Headache - 2 recorded events",
		"25 Aug 12:30",
		"02 Sep 08:10",
		"Severity: Moderate",
		"Tired - 2 recorded events",
		"28 Aug 07:45",
		"22 Sep 17:12",
		"Only been awake for about 10 minutes.",
	})
}

func TestReceiptOrder(t *testing.T) {
	t.Parallel()
	// Zebra first occurs before Aardvark, so it comes first whatever the alphabet
	// says; Aardvark has more events and that does not move it up either.
	events := []Event{
		event(1, 1, "Aardvark", at(t, 2026, time.September, 5, 9, 0)),
		event(2, 1, "Aardvark", at(t, 2026, time.September, 6, 9, 0)),
		event(3, 1, "Aardvark", at(t, 2026, time.September, 7, 9, 0)),
		event(4, 2, "Zebra", at(t, 2026, time.September, 4, 9, 0)),
	}
	receipt, err := BuildReceipt(events,
		Date{2026, time.September, 1}, Date{2026, time.September, 30}, london(t))
	if err != nil {
		t.Fatalf("BuildReceipt: %v", err)
	}
	if receipt.Groups[0].Symptom != "Zebra" || receipt.Groups[1].Symptom != "Aardvark" {
		t.Errorf("group order = %s, %s; want Zebra, Aardvark",
			receipt.Groups[0].Symptom, receipt.Groups[1].Symptom)
	}
	aardvark := receipt.Groups[1].Events
	if aardvark[0].ID != 1 || aardvark[2].ID != 3 {
		t.Errorf("events within a group should run oldest first: %+v", aardvark)
	}
}

func TestReceiptHoldsNoOtherText(t *testing.T) {
	t.Parallel()
	receipt, err := BuildReceipt(sample(t),
		Date{2026, time.August, 23}, Date{2026, time.September, 22}, london(t))
	if err != nil {
		t.Fatalf("BuildReceipt: %v", err)
	}
	allowed := map[LineKind]bool{
		LineTitle: true, LineRange: true, LineHeading: true,
		LineWhen: true, LineSeverity: true, LineNote: true,
	}
	notes := map[string]bool{"Only been awake for about 10 minutes.": true}
	for _, line := range receipt.Lines() {
		if !allowed[line.Kind] {
			t.Errorf("line %q has kind %q, which a receipt does not hold", line.Text, line.Kind)
		}
		if line.Kind == LineNote && !notes[line.Text] {
			t.Errorf("note line %q is not a note the user wrote", line.Text)
		}
	}
}

func TestReceiptAcrossAYearNamesTheYear(t *testing.T) {
	t.Parallel()
	events := []Event{
		event(1, 1, "Tired", at(t, 2025, time.December, 30, 9, 0)),
		event(2, 1, "Tired", at(t, 2026, time.January, 2, 9, 0)),
	}
	receipt, err := BuildReceipt(events,
		Date{2025, time.December, 23}, Date{2026, time.January, 22}, london(t))
	if err != nil {
		t.Fatalf("BuildReceipt: %v", err)
	}
	assertLines(t, receipt.Lines(), []string{
		"SYMPTOM RECORD",
		"23 December 2025 - 22 January 2026",
		"Tired - 2 recorded events",
		"30 Dec 2025 09:00",
		"02 Jan 2026 09:00",
	})
}

func TestReceiptOfOneDayAndOneEvent(t *testing.T) {
	t.Parallel()
	day := Date{2026, time.September, 22}
	receipt, err := BuildReceipt(sample(t)[:1], day, day, london(t))
	if err != nil {
		t.Fatalf("BuildReceipt: %v", err)
	}
	lines := lineTexts(receipt.Lines())
	if lines[1] != "22 September 2026" || lines[2] != "Tired - 1 recorded event" {
		t.Errorf("one day, one event reads %q", lines)
	}
}

func TestEmptyRangeGivesNoReceipt(t *testing.T) {
	t.Parallel()
	zone := london(t)
	_, err := BuildReceipt(sample(t), Date{2026, time.July, 1}, Date{2026, time.July, 31}, zone)
	if !errors.Is(err, ErrEmptyRange) {
		t.Errorf("July = %v, want ErrEmptyRange", err)
	}
	_, err = BuildReceipt(sample(t), Date{}, Date{2026, time.July, 31}, zone)
	if !errors.Is(err, ErrOpenRange) {
		t.Errorf("no start = %v, want ErrOpenRange", err)
	}
	_, err = BuildReceipt(sample(t), Date{2026, time.July, 31}, Date{2026, time.July, 1}, zone)
	if !errors.Is(err, ErrReversedRange) {
		t.Errorf("reversed = %v, want ErrReversedRange", err)
	}
}
