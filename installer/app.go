package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/symchit/internal/infrastructure/setup"
)

// uninstallExeName is the copy of setup left inside the install directory, so
// the Apps list has something to call after the downloaded setup file is gone.
const uninstallExeName = "uninstall.exe"

// App is the Wails facade for the setup program. Everything the page can do
// goes through a method here; the install policy itself lives in
// internal/infrastructure/setup, so this file owns none of it.
type App struct {
	ctx           context.Context
	payload       []byte
	version       string
	uninstallMode bool
}

// NewApp builds the facade. Started with -uninstall, as the Apps list starts
// it, setup opens on the removal screen rather than on the manage one.
func NewApp(payload []byte, version string) *App {
	uninstall := len(os.Args) > 1 && os.Args[1] == setup.UninstallFlag
	return &App{
		payload:       payload,
		version:       version,
		uninstallMode: uninstall,
	}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// StateDTO describes what setup should offer, given what is already installed.
//
// AppName is sent because the page must not write the product's name down. A
// page has no build step, so a name typed into it survives a rename with
// nothing to say so.
type StateDTO struct {
	AppName          string `json:"appName"`
	Mode             string `json:"mode"`
	Relation         string `json:"relation"`
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installedVersion"`
	ThisVersion      string `json:"thisVersion"`
	InstallDir       string `json:"installDir"`
	RecordFile       string `json:"recordFile"`
	StartMenu        bool   `json:"startMenu"`
	Desktop          bool   `json:"desktop"`
}

// OptionsDTO carries the choices made on the install or reinstall screen.
type OptionsDTO struct {
	StartMenu bool `json:"startMenu"`
	Desktop   bool `json:"desktop"`
}

// Progress is emitted on the "progress" event while a long operation runs.
type Progress struct {
	Pct int    `json:"pct"`
	Msg string `json:"msg"`
}

// relationNames turn the comparison into the word the page routes on.
var relationNames = map[setup.Relation]string{
	setup.Newer: "newer",
	setup.Same:  "same",
	setup.Older: "older",
}

// DetectState reads the machine once and answers the route setup should open
// in. Reading it once is what keeps the screen, its heading, its options and
// its buttons from drifting apart.
func (a *App) DetectState() StateDTO {
	dir, _ := setup.InstallDir()
	record, _ := setup.RecordFile()
	installedVersion, installed := setup.InstalledVersion()
	shortcuts := setup.CurrentShortcuts()

	mode := "install"
	switch {
	case a.uninstallMode:
		mode = "uninstall"
	case installed:
		mode = "manage"
	}
	relation := setup.Same
	if installed {
		relation = setup.Compare(a.version, installedVersion)
	}
	return StateDTO{
		AppName:          setup.AppName,
		Mode:             mode,
		Relation:         relationNames[relation],
		Installed:        installed,
		InstalledVersion: installedVersion,
		ThisVersion:      a.version,
		InstallDir:       dir,
		RecordFile:       record,
		StartMenu:        shortcuts.StartMenu,
		Desktop:          shortcuts.Desktop,
	}
}

// AppRunning reports whether SymChit is open, so the page can offer to close it
// rather than failing later on a locked executable.
func (a *App) AppRunning() bool { return setup.IsAppRunning() }

// CloseRunningApp ends the running application so setup can proceed.
func (a *App) CloseRunningApp() error { return setup.CloseRunningApp() }

// Install performs a fresh install, an update, a way back or a reinstall. All
// four are the same act: write the files, then apply the options as given.
func (a *App) Install(choices OptionsDTO) error { return a.write(choices) }

// Repair writes the files again and leaves every option as it stands. It is the
// quick fix for a damaged install, as distinct from a reinstall, which puts the
// choices back to those of a new install.
func (a *App) Repair() error {
	shortcuts := setup.CurrentShortcuts()
	return a.write(OptionsDTO{StartMenu: shortcuts.StartMenu, Desktop: shortcuts.Desktop})
}

// write is the single install path behind Install and Repair.
func (a *App) write(choices OptionsDTO) error {
	if setup.IsAppRunning() {
		return setup.ErrAppRunning
	}
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}

	// The weighting is measured rather than counted: extracting the payload is
	// most of the work, so the bar sits in it rather than reaching the end in a
	// twentieth of a second and waiting there.
	a.progress(10, "Writing the files...")
	if err := setup.ExtractZip(a.payload, dir); err != nil {
		return fmt.Errorf("write the files: %w", err)
	}
	exePath := filepath.Join(dir, setup.ExeName)

	a.progress(70, "Registering SymChit with Windows...")
	if err := a.register(dir, exePath); err != nil {
		return err
	}

	a.progress(90, "Applying your choices...")
	setup.ApplyShortcuts(exePath, dir, setup.Shortcuts{
		StartMenu: choices.StartMenu,
		Desktop:   choices.Desktop,
	})

	a.progress(100, "Done.")
	return nil
}

