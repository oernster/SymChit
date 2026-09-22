package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// ErrEmptyRange refuses a receipt over a range holding no events (FR-043).
var ErrEmptyRange = errors.New("no events were recorded in that range")

// ErrOpenRange refuses a receipt whose range lacks a start or an end.
var ErrOpenRange = errors.New("a receipt needs a start date and an end date")

// ReceiptGroup is one symptom's events within the receipt range, oldest first.
type ReceiptGroup struct {
	Definition DefinitionID
	Symptom    string
	Events     []Event
}

// Receipt is the printable symptom record for a date range (section 3.4). It
// holds the user's events and their counts; nothing else (FR-042).
type Receipt struct {
	From   Date
	To     Date
	Groups []ReceiptGroup
	zone   *time.Location
}

// BuildReceipt gathers the events of a range into symptom groups. Groups run in
// the order their symptoms first occurred in the range; within a group, events
// run oldest to newest. Labels play no part in the order (FR-044).
func BuildReceipt(events []Event, from, to Date, zone *time.Location) (Receipt, error) {
	if from.IsZero() || to.IsZero() {
		return Receipt{}, ErrOpenRange
	}
	if err := CheckRange(from, to); err != nil {
		return Receipt{}, err
	}
	inRange := Filter{From: from, To: to}
	kept := make([]Event, 0, len(events))
	for _, event := range events {
		if inRange.Matches(event, zone) {
			kept = append(kept, event)
		}
	}
	if len(kept) == 0 {
		return Receipt{}, ErrEmptyRange
	}
	slices.SortFunc(kept, chronological)

	var groups []ReceiptGroup
	position := map[DefinitionID]int{}
	for _, event := range kept {
		at, seen := position[event.Definition]
		if !seen {
			at = len(groups)
			position[event.Definition] = at
			groups = append(groups, ReceiptGroup{
				Definition: event.Definition, Symptom: event.Symptom,
			})
		}
		groups[at].Events = append(groups[at].Events, event)
	}
	return Receipt{From: from, To: to, Groups: groups, zone: zone}, nil
}

// LineKind says what a receipt line holds, so the page can style it without
// reading its words.
type LineKind string

// The kinds of receipt line. There are no others: a receipt holds a title, its
// range, a heading with a count per symptom, the recorded fields of events and
// the two framing lines of FR-045.
const (
	LineTitle      LineKind = "title"
	LineRange      LineKind = "range"
	LineHeading    LineKind = "heading"
	LineWhen       LineKind = "when"
	LineSeverity   LineKind = "severity"
	LineNote       LineKind = "note"
	LineProvenance LineKind = "provenance"
	LineStatement  LineKind = "statement"
)

// Framing is the fixed text printed around the record: what produced the sheet
// and what the sheet is (FR-045).
//
// It is handed in rather than read here. The words are the product's own and
// live in internal/product; the domain depends on nothing, so it is given them
// at the moment it writes the lines.
type Framing struct {
	// Provenance names the program and its address. It prints above the title
	// and again below the last event.
	Provenance string
	// Statement says the record is not a diagnosis. It prints below the range,
	// where a reader meets it before any recorded event.
	Statement string
}

// Line is one line of a receipt.
type Line struct {
	Kind LineKind
	Text string
}

// The receipt's own words. They are its only words besides what the user
// recorded, which is why they are gathered here.
const (
	receiptTitle     = "SYMPTOM RECORD"
	severityPrefix   = "Severity: "
	eventSingular    = "recorded event"
	eventPlural      = "recorded events"
	headingLayout    = "%s - %d %s"
	rangeLayout      = "%s - %s"
	dayMonthLayout   = "2 January"
	fullDateLayout   = "2 January 2006"
	whenLayout       = "02 Jan 15:04"
	whenYearedLayout = "02 Jan 2006 15:04"
)

// Lines writes the receipt as the lines it is printed in (FR-041), wrapped in
// the framing it is handed (FR-045).
//
// Event times leave out the year when the whole range lies in one year, as the
// source example does; a range that crosses a year names the year on every
// event, so no line can be read as the wrong year (Amendment 2).
func (r Receipt) Lines(framing Framing) []Line {
	lines := []Line{
		{Kind: LineProvenance, Text: framing.Provenance},
		{Kind: LineTitle, Text: receiptTitle},
		{Kind: LineRange, Text: r.rangeText()},
		{Kind: LineStatement, Text: framing.Statement},
	}
	layout := whenLayout
	if r.From.Year != r.To.Year {
		layout = whenYearedLayout
	}
	for _, group := range r.Groups {
		lines = append(lines, Line{Kind: LineHeading, Text: heading(group)})
		for _, event := range group.Events {
			lines = append(lines, Line{
				Kind: LineWhen, Text: event.OccurredAt.In(r.zone).Format(layout),
			})
			if event.Severity != SeverityNone {
				lines = append(lines, Line{
					Kind: LineSeverity, Text: severityPrefix + event.Severity.String(),
				})
			}
			if event.Note != "" {
				lines = append(lines, Line{Kind: LineNote, Text: event.Note})
			}
		}
	}
	return append(lines, Line{Kind: LineProvenance, Text: framing.Provenance})
}

// heading writes a group's heading: its symptom and how many events it holds.
func heading(group ReceiptGroup) string {
	noun := eventPlural
	if len(group.Events) == 1 {
		noun = eventSingular
	}
	return fmt.Sprintf(headingLayout, group.Symptom, len(group.Events), noun)
}

// rangeText writes the range as the source example does: 23 August - 22
// September 2026, naming the first year only where it differs.
func (r Receipt) rangeText() string {
	from := r.From.start(time.UTC)
	to := r.To.start(time.UTC)
	if r.From == r.To {
		return to.Format(fullDateLayout)
	}
	startLayout := dayMonthLayout
	if r.From.Year != r.To.Year {
		startLayout = fullDateLayout
	}
	return fmt.Sprintf(rangeLayout, from.Format(startLayout), to.Format(fullDateLayout))
}
