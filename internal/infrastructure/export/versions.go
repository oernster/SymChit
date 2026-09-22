package export

// The export's version contract.
//
// The file a user exports is their own copy of their record, so SymChit has to
// go on reading it after the format has moved on. That is a promise about every
// version ever written, not only the current one, so it is stated as a table
// rather than left to whatever struct the code happens to hold today.
//
// Adding a version means three things, all of which the guard in
// versions_test.go refuses the change until you have done: raise formatVersion,
// add a reader here, then commit a sample of the OLD version under testdata so
// its reader is tested against the bytes that were really written rather than
// against a struct that has since changed.
//
// When version 2 arrives, the struct now called fileShape is version 1's shape.
// Copy it to a v1Shape, point readVersion1 at that, then let fileShape move on
// with Write. There is no second struct today because there is no second
// version; inventing one now would be a copy with nothing to say.

import (
	"encoding/json"

	"github.com/oernster/symchit/internal/application"
)

// firstVersion is the earliest version that ever existed. A file declaring
// anything below it, the absent version included, was not written by SymChit.
const firstVersion = 1

// readers answers each format version this SymChit can read. A version with no
// entry here cannot be read, whatever formatVersion says.
var readers = map[int]func([]byte) (application.Record, error){
	1: readVersion1,
}

// envelope is the part of the file that is the same in every version: what it
// is and which version it is. It is read first, on its own, so the version
// decides which reader unmarshals the rest. Reading the whole file into one
// struct and hoping it fits is exactly what a versioned format is for avoiding.
type envelope struct {
	Format  string `json:"format"`
	Version int    `json:"version"`
}

// readVersion1 reads the first format: a flat list of definitions and a flat
// list of events, each time an RFC 3339 instant carrying its offset.
func readVersion1(raw []byte) (application.Record, error) {
	var shape fileShape
	if err := json.Unmarshal(raw, &shape); err != nil {
		return application.Record{}, ErrNotAnExport
	}
	return shape.record()
}
