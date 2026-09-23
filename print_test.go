package main

import "testing"

// countingPrinter stands in for the window, counting the print dialogs it was
// asked for so a test never opens a real one.
type countingPrinter struct{ asked int }

// Print counts the request instead of opening a dialog.
func (p *countingPrinter) Print() { p.asked++ }

func TestPrintAsksTheWindowForItsPrintDialog(t *testing.T) {
	t.Parallel()
	app, _, _ := facade(t)
	printer := &countingPrinter{}
	app.printer = printer

	if err := app.Print(); err != nil {
		t.Fatalf("Print = %v, want no error", err)
	}
	if printer.asked != 1 {
		t.Errorf("the window was asked for %d print dialogs, want exactly 1", printer.asked)
	}
}

func TestPrintOpensNothingWhenTheWindowIsMissing(t *testing.T) {
	t.Parallel()
	// A facade built without a printer is a wiring mistake rather than a state
	// the user can reach, so it must answer the same internal fault every other
	// bound method answers, not end the window.
	app, _, _ := facade(t)
	app.printer = nil
	if err := app.Print(); err == nil {
		t.Error("Print with no printer answered no error, want the internal fault")
	}
}
