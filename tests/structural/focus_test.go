package structural

// A pane is chrome: it holds the controls, it is not one of them. A ring drawn
// round it marks nothing the reader can act on and costs them a keypress to
// step past, so no container may take focus or paint a border. The guards here
// hold that rule over both front ends, the application's and the setup
// program's, the latter because its page has no build step of its own and so is
// checked by nothing else.
//
// They also hold the self-reading cycle's own requirement: a dialog body that
// scrolls pins its action row beneath it and wears the cycle, so Close never
// drifts away as the words read themselves.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// setupPage is the setup program's hand-written front end. It is generated
// output as far as the other walks are concerned (it lives under dist), so it
// is named here rather than found.
var setupPage = []string{
	filepath.Join("installer", "frontend", "dist", "setup.css"),
	filepath.Join("installer", "frontend", "dist", "index.html"),
	filepath.Join("installer", "frontend", "dist", "setup-routes.js"),
	filepath.Join("installer", "frontend", "dist", "setup-shell.js"),
}

// ringStates are the pseudo-classes that make a rule a ring rule. focus-within
// counts: a ring drawn on a wrapper because something inside it took focus is
// still a ring on the wrapper.
var ringStates = []string{":focus-visible", ":focus-within", ":focus", ":hover"}

// containerTags are the elements whose whole job is to hold other things. None
// of them may carry a ring rule or a tabindex.
var containerTags = map[string]bool{
	"div": true, "section": true, "main": true, "article": true, "aside": true,
	"header": true, "footer": true, "nav": true, "form": true, "fieldset": true,
	"ul": true, "ol": true, "li": true, "table": true, "tbody": true, "tr": true,
	"p": true, "span": true, "label": true,
}

// controlTags are the elements a reader acts on; the only ones that may carry a
// tabindex.
var controlTags = map[string]bool{
	"button": true, "input": true, "textarea": true, "select": true, "a": true,
}

// invisibleValues are the ways a declaration says "draw nothing", which is the
// opposite of a ring and always allowed. `* { outline: none }` suppressing the
// native indicator is the sanctioned use of the universal selector.
var invisibleValues = map[string]bool{
	"none": true, "0": true, "transparent": true, "unset": true, "initial": true, "hidden": true,
}

// borderDeclaration matches a declaration that could draw a ring.
var borderDeclaration = regexp.MustCompile(`(?m)(outline|border|border-color|border-top|border-bottom|border-left|border-right)\s*:\s*([^;}]+)`)

// pseudoSuffix strips the pseudo-classes and pseudo-elements off a selector
// part, leaving what the rule is actually aimed at.
var pseudoSuffix = regexp.MustCompile(`::?[a-z-]+(\([^)]*\))?`)

// tabIndexAttribute matches the attribute in JSX and in plain HTML alike.
var tabIndexAttribute = regexp.MustCompile(`(?i)\btabindex\s*=`)

// These are vacuity floors: a rename that stopped the scans matching would
// otherwise turn every assertion below into a silent pass over an empty list.
const (
	leastRingRules     = 3
	leastScrollingBody = 2
)

// ringSurfaces returns every file the focus rules govern: the application's
// front end plus the setup program's page.
func ringSurfaces(t *testing.T) []string {
	t.Helper()
	found := frontendFiles(t)
	root := repoRoot(t)
	for _, relative := range setupPage {
		path := filepath.Join(root, relative)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("the setup page names %s, which is not there: %v", filepath.ToSlash(relative), err)
		}
		found = append(found, path)
	}
	return found
}

// readFile answers a file's contents, failing the test rather than the caller.
func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(raw)
}

// cssRules splits a stylesheet into selector and block pairs. It is a naive
// split on the braces, which is enough because no rule here nests.
func cssRules(source string) [][2]string {
	var rules [][2]string
	for _, chunk := range strings.Split(source, "}") {
		open := strings.Index(chunk, "{")
		if open < 0 {
			continue
		}
		selector := strings.TrimSpace(stripComments(chunk[:open]))
		if selector == "" || strings.HasPrefix(selector, "@") {
			continue
		}
		rules = append(rules, [2]string{selector, chunk[open+1:]})
	}
	return rules
}

