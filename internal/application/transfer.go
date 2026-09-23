package application

import "github.com/oernster/symdiary/internal/domain"

// Transfer moves the whole record out to a file and back in (FR-050, FR-053).
type Transfer struct {
	store Store
	file  RecordFile
}

// NewTransfer answers a transfer over a store and an export file format.
func NewTransfer(store Store, file RecordFile) Transfer {
	return Transfer{store: store, file: file}
}

// ImportResult says what an import did.
type ImportResult struct {
	Added   int
	Skipped int
}

// Export writes every definition and every event to path.
func (t Transfer) Export(path string) error {
	definitions, err := t.store.Definitions()
	if err != nil {
		return because(ErrNotExported, err)
	}
	events, err := t.store.Events()
	if err != nil {
		return because(ErrNotExported, err)
	}
	record := Record{Definitions: definitions, Events: events}
	if err := t.file.Write(path, record); err != nil {
		return because(ErrNotExported, err)
	}
	return nil
}

// Import reads path and adds every event the record does not already hold,
// skipping repeats within the file as well. Definitions come too, so a symptom
// with no events is not lost; a label already held keeps its held form.
func (t Transfer) Import(path string) (ImportResult, error) {
	incoming, err := t.file.Read(path)
	if err != nil {
		return ImportResult{}, because(ErrNotImported, err)
	}
	held, err := t.store.Events()
	if err != nil {
		return ImportResult{}, because(ErrNotImported, err)
	}
	seen := make(map[domain.Observation]bool, len(held)+len(incoming.Events))
	for _, event := range held {
		seen[domain.ObservationOf(event)] = true
	}
	var result ImportResult
	var events []EventInput
	for _, event := range incoming.Events {
		observation := domain.ObservationOf(event)
		if seen[observation] {
			result.Skipped++
			continue
		}
		seen[observation] = true
		events = append(events, EventInput{
			Label: event.Symptom, Key: observation.Symptom,
			OccurredAt: event.OccurredAt, RecordedAt: event.RecordedAt,
			Severity: event.Severity, Note: event.Note,
		})
	}
	definitions := make([]DefinitionInput, 0, len(incoming.Definitions))
	for _, definition := range incoming.Definitions {
		definitions = append(definitions, DefinitionInput{
			Label: definition.Label, Key: domain.KeyOf(definition.Label),
			LastUsed: definition.LastUsed,
		})
	}
	if err := t.store.Import(definitions, events); err != nil {
		return ImportResult{}, because(ErrNotImported, err)
	}
	result.Added = len(events)
	return result, nil
}
