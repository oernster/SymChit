package structural

// What keeps the printed record off the edge of the paper, plus what numbers it.
//
// Three edges of the sheet have no page margin at all, which is the only way to
// leave the browser nowhere to draw its own header and footer (FR-045,
// Amendment 14). What holds the record clear of those three is the page's own
// doing: horizontal padding, plus a table head the print engine lays out again
// on every page. Padding cannot do the vertical band, because it applies once to
// the element rather than once per sheet.
//
// The fourth edge, the foot, does carry a margin, for one reason:
// a page counter can live nowhere but an @page margin box; a margin box needs
// a margin to sit in. Declaring one is also what stops the browser filling
// that margin with its own date, address and page count.
//
// Take any of these away and the record still looks right in the window, still
// passes every other test in this suite and comes off the printer with its first
// line sliced through or its pages unnumbered. That is what this file stops.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	// printSheet is where the printed record's rules live.
	printSheet = filepath.Join("frontend", "src", "receipt.css")
	// printPane is the page that draws the sheet.
	printPane = filepath.Join("frontend", "src", "ReceiptPane.tsx")
)

// printNamed are the measurements the printed sheet is built from. They are
// named rather than written into the rules because more than one rule reads
// each; together they make up the whole of the space between the record
// and the edge of the paper.
var printNamed = []string{
	"--print-gutter", "--print-side", "--print-foot", "--print-foot-gap",
}

var (
	printBlock  = regexp.MustCompile(`(?s)@media print \{.*\n\}`)
	pageMargin  = regexp.MustCompile(`@page\s*\{[\s\S]*?margin:\s*0\s+0\s+var\(--print-foot\)\s*;`)
	pageNumber  = regexp.MustCompile(`@bottom-center\s*\{[\s\S]*?content:[^;]*counter\(page\)[^;]*counter\(pages\)`)
	topGutter   = regexp.MustCompile(`thead \.gutter\s*\{[^}]*height:\s*var\(--print-gutter\)`)
	footGutter  = regexp.MustCompile(`tfoot \.gutter\s*\{[^}]*height:\s*var\(--print-foot-gap\)`)
	sidePadding = regexp.MustCompile(`padding:\s*0\s+var\(--print-side\)`)
)

// read answers one of the sheet's files, failing the run rather than the check
// when it is not there: a renamed file is not a passing test.
func read(t *testing.T, relative string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), relative))
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(relative), err)
	}
	return string(raw)
}

func TestThePrintedSheetKeepsTheRecordOffTheEdgeOfThePaper(t *testing.T) {
	sheet := read(t, printSheet)

	for _, named := range printNamed {
		if !strings.Contains(sheet, named+":") {
			t.Errorf("%s defines no %s", filepath.ToSlash(printSheet), named)
		}
	}

	block := printBlock.FindString(sheet)
	if block == "" {
		t.Fatalf("%s has no @media print block", filepath.ToSlash(printSheet))
	}

	if !pageMargin.MatchString(block) {
		t.Error("the print block does not set `margin: 0 0 var(--print-foot)` on " +
			"@page: a margin at the top or the sides is room the browser prints " +
			"its own header and footer in; a page can reach them no other way")
	}
	if !topGutter.MatchString(block) {
		t.Error("the print block gives the thead gutter no height from " +
			"--print-gutter: with no page margin at the top, that repeating row is " +
			"the only thing between the record and the edge of every sheet")
	}
	if !footGutter.MatchString(block) {
		t.Error("the print block gives the tfoot gutter no height from " +
			"--print-foot-gap: without it the last line of the record sits against " +
			"the page number and the two read as one")
	}
	if !sidePadding.MatchString(block) {
		t.Error("the print block sets no `padding: 0 var(--print-side)` on the " +
			"receipt: with no page margin, nothing else holds the record off the " +
			"left and right edges")
	}
}

func TestEveryPrintedPageSaysWhichOneItIs(t *testing.T) {
	block := printBlock.FindString(read(t, printSheet))
	if block == "" {
		t.Fatalf("%s has no @media print block", filepath.ToSlash(printSheet))
	}

	// A record handed across a desk can be dropped; a symptom record read in
	// the wrong order is worse than one that is hard to read, so every sheet says
	// where it belongs and how many there are.
	if !pageNumber.MatchString(block) {
		t.Error("the print block has no @bottom-center numbering the pages from " +
			"counter(page) and counter(pages): a page counter can live nowhere " +
			"else; declaring that box is also what keeps the browser's own " +
			"date, address and page count off the paper")
	}
}

func TestTheSheetIsDrawnWithARepeatingHeadAndFoot(t *testing.T) {
	pane := read(t, printPane)

	// A gutter row in the body rather than the head would be laid out once, which
	// is the failure this whole arrangement exists to avoid, so each is asked for
	// by name and in its own place.
	for _, part := range []string{"<thead>", "<tfoot>", `className="gutter"`} {
		if !strings.Contains(pane, part) {
			t.Errorf("%s draws no %s: the printed band at the top of every page "+
				"after the first comes from nothing else",
				filepath.ToSlash(printPane), part)
		}
	}
	if strings.Count(pane, `className="gutter"`) != 2 {
		t.Errorf("%s draws %d gutter rows, want one in the head and one in the foot",
			filepath.ToSlash(printPane), strings.Count(pane, `className="gutter"`))
	}
}
