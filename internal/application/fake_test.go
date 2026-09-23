package application

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// errBroken is the failure the fakes inject.
var errBroken = errors.New("disk on fire")

// fixedClock is a clock that always reads the same instant.
type fixedClock struct{ now time.Time }

// Now answers the fixed instant.
func (c fixedClock) Now() time.Time { return c.now }

// fakeStore keeps the record in memory, following the Store contract, with a
// switch per operation to make it fail.
type fakeStore struct {
	definitions []domain.Definition
	events      []domain.Event
	nextID      int64

	failReads, failEvent, failAdd, failUpdate bool
	failDelete, failRename, failImport        bool
}

// newFakeStore answers an empty store.
func newFakeStore() *fakeStore { return &fakeStore{nextID: 1} }

func (s *fakeStore) id() int64 {
	s.nextID++
	return s.nextID
}

func (s *fakeStore) Definitions() ([]domain.Definition, error) {
	if s.failReads {
		return nil, errBroken
	}
	return slices.Clone(s.definitions), nil
}

func (s *fakeStore) Events() ([]domain.Event, error) {
	if s.failReads {
		return nil, errBroken
	}
	out := slices.Clone(s.events)
	for i := range out {
		out[i].Symptom = s.labelOf(out[i].Definition)
	}
	return out, nil
}

func (s *fakeStore) Event(id domain.EventID) (domain.Event, error) {
	if s.failEvent {
		return domain.Event{}, errBroken
	}
	for _, event := range s.events {
		if event.ID == id {
			event.Symptom = s.labelOf(event.Definition)
			return event, nil
		}
	}
	return domain.Event{}, domain.ErrNoSuchEvent
}

func (s *fakeStore) labelOf(id domain.DefinitionID) string {
	for _, definition := range s.definitions {
		if definition.ID == id {
			return definition.Label
		}
	}
	return ""
}

// resolve finds or creates the definition for a key, as the real store does.
func (s *fakeStore) resolve(label string, key domain.SymptomKey, used time.Time) domain.DefinitionID {
	for i, definition := range s.definitions {
		if domain.KeyOf(definition.Label) == key {
			if used.After(definition.LastUsed) {
				s.definitions[i].LastUsed = used
			}
			return definition.ID
		}
	}
	id := domain.DefinitionID(s.id())
	s.definitions = append(s.definitions, domain.Definition{ID: id, Label: label, LastUsed: used})
	return id
}

func (s *fakeStore) AddEvent(input EventInput) (domain.Event, error) {
	if s.failAdd {
		return domain.Event{}, errBroken
	}
	event := domain.Event{
		ID:         domain.EventID(s.id()),
		Definition: s.resolve(input.Label, input.Key, input.RecordedAt),
		OccurredAt: input.OccurredAt, RecordedAt: input.RecordedAt,
		Severity: input.Severity, Note: input.Note,
	}
	s.events = append(s.events, event)
	return s.Event(event.ID)
}

func (s *fakeStore) UpdateEvent(id domain.EventID, input EventInput) (domain.Event, error) {
	if s.failUpdate {
		return domain.Event{}, errBroken
	}
	for i, event := range s.events {
		if event.ID == id {
			s.events[i].Definition = s.resolve(input.Label, input.Key, event.RecordedAt)
			s.events[i].OccurredAt = input.OccurredAt
			s.events[i].Severity = input.Severity
			s.events[i].Note = input.Note
			return s.Event(id)
		}
	}
	return domain.Event{}, domain.ErrNoSuchEvent
}

func (s *fakeStore) DeleteEvents(ids []domain.EventID) error {
	if s.failDelete {
		return errBroken
	}
	s.events = slices.DeleteFunc(s.events, func(e domain.Event) bool {
		return slices.Contains(ids, e.ID)
	})
	return nil
}

func (s *fakeStore) RenameDefinition(id domain.DefinitionID, label string, _ domain.SymptomKey) error {
	if s.failRename {
		return errBroken
	}
	for i := range s.definitions {
		if s.definitions[i].ID == id {
			s.definitions[i].Label = label
		}
	}
	return nil
}

func (s *fakeStore) Import(definitions []DefinitionInput, events []EventInput) error {
	if s.failImport {
		return errBroken
	}
	for _, definition := range definitions {
		s.resolve(definition.Label, definition.Key, definition.LastUsed)
	}
	for _, input := range events {
		if _, err := s.AddEvent(input); err != nil {
			return err
		}
	}
	return nil
}

// fakeFile is an export file held in memory.
type fakeFile struct {
	written   map[string]Record
	failWrite bool
	failRead  bool
}

func (f *fakeFile) Write(path string, record Record) error {
	if f.failWrite {
		return errBroken
	}
	f.written[path] = record
	return nil
}

func (f *fakeFile) Read(path string) (Record, error) {
	if f.failRead {
		return Record{}, errBroken
	}
	record, ok := f.written[path]
	if !ok {
		return Record{}, errBroken
	}
	return record, nil
}

// london is the user's zone in every test.
func london(t *testing.T) *time.Location {
	t.Helper()
	zone, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("loading Europe/London: %v", err)
	}
	return zone
}

// at answers a London wall-clock instant on 22 September 2026.
func at(t *testing.T, hour, minute int) time.Time {
	t.Helper()
	return time.Date(2026, time.September, 22, hour, minute, 0, 0, london(t))
}

// wantIs fails unless err wraps target.
func wantIs(t *testing.T, what string, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s = %v, want %v", what, err, target)
	}
}
