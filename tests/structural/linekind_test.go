package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"
)

// kindSource is where the receipt's line kinds are defined.
var kindSource = filepath.Join("internal", "domain", "receipt.go")

// kindType is the Go type whose constants are the kinds.
const kindType = "LineKind"

// tsKindUnion pulls the kind field's union out of the front end's ReceiptLine.
// It reaches from the field name to the end of the interface, so a union laid
// out on one line or over several reads the same.
var tsKindUnion = regexp.MustCompile(`(?s)interface ReceiptLine \{.*?kind:(.*?)\n\s*text:`)

// tsKindMember picks one quoted member out of that union.
var tsKindMember = regexp.MustCompile(`'([a-z]+)'`)

// TestEveryLineKindIsDeclaredToThePage keeps the two statements of the receipt's
// vocabulary in step.
//
// The wire test compares field names, so it sees that a line carries a kind; it
// cannot see WHICH kinds exist. A kind added in Go alone arrives at a page whose
// union does not admit it, which the type checker cannot catch because nothing
// in the build compares the two. The failure is silent and cosmetic at first: a
// line renders with a class no stylesheet knows, so it prints looking like
// whatever sits above it. FR-045's framing is exactly that case.
func TestEveryLineKindIsDeclaredToThePage(t *testing.T) {
	root := repoRoot(t)
	inGo := goLineKinds(t, root)
	onPage := pageLineKinds(t, root)

	for _, kind := range absent(inGo, onPage) {
		t.Errorf("the domain writes a %q line but %s does not admit that kind: the page "+
			"would style it as nothing", kind, filepath.ToSlash(wireContract))
	}
	for _, kind := range absent(onPage, inGo) {
		t.Errorf("%s admits a %q line that the domain never writes: a kind nothing sends "+
			"is a rule a reader will look for and not find",
			filepath.ToSlash(wireContract), kind)
	}
}

// goLineKinds answers the kinds the domain declares, by their wire values.
func goLineKinds(t *testing.T, root string) []string {
	t.Helper()
	path := filepath.Join(root, kindSource)
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", filepath.ToSlash(kindSource), err)
	}
	var out []string
	for _, declaration := range parsed.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, spec := range general.Specs {
			valued, ok := spec.(*ast.ValueSpec)
			if !ok || !isKindType(valued.Type) {
				continue
			}
			for _, value := range valued.Values {
				literal, ok := value.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				text, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatalf("unreadable kind %s in %s", literal.Value, kindSource)
				}
				out = append(out, text)
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no %s constants found in %s, the scan is wrong",
			kindType, filepath.ToSlash(kindSource))
	}
	sort.Strings(out)
	return out
}

// isKindType answers whether a declared type is LineKind.
func isKindType(node ast.Expr) bool {
	named, ok := node.(*ast.Ident)
	return ok && named.Name == kindType
}

// pageLineKinds answers the kinds the front end's ReceiptLine admits.
func pageLineKinds(t *testing.T, root string) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, wireContract))
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(wireContract), err)
	}
	union := tsKindUnion.FindSubmatch(tsComment.ReplaceAll(raw, nil))
	if union == nil {
		t.Fatalf("no kind union found in %s, the pattern is wrong",
			filepath.ToSlash(wireContract))
	}
	var out []string
	for _, member := range tsKindMember.FindAllSubmatch(union[1], -1) {
		out = append(out, string(member[1]))
	}
	sort.Strings(out)
	return out
}
