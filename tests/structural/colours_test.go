package structural

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// themeFile holds every colour SymChit uses; nothing else may name one.
var themeFile = filepath.Join("frontend", "src", "theme.css")

// minimumContrast is WCAG 2.2 AA for body text (NFR-USE-003). SymChit's text
// is body size throughout, so the larger-text allowance never applies.
const minimumContrast = 4.5

// pairs are the text-on-background pairings the page actually draws. Each is
// held to the minimum in both modes.
var pairs = [][2]string{
	{"text", "bg"},
	{"text", "surface"},
	{"muted", "surface"},
	{"muted", "bg"},
	{"text", "selected"},
	{"on-accent", "accent"},
	{"on-danger", "danger"},
	{"danger", "surface"},
	{"accent", "surface"},
}

// hexColour matches a token declaration such as `--text: #1b1f24;`.
var hexColour = regexp.MustCompile(`--([a-z-]+):\s*(#[0-9a-fA-F]{6})`)

// modes splits the theme into the light block and the dark one. The dark
// block opens at the media query, so everything before it is light.
const darkMarker = "@media (prefers-color-scheme: dark)"

// channel converts one 8-bit component to its linear value.
func channel(value float64) float64 {
	value /= 255
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

// luminance answers a colour's relative luminance, as WCAG defines it.
func luminance(hex string) (float64, error) {
	value, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("unreadable colour %q: %w", hex, err)
	}
	red := channel(float64((value >> 16) & 0xff))
	green := channel(float64((value >> 8) & 0xff))
	blue := channel(float64(value & 0xff))
	return 0.2126*red + 0.7152*green + 0.0722*blue, nil
}

// contrast answers the WCAG contrast ratio between two colours.
func contrast(t *testing.T, first, second string) float64 {
	t.Helper()
	one, err := luminance(first)
	if err != nil {
		t.Fatal(err)
	}
	other, err := luminance(second)
	if err != nil {
		t.Fatal(err)
	}
	lighter, darker := math.Max(one, other), math.Min(one, other)
	return (lighter + 0.05) / (darker + 0.05)
}

// tokensIn reads the colour tokens out of a slice of the theme.
func tokensIn(source string) map[string]string {
	found := map[string]string{}
	for _, match := range hexColour.FindAllStringSubmatch(source, -1) {
		found[match[1]] = match[2]
	}
	return found
}

// TestEveryTextPairingMeetsAA holds the palette to WCAG 2.2 AA in both modes.
//
// Proved by planting a failing shade: setting --muted to #8d97a3 in the light
// block takes muted-on-surface to 3.0 and fails this by name.
func TestEveryTextPairingMeetsAA(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), themeFile))
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(themeFile), err)
	}
	theme := string(raw)
	split := strings.Index(theme, darkMarker)
	if split < 0 {
		t.Fatalf("%s carries no dark mode", filepath.ToSlash(themeFile))
	}
	modes := map[string]map[string]string{
		"light": tokensIn(theme[:split]),
		"dark":  tokensIn(theme[split:]),
	}
	for mode, tokens := range modes {
		for _, pair := range pairs {
			front, back := tokens[pair[0]], tokens[pair[1]]
			if front == "" || back == "" {
				t.Errorf("%s mode names no --%s or no --%s", mode, pair[0], pair[1])
				continue
			}
			if ratio := contrast(t, front, back); ratio < minimumContrast {
				t.Errorf("%s mode: --%s on --%s is %.2f:1, below %.1f:1",
					mode, pair[0], pair[1], ratio, minimumContrast)
			}
		}
	}
}

// TestNoColourIsNamedOutsideTheTheme keeps one home for every colour value.
// The print rules are the exception: paper is white and ink is black whatever
// the screen is doing, so those two are stated where they apply.
func TestNoColourIsNamedOutsideTheTheme(t *testing.T) {
	allowed := map[string]bool{"#ffffff": true, "#000000": true}
	for _, path := range frontendFiles(t) {
		if strings.HasSuffix(path, "theme.css") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		for _, found := range regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`).FindAllString(string(raw), -1) {
			if !allowed[strings.ToLower(found)] {
				t.Errorf("%s names the colour %s: colours live in %s",
					filepath.Base(path), found, filepath.ToSlash(themeFile))
			}
		}
	}
}
