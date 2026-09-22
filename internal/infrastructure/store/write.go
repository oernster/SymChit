package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/oernster/symchit/internal/application"
	"github.com/oernster/symchit/internal/domain"
)

// AddEvent writes a new event and answers it as stored.
func (s *Store) AddEvent(input application.EventInput) (domain.Event, error) {
	var saved domain.Event
	err := s.inTransaction(func(tx *sql.Tx) error {
		id, err := insertEvent(tx, input)
		if err != nil {
			return err
		}
		saved, err = eventIn(tx, id)
		return err
	})
	return saved, err
}

// UpdateEvent replaces an event's symptom, occurrence, severity and note. The
// statement names no recorded_at column, so recorded-at cannot change (FR-007).
func (s *Store) UpdateEvent(id domain.EventID, input application.EventInput) (domain.Event, error) {
	var saved domain.Event
	err := s.inTransaction(func(tx *sql.Tx) error {
		held, err := eventIn(tx, id)
		if err != nil {
			return err
		}
		definition, err := resolve(tx, input.Label, input.Key, held.RecordedAt)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE events SET definition_id = ?, occurred_at = ?,
			severity = ?, note = ? WHERE id = ?`,
			definition, input.OccurredAt.Format(timeLayout),
			input.Severity.String(), input.Note, id); err != nil {
			return err
		}
		saved, err = eventIn(tx, id)
		return err
	})
	return saved, err
}

// DeleteEvents removes every listed event, else none: an id that is not held
// rolls the whole deletion back.
func (s *Store) DeleteEvents(ids []domain.EventID) error {
	return s.inTransaction(func(tx *sql.Tx) error {
		for _, id := range ids {
			result, err := tx.Exec(`DELETE FROM events WHERE id = ?`, id)
			if err != nil {
				return err
			}
			if removed, err := result.RowsAffected(); err != nil || removed == 0 {
				return errors.Join(domain.ErrNoSuchEvent, err)
			}
		}
		return nil
	})
}

// RenameDefinition gives a definition a new label and key.
func (s *Store) RenameDefinition(id domain.DefinitionID, label string, key domain.SymptomKey) error {
	return s.inTransaction(func(tx *sql.Tx) error {
		result, err := tx.Exec(`UPDATE definitions SET label = ?, key = ? WHERE id = ?`,
			label, string(key), id)
		if err != nil {
			return err
		}
		if renamed, err := result.RowsAffected(); err != nil || renamed == 0 {
			return errors.Join(domain.ErrNoSuchDefinition, err)
		}
		return nil
	})
}

// Import adds the definitions not yet held and then the events, in one
// transaction.
func (s *Store) Import(definitions []application.DefinitionInput, events []application.EventInput) error {
	return s.inTransaction(func(tx *sql.Tx) error {
		for _, definition := range definitions {
			if _, err := resolve(tx, definition.Label, definition.Key, definition.LastUsed); err != nil {
				return err
			}
		}
		for _, input := range events {
			if _, err := insertEvent(tx, input); err != nil {
				return err
			}
		}
		return nil
	})
}

// insertEvent writes one event inside a transaction and answers its id.
func insertEvent(tx *sql.Tx, input application.EventInput) (domain.EventID, error) {
	definition, err := resolve(tx, input.Label, input.Key, input.RecordedAt)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO events
		(definition_id, occurred_at, recorded_at, severity, note) VALUES (?, ?, ?, ?, ?)`,
		definition, input.OccurredAt.Format(timeLayout), input.RecordedAt.Format(timeLayout),
		input.Severity.String(), input.Note)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return domain.EventID(id), err
}

// resolve answers the definition holding key, creating it with label when there
// is none; it moves its last use forward to used when that is later.
func resolve(tx *sql.Tx, label string, key domain.SymptomKey, used time.Time) (domain.DefinitionID, error) {
	var id domain.DefinitionID
	var held string
	err := tx.QueryRow(`SELECT id, last_used FROM definitions WHERE key = ?`, string(key)).
		Scan(&id, &held)
	if errors.Is(err, sql.ErrNoRows) {
		result, err := tx.Exec(`INSERT INTO definitions (label, key, last_used) VALUES (?, ?, ?)`,
			label, string(key), used.Format(timeLayout))
		if err != nil {
			return 0, err
		}
		created, err := result.LastInsertId()
		return domain.DefinitionID(created), err
	}
	if err != nil {
		return 0, err
	}
	previous, err := time.Parse(timeLayout, held)
	if err != nil {
		return 0, err
	}
	if used.After(previous) {
		if _, err := tx.Exec(`UPDATE definitions SET last_used = ? WHERE id = ?`,
			used.Format(timeLayout), id); err != nil {
			return 0, err
		}
	}
	return id, nil
}