// stripComments removes block comments, so a selector quoted in prose is never
// read as a rule.
func stripComments(source string) string {
	for {
		open := strings.Index(source, "/*")
		if open < 0 {
			return source
		}
		close := strings.Index(source[open:], "*/")
		if close < 0 {
			return source[:open]
		}
		source = source[:open] + source[open+close+2:]
	}
}

// drawsSomething reports whether a block sets a border or an outline that the
// reader would actually see.
func drawsSomething(block string) bool {
	for _, match := range borderDeclaration.FindAllStringSubmatch(block, -1) {
		value := strings.ToLower(strings.TrimSpace(match[2]))
		if invisibleValues[value] {
			continue
		}
		// A shorthand such as `border: 1px solid transparent` draws nothing
		// either; the colour is the last word that matters.
		if strings.Contains(value, "transparent") {
			continue
		}
		return true
	}
	return false
}

// isRingSelector reports whether a selector part is stated for a focus or hover
// state.
func isRingSelector(part string) bool {
	for _, state := range ringStates {
		if strings.Contains(part, state) {
			return true
		}
	}
	return false
}

// targetOf answers what a selector part is aimed at once its pseudo-classes are
// taken off: the last simple selector, since that is the element the rule
// paints.
func targetOf(part string) string {
	bare := strings.TrimSpace(pseudoSuffix.ReplaceAllString(part, ""))
	fields := strings.FieldsFunc(bare, func(r rune) bool {
		return r == ' ' || r == '>' || r == '+' || r == '~'
	})
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// TestNoRingRuleNamesAContainer holds the ring to the controls.
//
// Proved by planting: restoring the bare `:focus-visible { outline: 2px solid
// var(--focus) }` that stood in styles.css fails this by name, as does adding
// `div:hover { border: 1px solid var(--border) }`.
func TestNoRingRuleNamesAContainer(t *testing.T) {
	scanned := 0
	for _, path := range ringSurfaces(t) {
		if !strings.EqualFold(filepath.Ext(path), ".css") {
			continue
		}
		for _, rule := range cssRules(readFile(t, path)) {
			if !drawsSomething(rule[1]) {
				continue
			}
			for _, part := range strings.Split(rule[0], ",") {
				part = strings.TrimSpace(part)
				if !isRingSelector(part) {
					continue
				}
				scanned++
				target := targetOf(part)
				if target == "" {
					t.Errorf("%s: %q rings whatever takes focus; name the controls",
						filepath.Base(path), part)
					continue
				}
				if target == "*" || containerTags[strings.ToLower(target)] {
					t.Errorf("%s: %q rings a container; the ring belongs to a control",
						filepath.Base(path), part)
				}
			}
		}
	}
	if scanned < leastRingRules {
		t.Fatalf("only %d ring rules were scanned, fewer than the %d that exist: the scan is wrong",
			scanned, leastRingRules)
	}
}

// scrollDeclaration matches a container being made to scroll.
var scrollDeclaration = regexp.MustCompile(`(?m)overflow(-y)?\s*:\s*(auto|scroll)\b`)

// suppressesNativeRing reports whether a block turns the engine's own focus
// indicator off.
func suppressesNativeRing(block string) bool {
	for _, match := range borderDeclaration.FindAllStringSubmatch(block, -1) {
		if match[1] != "outline" {
			continue
		}
		if invisibleValues[strings.ToLower(strings.TrimSpace(match[2]))] {
			return true
		}
	}
	return false
}

// TestEveryScrollingSurfaceSuppressesTheNativeRing closes the half of the rule
// that removing a stylesheet rule cannot reach.
//
// Measured in Chromium against the built stylesheet: an overflowing scroll
// container is keyboard focusable with no tabindex at all; a Tab press drew the
// engine's own `1px auto` ring round the whole of the Guide. The surface
// stays a stop, since it carries no controls of its own and a reader has to be
// able to scroll it; it must paint nothing while it is one.
//
// Proved by planting: deleting the `.dialog-body:focus` suppression from
// dialogs.css fails this by name.
func TestEveryScrollingSurfaceSuppressesTheNativeRing(t *testing.T) {
	for _, path := range ringSurfaces(t) {
		if !strings.EqualFold(filepath.Ext(path), ".css") {
			continue
		}
		rules := cssRules(readFile(t, path))
		// A sheet may suppress the indicator for everything at once, which
		// covers each of its scrollers without naming them.
		everywhere := false
		for _, rule := range rules {
			if targetOf(rule[0]) == "" && isRingSelector(rule[0]) && suppressesNativeRing(rule[1]) {
				everywhere = true
			}
		}
		if everywhere {
			continue
		}
		for _, rule := range rules {
			if !scrollDeclaration.MatchString(rule[1]) {
				continue
			}
			if !suppressedFor(rules, rule[0]) {
				t.Errorf("%s: %q scrolls without turning the engine's own focus ring off: Tab would ring the whole pane",
					filepath.Base(path), rule[0])
			}
		}
	}
}

// suppressedFor reports whether a sheet turns the native ring off for a
// selector that was made to scroll.
func suppressedFor(rules [][2]string, scroller string) bool {
	for _, rule := range rules {
		if !isRingSelector(rule[0]) || !suppressesNativeRing(rule[1]) {
			continue
		}
		for _, part := range strings.Split(rule[0], ",") {
			if targetOf(part) == targetOf(scroller) {
				return true
			}
		}
	}
	return false
}

// TestNoContainerTakesFocus keeps containers out of the tab order, which is the
// other half of the rule: a pane that cannot be reached cannot be ringed.
//
// Proved by planting: `<div className="dialog-body" tabIndex={0}>` in
// GuideDialog.tsx fails this by name.
func TestNoContainerTakesFocus(t *testing.T) {
	for _, path := range ringSurfaces(t) {
		source := readFile(t, path)
		for _, where := range tabIndexAttribute.FindAllStringIndex(source, -1) {
			tag := tagHolding(source, where[0])
			if !controlTags[tag] {
				t.Errorf("%s: <%s> carries a tabindex; only a control may be a stop",
					filepath.Base(path), tag)
			}
		}
	}
}

// tagHolding answers the name of the element whose opening tag an attribute
// sits in, by walking back to the nearest `<`.
func tagHolding(source string, at int) string {
	open := strings.LastIndex(source[:at], "<")
	if open < 0 {
		return "?"
	}
	rest := source[open+1:]
	end := strings.IndexFunc(rest, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '>' || r == '/'
	})
	if end < 0 {
		return "?"
	}
	return strings.ToLower(rest[:end])
}

