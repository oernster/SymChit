package domain

import (
	"errors"
	"testing"
	"time"
)

// definitions answers Tired, Headache and Back pain, Headache used most recently.
func definitions(t *testing.T) []Definition {
	t.Helper()
	return []Definition{
		{ID: 1, Label: "Tired", LastUsed: at(t, 2026, time.September, 20, 9, 0)},
		{ID: 2, Label: "Headache", LastUsed: at(t, 2026, time.September, 22, 9, 0)},
		{ID: 3, Label: "Back pain", LastUsed: at(t, 2026, time.September, 20, 9, 0)},
	}
}

// labels answers the labels of some definitions, in order.
func labels(found []Definition) []string {
	out := make([]string, 0, len(found))
	for _, definition := range found {
		out = append(out, definition.Label)
	}
	return out
}

func TestSuggestionsMatchIgnoringCase(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"he":  {"Headache"},
		"HE":  {"Headache"},
		"a":   {"Headache", "Back pain"},
		"":    {"Headache", "Tired", "Back pain"},
		"zzz": {},
	}
	for typed, want := range cases {
		got := labels(Suggest(definitions(t), typed))
		if len(got) != len(want) {
			t.Errorf("Suggest(%q) = %v, want %v", typed, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("Suggest(%q) = %v, want %v", typed, got, want)
				break
			}
		}
	}
}

func TestCaseVariantReusesDefinition(t *testing.T) {
	t.Parallel()
	found, ok := FindDefinition(definitions(t), "  tired ")
	if !ok || found.ID != 1 || found.Label != "Tired" {
		t.Errorf("FindDefinition(\"  tired \") = %+v, %v; want Tired as first typed", found, ok)
	}
	if _, ok := FindDefinition(definitions(t), "Nausea"); ok {
		t.Error("an unknown symptom should not be found")
	}
}

func TestRenameChecks(t *testing.T) {
	t.Parallel()
	held := definitions(t)
	if err := CheckRename(held, 1, "Exhausted"); err != nil {
		t.Errorf("a fresh name = %v, want nil", err)
	}
	if err := CheckRename(held, 1, "TIRED"); err != nil {
		t.Errorf("a case variant of its own label = %v, want nil", err)
	}
	if err := CheckRename(held, 1, "headache"); !errors.Is(err, ErrDefinitionExists) {
		t.Errorf("onto another label = %v, want ErrDefinitionExists", err)
	}
	if err := CheckRename(held, 1, " "); !errors.Is(err, ErrBlankSymptom) {
		t.Errorf("blank = %v, want ErrBlankSymptom", err)
	}
	if err := CheckRename(held, 99, "Nausea"); !errors.Is(err, ErrNoSuchDefinition) {
		t.Errorf("unknown definition = %v, want ErrNoSuchDefinition", err)
	}
}
