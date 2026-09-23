package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// timeLayout is how an instant is kept: RFC 3339 with its offset, to the
// nanosecond, so it reads back as the same instant.
const timeLayout = time.RFC3339Nano

// selectEvents reads events with their definitions' labels.
const selectEvents = `SELECT e.id, e.definition_id, d.label, e.occurred_at,
	e.recorded_at, e.severity, e.note
	FROM events e JOIN definitions d ON d.id = e.definition_id`

// Definitions answers every definition held.
func (s *Store) Definitions() ([]domain.Definition, error) {
	rows, err := s.db.Query(`SELECT id, label, last_used FROM definitions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Definition
	for rows.Next() {
		var definition domain.Definition
		var used string
		if err := rows.Scan(&definition.ID, &definition.Label, &used); err != nil {
			return nil, err
		}
		if definition.LastUsed, err = time.Parse(timeLayout, used); err != nil {
			return nil, fmt.Errorf("definition %d: %w", definition.ID, err)
		}
		out = append(out, definition)
	}
	return out, rows.Err()
}

// Events answers every event held.
func (s *Store) Events() ([]domain.Event, error) {
	rows, err := s.db.Query(selectEvents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

// Event answers one event; domain.ErrNoSuchEvent when it is not held.
func (s *Store) Event(id domain.EventID) (domain.Event, error) {
	return eventIn(s.db, id)
}

// querier is what a read needs, satisfied by the database and a transaction.
type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// eventIn reads one event through a querier.
func eventIn(q querier, id domain.EventID) (domain.Event, error) {
	event, err := scanEvent(q.QueryRow(selectEvents+` WHERE e.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Event{}, domain.ErrNoSuchEvent
	}
	return event, err
}

// scanner is a row or a cursor over rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanEvent reads one event from a row of selectEvents.
func scanEvent(row scanner) (domain.Event, error) {
	var event domain.Event
	var occurred, recorded, severity string
	if err := row.Scan(&event.ID, &event.Definition, &event.Symptom,
		&occurred, &recorded, &severity, &event.Note); err != nil {
		return domain.Event{}, err
	}
	var err error
	if event.OccurredAt, err = time.Parse(timeLayout, occurred); err != nil {
		return domain.Event{}, fmt.Errorf("event %d: %w", event.ID, err)
	}
	if event.RecordedAt, err = time.Parse(timeLayout, recorded); err != nil {
		return domain.Event{}, fmt.Errorf("event %d: %w", event.ID, err)
	}
	if event.Severity, err = domain.ParseSeverity(severity); err != nil {
		return domain.Event{}, fmt.Errorf("event %d: %w", event.ID, err)
	}
	return event, nil
}