// TestEveryScrollingDialogPinsItsActionsAndReadsItself holds the two things a
// scrolling dialog body needs: the action row outside it, so Close cannot
// scroll off a short window; the self-reading cycle on it, so a new dialog
// cannot quietly ship without the cycle every other one has.
//
// Proved by planting: dropping `pinnedActions` from GuideDialog.tsx fails the
// first check by name; dropping `ref={autoScroll}` fails the second.
func TestEveryScrollingDialogPinsItsActionsAndReadsItself(t *testing.T) {
	const (
		bodyClass = `className="dialog-body"`
		pinned    = "pinnedActions"
		cycle     = "ref={autoScroll}"
	)
	scrolling := 0
	for _, path := range frontendFiles(t) {
		if !strings.EqualFold(filepath.Ext(path), ".tsx") || strings.Contains(path, ".test.") {
			continue
		}
		source := readFile(t, path)
		// Only a file that draws a dialog is judged. The shell that implements
		// the layout declares the prop without wearing it.
		if !strings.Contains(source, "<Modal") {
			continue
		}
		holdsBody := strings.Contains(source, bodyClass)
		if holdsBody {
			scrolling++
		}
		if holdsBody && !strings.Contains(source, pinned) {
			t.Errorf("%s scrolls a body without pinning its actions: Close would scroll away with the words",
				filepath.Base(path))
		}
		if holdsBody && !strings.Contains(source, cycle) {
			t.Errorf("%s scrolls a body that does not read itself: it wants %s",
				filepath.Base(path), cycle)
		}
		if !holdsBody && strings.Contains(source, pinned) {
			t.Errorf("%s pins its actions with nothing to scroll: the dialog would clip its own content",
				filepath.Base(path))
		}
	}
	if scrolling < leastScrollingBody {
		t.Fatalf("found %d scrolling dialog bodies, fewer than the %d that exist: the scan is wrong",
			scrolling, leastScrollingBody)
	}
}
