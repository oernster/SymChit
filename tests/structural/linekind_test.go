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
	var out []string
	for _, value := range goLineKindsByName(t, root) {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

// goLineKindsByName answers the same constants by their Go names, which is what
// anything reading them in Go refers to them as.
func goLineKindsByName(t *testing.T, root string) map[string]string {
	t.Helper()
	path := filepath.Join(root, kindSource)
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", filepath.ToSlash(kindSource), err)
	}
	out := map[string]string{}
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
			for at, value := range valued.Values {
				literal, ok := value.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				text, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatalf("unreadable kind %s in %s", literal.Value, kindSource)
				}
				out[valued.Names[at].Name] = text
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no %s constants found in %s, the scan is wrong",
			kindType, filepath.ToSlash(kindSource))
	}
	return out
}

// styleSource is where the document says how each kind is drawn.
var styleSource = filepath.Join("internal", "infrastructure", "pdf", "layout.go")

// styleTable reaches the styles table alone.
//
// It has to: the same file names kinds elsewhere, in the switch deciding what
// starts a block, where those names carry a colon after them too. A pattern read
// over the whole file therefore passed while the styles table was missing an
// entry, which a planted violation showed and nothing else would have.
var styleTable = regexp.MustCompile(`(?s)var styles = map\[domain\.LineKind\]style\{(.*?)\n\}`)

// styledKind picks one kind out of that table.
var styledKind = regexp.MustCompile(`domain\.(Line[A-Za-z]+):`)

// TestEveryLineKindIsDrawnInTheDocument keeps the record the reader takes away
// in step with the record the domain writes.
//
// The page's half of this is checked above. The document is the other half and
// it fails more quietly: a kind with no entry in the styles table is drawn at
// no size at all, so the line does not appear on the paper and nothing says so.
// A symptom record that silently omits a line is the worst failure this program
// has, which is why a missing style is a failing build rather than a blank.
func TestEveryLineKindIsDrawnInTheDocument(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, styleSource))
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(styleSource), err)
	}
	table := styleTable.FindStringSubmatch(string(raw))
	if table == nil {
		t.Fatalf("no styles table found in %s, the pattern is wrong",
			filepath.ToSlash(styleSource))
	}
	styled := map[string]bool{}
	for _, match := range styledKind.FindAllStringSubmatch(table[1], -1) {
		styled[match[1]] = true
	}
	if len(styled) == 0 {
		t.Fatalf("the styles table in %s is empty", filepath.ToSlash(styleSource))
	}

	declared := goLineKindsByName(t, root)
	var names []string
	for name := range declared {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if !styled[name] {
			t.Errorf("the domain writes a %s line and %s gives it no style: it would be "+
				"drawn at no size and the reader's own words would be missing from the "+
				"document with nothing to say so", name, filepath.ToSlash(styleSource))
		}
		delete(styled, name)
	}
	for name := range styled {
		t.Errorf("%s styles a %s line that the domain never writes",
			filepath.ToSlash(styleSource), name)
	}
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
