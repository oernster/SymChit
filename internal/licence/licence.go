// Package licence carries the licence SymChit is given under, both as the
// published text and as a short reading of what it means for the person
// installing it.
//
// Why the text sits here as well as at the repository root: Go's embedding
// cannot reach above the package it is written in; the setup program is a
// separate main package, so neither binary can embed the root LICENSE
// directly. A second copy is the price of showing the licence offline. It is
// not left to trust: tests/structural/licence_test.go compares the two byte
// for byte, so the two cannot drift apart and a change to one that misses the
// other fails the gate rather than shipping.
//
// The plain reading below is not a substitute for the text and does not try to
// be. It is there because "GNU General Public Licence, version 3" tells most
// people nothing at all; somebody about to install a program deserves to
// know what they are being given in words they already use.
package licence

import (
	_ "embed"
	"strings"
)

// Name is the licence as it is properly called.
const Name = "GNU General Public Licence, version 3"

// Holder is who the copyright belongs to.
const Holder = "Oliver Ernster"

//go:embed LICENSE
var text string

// Text answers the licence in full, exactly as published.
func Text() string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

// Plainly answers what the licence means for the person reading it, one
// statement at a time.
//
// Each line says a thing they can do rather than a thing the licence says,
// because the question somebody asks at a setup screen is "what am I allowed
// to do with this", never "what clause governs it". The last line is the one
// that makes it a licence rather than a gift, so it is stated as plainly as
// the rest instead of being softened.
func Plainly() []string {
	return []string{
		"You can use it for anything you like, for as long as you like, " +
			"without paying for it.",
		"You can give copies to anybody: a friend, a relative, a waiting room.",
		"You can read every line of the source. You can change it too, if " +
			"you can program; if not, somebody else can do it for you.",
		"If you pass on a changed version, it has to come with its source and " +
			"carry this same licence, so whoever gets it next has everything " +
			"you had.",
		"It comes with no warranty. Nobody is promising it will suit your " +
			"purpose; nobody is liable if it does not.",
	}
}
