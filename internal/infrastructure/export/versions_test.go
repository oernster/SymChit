package export

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// samplePattern names the frozen sample of one format version. Each is the
// bytes that version really wrote, committed and never edited again: a reader
// tested against the current struct is only tested against itself.
const samplePattern = "version-%d.json"

// TestEveryVersionEverWrittenCanStillBeRead is the export's whole promise. A
// user's export is their copy of their own record, so a SymDiary that cannot
// open a file it wrote two years ago has taken their record off them.
//
// Proved by planting: raising formatVersion to 2 without adding a reader fails
// by name; so does adding the reader without committing the sample.
func TestEveryVersionEverWrittenCanStillBeRead(t *testing.T) {
	t.Parallel()
	for version := firstVersion; version <= formatVersion; version++ {
		if _, known := readers[version]; !known {
			t.Errorf("format %d can be written but not read: it wants a reader in versions.go", version)
			continue
		}
		path := filepath.Join("testdata", fmt.Sprintf(samplePattern, version))
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("format %d has a reader and no sample: commit %s, the bytes that version wrote",
				version, filepath.ToSlash(path))
			continue
		}
		record, err := readVersioned([]byte(strings.ReplaceAll(string(raw), "\r\n", "\n")))
		if err != nil {
			t.Errorf("format %d no longer reads its own sample: %v", version, err)
			continue
		}
		if len(record.Events) == 0 || len(record.Definitions) == 0 {
			t.Errorf("format %d read its sample as an empty record, so the sample proves nothing", version)
		}
	}
}

// TestNoReaderClaimsAVersionThatWasNeverWritten keeps the table honest in the
// other direction: a reader for a version no SymDiary ever wrote would accept a
// file nothing produced.
func TestNoReaderClaimsAVersionThatWasNeverWritten(t *testing.T) {
	t.Parallel()
	for version := range readers {
		if version < firstVersion || version > formatVersion {
			t.Errorf("there is a reader for format %d, which no SymDiary ever wrote", version)
		}
	}
}

// TestAFileNamingNoVersionIsRefused closes the hole that read an unversioned
// file as though it were the current format. A file with the right marker and
// no version is not an export SymDiary wrote, whatever else is in it; a record
// is not the thing to be generous about.
func TestAFileNamingNoVersionIsRefused(t *testing.T) {
	t.Parallel()
	for _, declared := range []string{"", `"version": 0,`, `"version": -1,`} {
		body := fmt.Sprintf(`{"format": %q, %s "definitions": [], "events": []}`,
			formatName, declared)
		if _, err := readVersioned([]byte(body)); !errors.Is(err, ErrNoVersion) {
			t.Errorf("a file declaring %q answered %v, want ErrNoVersion", declared, err)
		}
	}
}

// TestAGapInTheTableRefusesRatherThanGuesses proves the unreachable arm of
// readVersioned does what it says. The table is emptied for the length of the
// test, which is the only way to reach a version that is writable and has no
// reader while the guard above holds.
// It restores the table by putting back exactly what was there, absence
// included. An earlier form wrote the key back unconditionally, which left a
// nil reader behind for a version that had none; the next test then read the
// table as though that version were covered and reported the wrong reason for a
// failure it was right about.
func TestAGapInTheTableRefusesRatherThanGuesses(t *testing.T) {
	held, existed := readers[formatVersion]
	delete(readers, formatVersion)
	defer func() {
		if existed {
			readers[formatVersion] = held
		}
	}()

	body := fmt.Sprintf(`{"format": %q, "version": %d, "definitions": [], "events": []}`,
		formatName, formatVersion)
	if _, err := readVersioned([]byte(body)); !errors.Is(err, ErrNoVersion) {
		t.Errorf("a version with no reader answered %v, want a refusal", err)
	}
}
