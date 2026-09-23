package structural

// What keeps the printed record off the edge of the paper.
//
// The sheet prints with no page margin at all, which is the only way to leave
// the browser nowhere to draw its own header and footer (FR-045, Amendment 14).
// Everything between the record and the edge of the sheet is then the page's own
// doing: horizontal padding, plus a table head and foot that the print engine
// lays out again on every page. Padding cannot do the vertical band, because it
// applies once to the element rather than once per sheet.
//
// Take any one of those three away and the record still looks right in the
// window, still passes every other test in this suite and comes off the printer
// with its first line sliced through. That is what this file is here to stop.

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

var (
	printBlock  = regexp.MustCompile(`(?s)@media print \{.*\n\}`)
	pageMargin  = regexp.MustCompile(`@page\s*\{[^}]*margin:\s*0\s*[;}]`)
	gutterSize  = regexp.MustCompile(`\.gutter\s*\{[^}]*height:\s*var\(--print-gutter\)`)
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

	for _, named := range []string{"--print-gutter", "--print-side"} {
		if !strings.Contains(sheet, named+":") {
			t.Errorf("%s defines no %s: the two measurements the whole page margin "+
				"comes from are named so that the rules reading them cannot drift",
				filepath.ToSlash(printSheet), named)
		}
	}

	block := printBlock.FindString(sheet)
	if block == "" {
		t.Fatalf("%s has no @media print block", filepath.ToSlash(printSheet))
	}

	if !pageMargin.MatchString(block) {
		t.Error("the print block sets no `@page { margin: 0 }`: with a page margin " +
			"the browser prints its own header and footer inside it; a page has " +
			"no other way to reach them")
	}
	if !gutterSize.MatchString(block) {
		t.Error("the print block gives .gutter no height from --print-gutter: with " +
			"no page margin, those repeating rows are the only thing between the " +
			"record and the top and bottom edges of every sheet")
	}
	if !sidePadding.MatchString(block) {
		t.Error("the print block sets no `padding: 0 var(--print-side)` on the " +
			"receipt: with no page margin, nothing else holds the record off the " +
			"left and right edges")
	}
}

func TestTheSheetIsDrawnWithARepeatingHeadAndFoot(t *testing.T) {
	pane := read(t, printPane)

	// A gutter row in the body rather than the head would be laid out once, which
	// is the failure this whole arrangement exists to avoid, so each is asked for
	// by name and in its own place.
	for _, part := range []string{"<thead>", "<tfoot>", `className="gutter"`} {
		if !strings.Contains(pane, part) {
			t.Errorf("%s draws no %s: the printed band at the top and foot of every "+
				"page after the first comes from nothing else",
				filepath.ToSlash(printPane), part)
		}
	}
	if strings.Count(pane, `className="gutter"`) != 2 {
		t.Errorf("%s draws %d gutter rows, want one in the head and one in the foot",
			filepath.ToSlash(printPane), strings.Count(pane, `className="gutter"`))
	}
}
