package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadsIsWhereTheDialogsOpen(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	if got := downloads(); got != "" {
		t.Errorf("with no Downloads folder = %q, want the empty string", got)
	}
	folder := filepath.Join(home, downloadsName)
	if err := os.Mkdir(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := downloads(); got != folder {
		t.Errorf("downloads = %q, want %q", got, folder)
	}
	file := filepath.Join(home, "elsewhere")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USERPROFILE", file)
	if got := downloads(); got != "" {
		t.Errorf("a Downloads that is not a folder = %q, want the empty string", got)
	}
}

func TestTheSystemClockReadsTheWallClock(t *testing.T) {
	t.Parallel()
	before := time.Now()
	read := systemClock{}.Now()
	if read.Before(before) || time.Since(read) > time.Minute {
		t.Errorf("the clock read %v, which is not now", read)
	}
}
