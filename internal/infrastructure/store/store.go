// Package store keeps the record in one SQLite file (FR-060).
//
// Every write is one transaction. The file runs in WAL mode with synchronous
// FULL, so an event reported as recorded survives a power loss (NFR-REL-001).
// The store never logs: nothing it holds may reach the log (FR-065).
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	// The pure-Go SQLite driver, registered as "sqlite".
	_ "modernc.org/sqlite"
)

// driverName is the name modernc.org/sqlite registers itself under.
const driverName = "sqlite"

// busyTimeoutMillis is how long a statement waits for a lock held by another
// connection before failing. SymDiary runs one instance with one connection, so
// the wait covers only an external tool holding the file open.
const busyTimeoutMillis = 5000

// dsnLayout opens the file with the durability pragmas applied to the
// connection.
const dsnLayout = "file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)" +
	"&_pragma=foreign_keys(1)&_pragma=busy_timeout(%d)"

// directoryMode is the permission a missing parent directory is created with:
// the user alone.
const directoryMode = 0o700

// ErrNewerSchema refuses a file written by a later SymDiary, which this one
// would not know how to keep.
var ErrNewerSchema = errors.New("the record was written by a newer SymDiary")

// Store is the SQLite record.
type Store struct {
	db *sql.DB
}

// Open opens the record at path, creating it and its directory when absent
// (FR-061) and upgrading an older schema (FR-063). A file that cannot be read is
// reported and left exactly as it was (FR-062).
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), directoryMode); err != nil {
		return nil, fmt.Errorf("creating the folder for %s: %w", path, err)
	}
	db, err := sql.Open(driverName, fmt.Sprintf(dsnLayout, filepath.ToSlash(path), busyTimeoutMillis))
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	// One connection: SymDiary is the single writer; a pragma set on one
	// connection is not set on another.
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return store, nil
}

// Close closes the file.
func (s *Store) Close() error { return s.db.Close() }

// migrations holds each schema version's statements, oldest first. Version N is
// migrations[N-1]; the file's user_version says how many have been applied. A
// migration is never edited once released; a change is a new entry.
var migrations = []string{
	`CREATE TABLE definitions (
		id        INTEGER PRIMARY KEY,
		label     TEXT NOT NULL,
		key       TEXT NOT NULL UNIQUE,
		last_used TEXT NOT NULL
	);
	CREATE TABLE events (
		id            INTEGER PRIMARY KEY,
		kind          TEXT NOT NULL DEFAULT 'symptom',
		definition_id INTEGER NOT NULL REFERENCES definitions(id),
		occurred_at   TEXT NOT NULL,
		recorded_at   TEXT NOT NULL,
		severity      TEXT NOT NULL,
		note          TEXT NOT NULL
	);
	CREATE INDEX events_by_definition ON events(definition_id);`,
}

// migrate applies every migration the file has not had, each in its own
// transaction together with the version it reaches.
func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > len(migrations) {
		return fmt.Errorf("%w (schema %d, this one knows %d)", ErrNewerSchema, version, len(migrations))
	}
	for next := version; next < len(migrations); next++ {
		if err := s.inTransaction(func(tx *sql.Tx) error {
			if _, err := tx.Exec(migrations[next]); err != nil {
				return err
			}
			_, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, next+1))
			return err
		}); err != nil {
			return fmt.Errorf("upgrading to schema %d: %w", next+1, err)
		}
	}
	return nil
}

// inTransaction runs work in one transaction, committing only when it succeeds.
func (s *Store) inTransaction(work func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
