// Command installer is the bespoke SymChit setup program.
//
// It is built as a second Wails application in the same module, so it wears the
// same WebView and the same palette as the application it installs. It carries
// the built application as an embedded zip and covers install, update, going
// back a version, repair, reinstall and uninstall, all per user with no
// administrator rights.
//
// Ported from ED Voyage Companion's installer, which is PigeonPost's model with
// the install policy moved into infrastructure.
package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/oernster/symchit/internal/infrastructure/setup"
	"github.com/oernster/symchit/internal/product"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed payload.zip
var payload []byte

// appVersion is overridden at build time with -ldflags "-X main.appVersion=x.y.z",
// which only reaches a var. The setup program holds no version literal.
var appVersion = "dev"

const (
	windowTitle = product.Name + " Setup"
	// The window is fixed, so its height has to clear the tallest screen: the
	// install one, which carries the path box and three options. Both figures
	// are measured against that screen at the shipped type sizes; text that
	// grows without the window growing with it turns a fixed dialog into a
	// scrolling one.
	windowWidth  = 860
	windowHeight = 720
	// webviewFolder holds the setup window's own WebView2 cache, pinned under
	// TEMP rather than left to default into %APPDATA%, so running setup leaves
	// no folder beside the application's own.
	webviewFolder = product.Name + "Setup"
)

// dark and light are the surface colours from the application's own palette, so
// setup never flashes a ground the application would not draw.
var (
	dark  = options.RGBA{R: 0x14, G: 0x17, B: 0x1c, A: 1}
	light = options.RGBA{R: 0xf4, G: 0xf6, B: 0xf8, A: 1}
)

func main() {
	prefersDark := setup.SystemPrefersDark()
	background := light
	if prefersDark {
		background = dark
	}
	app := NewApp(payload, appVersion, prefersDark)
	_ = wails.Run(&options.App{
		Title:            windowTitle,
		Width:            windowWidth,
		Height:           windowHeight,
		DisableResize:    true,
		BackgroundColour: &background,
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
		Windows: &windowsoptions.Options{
			WebviewUserDataPath: filepath.Join(os.TempDir(), webviewFolder),
		},
	})
}
