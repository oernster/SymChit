package application

import "github.com/oernster/symchit/internal/product"

// Credit names one open-source work SymChit is built with, with its licence as
// its own LICENSE file states it.
type Credit struct {
	Work    string
	Licence string
	Holder  string
}

// About is what the About dialog shows (FR-066, FR-067).
type About struct {
	Name      string
	Version   string
	Author    string
	Copyright string
	Statement string
	Credits   []Credit
}

// credits lists the works that ship inside SymChit. Build tools that ship
// nothing (Vite, TypeScript, the test runners) are not credited. Each licence
// was read from the work's own LICENSE file on 2026-09-22.
var credits = []Credit{
	{Work: "Go and golang.org/x/sys", Licence: "BSD 3-Clause", Holder: "The Go Authors"},
	{Work: "Wails", Licence: "MIT", Holder: "Lea Anthony"},
	{Work: "modernc.org/sqlite", Licence: "BSD 3-Clause", Holder: "The Sqlite Authors"},
	{Work: "SQLite", Licence: "Public domain", Holder: "D. Richard Hipp and contributors"},
	{Work: "React", Licence: "MIT", Holder: "Facebook, Inc. and its affiliates"},
}

// NewAbout answers the About content for a version.
func NewAbout(version string) About {
	return About{
		Name:      product.Name,
		Version:   version,
		Author:    product.Author,
		Copyright: product.Copyright,
		Statement: product.Statement,
		Credits:   append([]Credit(nil), credits...),
	}
}