// register leaves a copy of setup beside the application and writes the Apps
// list entry that points at it.
func (a *App) register(dir, exePath string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate setup: %w", err)
	}
	uninstallExe := filepath.Join(dir, uninstallExeName)
	if err := setup.CopyFile(self, uninstallExe); err != nil {
		return fmt.Errorf("write the uninstaller: %w", err)
	}
	sizeKB, _ := setup.DirSizeKB(dir)
	if err := setup.WriteUninstallEntry(setup.UninstallInfo{
		Version:      a.version,
		InstallDir:   dir,
		UninstallExe: uninstallExe,
		IconPath:     exePath,
		EstimatedKB:  sizeKB,
	}); err != nil {
		return fmt.Errorf("register SymChit: %w", err)
	}
	return nil
}

// Uninstall removes the shortcuts, the registry entry, the log and the
// installed files, plus the window's own WebView2 folder, which holds no part
// of the record.
//
// The record itself is kept unless removeRecord says otherwise (FR-072).
// Removing the program is not a decision to throw away years of observations,
// so the screen names the file and asks.
func (a *App) Uninstall(removeRecord bool) error {
	if setup.IsAppRunning() {
		return setup.ErrAppRunning
	}
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}

	a.progress(20, "Removing shortcuts...")
	setup.RemoveShortcuts()

	a.progress(45, "Removing the registry entry...")
	_ = setup.RemoveUninstallEntry()

	a.progress(60, "Clearing what the window kept...")
	if webview, webviewErr := setup.WebViewDir(); webviewErr == nil {
		_ = setup.RemoveTree(webview)
	}
	if logs, logErr := setup.LogDir(); logErr == nil {
		_ = setup.RemoveTree(logs)
	}

	if removeRecord {
		a.progress(75, "Deleting your symptom record...")
		if record, recordErr := setup.RecordDir(); recordErr == nil {
			_ = setup.RemoveTree(record)
		}
	}

	a.progress(90, "Removing the files...")
	setup.ScheduleDirDeletion(dir)

	a.progress(100, "Done.")
	return nil
}

// LaunchApp starts the installed application, backing the "start it when this
// finishes" option.
func (a *App) LaunchApp() error { return setup.LaunchApp() }

// SetShortcuts applies the shortcut boxes live from the manage screen, where
// there is nothing to install, so a box that waited for a go-ahead would never
// take effect at all.
func (a *App) SetShortcuts(startMenu, desktop bool) error {
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}
	setup.ApplyShortcuts(filepath.Join(dir, setup.ExeName), dir, setup.Shortcuts{
		StartMenu: startMenu,
		Desktop:   desktop,
	})
	return nil
}

// Quit closes the setup program.
func (a *App) Quit() { wailsruntime.Quit(a.ctx) }

// progress reports how far a long operation has got.
func (a *App) progress(pct int, msg string) {
	wailsruntime.EventsEmit(a.ctx, "progress", Progress{Pct: pct, Msg: msg})
}
