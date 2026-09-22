package main

import (
	"os"
	"path/filepath"

	"github.com/oernster/symchit/internal/product"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// exportFilter limits the file dialogs to export files.
var exportFilter = []runtime.FileFilter{
	{DisplayName: product.Name + " record (*.json)", Pattern: "*.json"},
}

// downloadsName is the folder both file dialogs open in, where a person
// already looks for files they have saved. The user is free to go elsewhere;
// Windows then remembers where they went.
const downloadsName = "Downloads"

// downloads answers the user's Downloads folder, else the empty string where
// there is none, which leaves the dialog to choose for itself.
func downloads() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	folder := filepath.Join(home, downloadsName)
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return ""
	}
	return folder
}

// windowChooser asks through the window's own file dialogs.
type windowChooser struct{ app *App }

// SavePath opens the save dialog with a suggested name.
func (c windowChooser) SavePath(suggested string) (string, error) {
	return runtime.SaveFileDialog(c.app.ctx, runtime.SaveDialogOptions{
		Title: "Export your record", DefaultFilename: suggested,
		DefaultDirectory: downloads(), Filters: exportFilter,
	})
}

// OpenPath opens the open dialog.
func (c windowChooser) OpenPath() (string, error) {
	return runtime.OpenFileDialog(c.app.ctx, runtime.OpenDialogOptions{
		Title: "Import a record", DefaultDirectory: downloads(), Filters: exportFilter,
	})
}

// windowOpener asks the desktop to open an address, through the window's own
// runtime. Wails answers nothing, so a refusal by the desktop is invisible
// here; there is no outcome to report and none is invented.
type windowOpener struct{ app *App }

// Open hands the address to whatever opens links on this machine.
func (o windowOpener) Open(address string) {
	runtime.BrowserOpenURL(o.app.ctx, address)
}

// windowFocus hands the keyboard to the page through the window's own runtime.
type windowFocus struct{ app *App }

// Focus asks the window to show itself, which is what reaches the webview.
//
// Wails' Windows frontend answers Show with SetForegroundWindow plus SetFocus
// on the main window; the WM_SETFOCUS that follows is where it calls Focus on
// the webview, which is the step never taken when the window first opens.
// Read from the Wails 2.12.0 source rather than inferred.
//
// The context is checked because the runtime ENDS THE PROCESS on a nil one
// (`log.Fatalf`); a window that dies rather than opening is the worst outcome
// available here.
func (f windowFocus) Focus() {
	if f.app.ctx == nil {
		return
	}
	runtime.Show(f.app.ctx)
}

// instanceID names SymChit's single-instance lock, per user (FR-064).
const instanceID = "uk.codecrafter.symchit"

// singleInstance brings the running window forward when SymChit is started
// again; the second copy then ends.
func singleInstance(app *App) *options.SingleInstanceLock {
	return &options.SingleInstanceLock{
		UniqueId: instanceID,
		OnSecondInstanceLaunch: func(options.SecondInstanceData) {
			runtime.WindowUnminimise(app.ctx)
			runtime.Show(app.ctx)
		},
	}
}
