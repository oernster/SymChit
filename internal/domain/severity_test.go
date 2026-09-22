package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestSeverityIsOptional(t *testing.T) {
	t.Parallel()
	var unset Severity
	if unset != SeverityNone {
		t.Error("the zero severity should be SeverityNone")
	}
	if slices.Contains(Severities(), SeverityNone) {
		t.Error("SeverityNone is the absence of a choice, not a choice")
	}
	parsed, err := ParseSeverity("")
	if err != nil || parsed != SeverityNone {
		t.Errorf("ParseSeverity(\"\") = %v, %v; want SeverityNone", parsed, err)
	}
}

func TestSeverityListAndNames(t *testing.T) {
	t.Parallel()
	var names []string
	for _, severity := range Severities() {
		names = append(names, severity.String())
		parsed, err := ParseSeverity(severity.String())
		if err != nil || parsed != severity {
			t.Errorf("%s does not round-trip: %v, %v", severity, parsed, err)
		}
	}
	if want := []string{"Mild", "Moderate", "Severe"}; !slices.Equal(names, want) {
		t.Errorf("choices = %v, want %v (Q-1)", names, want)
	}
}

func TestUnknownSeverityRefused(t *testing.T) {
	t.Parallel()
	if _, err := ParseSeverity("mild"); !errors.Is(err, ErrUnknownSeverity) {
		t.Errorf("ParseSeverity(\"mild\") = %v, want ErrUnknownSeverity", err)
	}
	if got := Severity(9).String(); got != "Severity(9)" {
		t.Errorf("an out-of-range severity reads %q", got)
	}
	if got := Severity(-1).String(); got != "Severity(-1)" {
		t.Errorf("a negative severity reads %q", got)
	}
}
