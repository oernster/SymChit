package pdf

import (
	"strings"
	"testing"

	"github.com/oernster/symchit/internal/domain"
)

// evenWidths measures every character the same, so what lands on which page is
// decided by the layout rather than by a font's opinion of the letter m.
type evenWidths struct{}

// WidthOf answers a width from the rune count and the size alone.
func (evenWidths) WidthOf(text string, size float64, _ bool) float64 {
	return float64(len([]rune(text))) * size * 0.2
}

func TestEveryStyleDrawsSomethingAReaderCanSee(t *testing.T) {
	t.Parallel()
	// That every kind HAS a style is held by the structural suite, against the
	// domain's own constants. This is the other half: that the style it has
	// puts something on the paper.
	for kind, drawn := range styles {
		if drawn.size <= 0 {
			t.Errorf("a %q line is drawn at size %v, which is nothing at all", kind, drawn.size)
		}
		if drawn.height() <= 0 {
			t.Errorf("a %q line is given no height", kind)
		}
	}
}

func TestALongNoteWrapsToThePage(t *testing.T) {
	t.Parallel()
	note := strings.Repeat("word ", 80)
	pages := Pages([]domain.Line{{Kind: domain.LineNote, Text: note}}, evenWidths{})

	if len(pages) == 0 {
		t.Fatal("a note laid out to nothing")
	}
	rows := pages[0]
	if len(rows) < 2 {
		t.Fatalf("a note of %d words came to %d rows, want it wrapped", 80, len(rows))
	}
	// Nothing the user wrote may be lost in the wrapping: a record that alters
	// what was written is worse than one that reads awkwardly.
	var rebuilt []string
	for _, row := range rows {
		rebuilt = append(rebuilt, row.Text)
	}
	if strings.Join(rebuilt, " ") != strings.TrimSpace(note) {
		t.Error("the wrapped note is not the note that went in")
	}
}

func TestAWordWiderThanThePageIsLeftWhole(t *testing.T) {
	t.Parallel()
	// It is something the user typed. Running into the margin is a worse look
	// and a better record than cutting their word in half.
	long := strings.Repeat("z", 400)
	pages := Pages([]domain.Line{{Kind: domain.LineNote, Text: long}}, evenWidths{})

	if len(pages[0]) != 1 || pages[0][0].Text != long {
		t.Errorf("a %d character word came out as %d rows", len(long), len(pages[0]))
	}
}

func TestAnEmptyLineStillTakesItsPlace(t *testing.T) {
	t.Parallel()
	pages := Pages([]domain.Line{{Kind: domain.LineNote, Text: "   "}}, evenWidths{})
	if len(pages) != 1 || len(pages[0]) != 1 {
		t.Fatalf("an empty note came to %d page(s)", len(pages))
	}
	if pages[0][0].Text != "" {
		t.Errorf("an empty note drew %q", pages[0][0].Text)
	}
}

// anEvent is the lines one recorded event becomes.
func anEvent(note string) []domain.Line {
	return []domain.Line{
		{Kind: domain.LineWhen, Text: "22 Sep 17:12"},
		{Kind: domain.LineSeverity, Text: "Severity: Moderate"},
		{Kind: domain.LineNote, Text: note},
	}
}

func TestAnEventIsNeverSplitAcrossTwoPages(t *testing.T) {
	t.Parallel()
	// A reader must never meet a time at the foot of one sheet and what was
	// observed at the head of the next: that is how a record gets misread.
	var lines []domain.Line
	for i := 0; i < 60; i++ {
		lines = append(lines, anEvent("Felt unusually tired shortly after getting up.")...)
	}
	pages := Pages(lines, evenWidths{})
	if len(pages) < 2 {
		t.Fatalf("60 events came to %d page(s), want more than one", len(pages))
	}

	for number, rows := range pages {
		for i, row := range rows {
			if row.Kind != domain.LineWhen {
				continue
			}
			if i+2 >= len(rows) {
				t.Errorf("page %d ends with a time and not what was observed", number+1)
				break
			}
		}
		if rows[0].Kind == domain.LineSeverity || rows[0].Kind == domain.LineNote {
			t.Errorf("page %d opens with %q, which belongs to a time on the page before",
				number+1, rows[0].Kind)
		}
	}
}

func TestEveryRowSitsInsideThePage(t *testing.T) {
	t.Parallel()
	var lines []domain.Line
	for i := 0; i < 40; i++ {
		lines = append(lines, anEvent(strings.Repeat("long ", 40))...)
	}

	for number, rows := range Pages(lines, evenWidths{}) {
		for _, row := range rows {
			if row.Top < 0 {
				t.Errorf("page %d has a row above the paper at %.1fmm", number+1, row.Top)
			}
			// The foot margin is where the page number goes, so a row reaching
			// into it would be drawn over the top of that number.
			if bottom := row.Top + row.Style.height(); bottom > PageHeight-marginBottom {
				t.Errorf("page %d has a row reaching %.1fmm, past the %.1fmm the record ends at",
					number+1, bottom, PageHeight-marginBottom)
			}
		}
	}
}

func TestAPageOpensAtTheSameHeightAsTheOneBeforeIt(t *testing.T) {
	t.Parallel()
	// The gap a heading leaves above itself is separation from what came before.
	// At the head of a fresh page there is nothing above it; keeping the gap
	// would give each sheet a different top margin.
	var lines []domain.Line
	for i := 0; i < 30; i++ {
		lines = append(lines, domain.Line{Kind: domain.LineHeading, Text: "Tiredness - 2 recorded events"})
		lines = append(lines, anEvent("A note.")...)
	}

	pages := Pages(lines, evenWidths{})
	if len(pages) < 2 {
		t.Fatalf("laid out to %d page(s), want more than one", len(pages))
	}
	for number, rows := range pages {
		if rows[0].Top != marginTop {
			t.Errorf("page %d opens at %.1fmm, want every page to open at %.1fmm",
				number+1, rows[0].Top, marginTop)
		}
	}
}

func TestARecordOfNothingLaysOutToNothing(t *testing.T) {
	t.Parallel()
	if pages := Pages(nil, evenWidths{}); len(pages) != 0 {
		t.Errorf("no lines came to %d page(s), want none", len(pages))
	}
}
