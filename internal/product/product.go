// Package product holds the facts about SymChit that every layer may name: what
// it is called and who made it. They live here once so that a rename is one
// edit, never a search.
package product

// Name is the product name as a reader sees it.
const Name = "SymChit"

// Author is the copyright holder.
const Author = "Oliver Ernster"

// Statement is what SymChit is and is not, shown in About (FR-067).
const Statement = "SymChit records what you observed and when. " +
	"It gives no medical advice and does not interpret your symptoms; " +
	"that is for you and your healthcare professional."
