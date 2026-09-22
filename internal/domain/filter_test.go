package domain

import (
	"errors"
	"testing"
	"time"
)

func TestFiltersCombine(t *testing.T) {
	t.Parallel()
	zone := london(t)
	var events []Event
	for day := 1; day <= 12; day++ {
		events = append(events, event(EventID(day), 1, "Tired", at(t, 2026, time.September, day, 8, 0)))
	}
	for i, day := range []int{3, 10, 20} {
		headache := event(EventID(100+i), 2, "Headache", at(t, 2026, time.September, day, 14, 0))
		headache.Severity = SeverityModerate
		events = append(events, headache)
	}
	events = append(events, event(200, 2, "Headache", at(t, 2026, time.October, 1, 9, 0)))

	september := Filter{
		From: Date{2026, time.September, 1}, To: Date{2026, time.September, 30},
		Definitions: []DefinitionID{2},
	}
	if got := len(Select(events, september, zone)); got != 3 {
		t.Errorf("Headache in September = %d events, want 3", got)
	}
	openEnded := Filter{From: Date{2026, time.September, 11}}
	if got := len(Select(events, openEnded, zone)); got != 4 {
		t.Errorf("from 11 September = %d events, want 4", got)
	}
	upTo := Filter{To: Date{2026, time.September, 2}}
	if got := len(Select(events, upTo, zone)); got != 2 {
		t.Errorf("up to 2 September = %d events, want 2", got)
	}
	unrated := Filter{Severities: []Severity{SeverityNone}}
	if got := len(Select(events, unrated, zone)); got != 13 {
		t.Errorf("events given no severity = %d, want 13", got)
	}
	if got := len(Select(events, Filter{}, zone)); got != len(events) {
		t.Errorf("an empty filter kept %d of %d", got, len(events))
	}
}

func TestDateBoundaryIsLocal(t *testing.T) {
	t.Parallel()
	// 23:30 on 30 September in London is 22:30 UTC the same day; 00:30 on 1 October
	// in London is still 30 September in UTC. The local date decides.
	lateNight := event(1, 1, "Tired", at(t, 2026, time.September, 30, 23, 30))
	pastMidnight := event(2, 1, "Tired", at(t, 2026, time.October, 1, 0, 30))
	september := Filter{To: Date{2026, time.September, 30}}
	got := Select([]Event{lateNight, pastMidnight}, september, london(t))
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("September kept %+v, want only the 23:30 event", got)
	}
}

func TestNewestFirst(t *testing.T) {
	t.Parallel()
	events := []Event{
		event(1, 1, "Tired", at(t, 2026, time.September, 19, 9, 30)),
		event(2, 1, "Tired", at(t, 2026, time.September, 22, 17, 12)),
		event(3, 2, "Headache", at(t, 2026, time.September, 20, 12, 0)),
	}
	got := Select(events, Filter{}, london(t))
	want := []EventID{2, 3, 1}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("order = %v, want ids %v", got, want)
		}
	}
}

func TestDates(t *testing.T) {
	t.Parallel()
	zone := london(t)
	parsed, err := ParseDate("2026-09-22")
	if err != nil || parsed != (Date{2026, time.September, 22}) || parsed.String() != "2026-09-22" {
		t.Errorf("ParseDate = %v, %v", parsed, err)
	}
	if _, err := ParseDate("22/09/2026"); !errors.Is(err, ErrBadDate) {
		t.Errorf("a non-ISO date = %v, want ErrBadDate", err)
	}
	local, err := ParseLocal("2026-09-22T12:30", zone)
	if err != nil || !local.Equal(at(t, 2026, time.September, 22, 12, 30)) {
		t.Errorf("ParseLocal = %v, %v", local, err)
	}
	if FormatLocal(local.UTC(), zone) != "2026-09-22T12:30" {
		t.Errorf("FormatLocal = %q", FormatLocal(local.UTC(), zone))
	}
	if _, err := ParseLocal("12:30", zone); !errors.Is(err, ErrBadDate) {
		t.Errorf("a time with no date = %v, want ErrBadDate", err)
	}
	if got := FormatDisplay(local.UTC(), zone); got != "22 Sep 2026 12:30" {
		t.Errorf("FormatDisplay = %q", got)
	}
	if !(Date{}).IsZero() || parsed.IsZero() {
		t.Error("IsZero is wrong")
	}
}

func TestRangeChecks(t *testing.T) {
	t.Parallel()
	early, late := Date{2026, time.August, 23}, Date{2026, time.September, 22}
	if err := CheckRange(late, early); !errors.Is(err, ErrReversedRange) {
		t.Errorf("reversed = %v, want ErrReversedRange", err)
	}
	for _, pair := range [][2]Date{{early, late}, {early, early}, {{}, early}, {late, {}}} {
		if err := CheckRange(pair[0], pair[1]); err != nil {
			t.Errorf("CheckRange(%v, %v) = %v, want nil", pair[0], pair[1], err)
		}
	}
}
