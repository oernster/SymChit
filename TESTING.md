# SymChit: testing

## The gate

```powershell
./test.ps1
```

It runs these in order, stopping at the first failure:

1. `gofmt -l` over the Go, ignoring the front end.
2. `go vet` over every package but `node_modules`.
3. A build and a vet for Linux and for macOS. SymChit ships to all three, so a
   Windows-only import is a defect the day it is written rather than the day
   someone tries the Flatpak. This needs no Linux machine and no Mac; it fails
   naming the import and the platform: proved by planting
   `golang.org/x/sys/windows` in a file with no build tag.
4. `staticcheck`, fetched with `go run honnef.co/go/tools/cmd/staticcheck@latest`.
5. The whole Go suite.
6. Coverage of `internal/domain` and `internal/application`, which must be 100%.
7. Coverage of every other Go package against its measured floor.
8. The front end: `eslint`, `tsc --noEmit` and `vite build`.
9. The front-end suite: Vitest over jsdom.

`./test.ps1 -SkipFrontend` runs the Go half alone while working on it. There is
no switch that skips the gate inside `build.ps1`.

## How to read a result

Read the exit code, never the last line of output. The coverage run prints a
table last, so a `tail` shows coverage rows rather than a verdict; a grep for `error` matches file names.

```powershell
./test.ps1; "EXIT=$LASTEXITCODE"
```

`0` means every check passed and every floor held. Anything else means read the
failure the script threw.

## What a first run needs from the machine

- Go 1.26 or later; Node 20 or later.
- `npm install` inside `frontend/` once. `test.ps1` does not install for you.
- Step 3 downloads staticcheck the first time; it needs the network once.
- On Windows, allow Go's scratch directory (`go env GOTMPDIR`) in your
  anti-virus. A quarantined test binary stops the suite running at all rather
  than failing a test.

## The floors

Each is the measured number at the time of writing, not a target. A floor at an
aspiration only teaches people to lower it; a floor at the measured number
fails the moment cover is lost, which is the only moment it is worth being
told.

| Package | Floor | Why not 100 |
|---|---|---|
| `internal/domain` | 100 | Nothing to excuse: pure rules. |
| `internal/application` | 100 | Every port is faked, so every path is reachable. |
| `internal/infrastructure/store` | 92 | The rest is SQLite write failures that cannot be forced without breaking the disk. |
| `internal/infrastructure/export` | 91 | The rest is operating-system write failures on a temporary file. |
| `internal/infrastructure/runlog` | 81 | The rest is Win32 standard-handle work, reachable only in a windowed process with no error output. It was 74 until the folder rule became a pure function taking the platform as an argument: all three platforms' answers are now exercised on whichever platform the suite runs on, rather than two of them waiting for a user to report the answer. |
| `internal/infrastructure/setup` | 56 | The portable half and the shortcut writing are tested. The registry writes, the process work and the scheduled deletion change the machine, so a real install exercises them instead. |
| root package (the facade) | 76 | The facade itself is covered. `main`, the log handover, the file dialogs, the browser opener and the single-instance lock need a real window. The floor was 77 until the donate button landed: the opener is one more line of Wails runtime no test can reach, so the blend fell by half a point and the floor was re-measured rather than the facade going untested. |

Not gated at all: `internal/product` holds two constants; `tests/structural` is itself the guard.

## What each suite proves

| Suite | What it settles |
|---|---|
| `internal/domain` | The rules: blank and future occurrences refused, case variants matched, filters combined, the receipt's grouping, ordering and exact lines, including a range that crosses a year. |
| `internal/application` | The use cases over fakes: recording reads the clock once, an edit leaves recorded-at alone, a failed write loses nothing the user typed, an import skips what is held. |
| `internal/infrastructure/store` | Real SQLite in a temporary folder: notes and symptoms read back byte for byte, a deletion is all or nothing, a garbage file is never overwritten, a newer schema is refused, the durability pragmas are set, plus a planted trigger proving a failed write leaves nothing behind. |
| `internal/infrastructure/export` | The format against a committed sample, a refusal for anything that is not a SymChit export, plus a size cap read before the file is. |
| `internal/infrastructure/runlog` | Crash reporting, by starting this test binary again as a child and making it panic, plus each platform's rule for where the log lives, all three exercised wherever the suite runs. |
| root package | The facade end to end over a real record: every conversion, every refusal, plus a panic in a bound method becoming an error rather than a dead window. |
| `internal/infrastructure/setup` | The install policy: the payload fence refusing an entry that climbs out of the install folder, the paths, the version comparison that picks the route, plus real shortcuts written into temporary folders with plain paths rather than doubled separators. |
| `tests/structural` | The invariants in ARCHITECTURE.md. |
| `frontend` | The page over a fake facade: the keyboard path to a recorded event, the form keeping everything when a save is refused, an edit sending no time unless it changed, the confirmation naming what will go, plus the receipt showing exactly the lines it was given. The self-reading cycle is covered twice: the pure machine tick by tick, then the hook under jsdom for what suspends it and what freezes it. The focus ring is covered the same way: the rules alone, then the ring against a real page, where Tab and Right agree, Shift+Tab and Left agree, both wrap, a disabled control is passed over and a text field keeps its arrows. |

## What the tests never do

- They never reach the network. Nothing in the suite may.
- They never touch your real record: every store test opens a file under `t.TempDir()`; the facade tests do the same.
- They never mock SQLite. The store is tested against the real engine.
- They never assert against Wails' generated bindings. The wire contract is
  `frontend/src/api.ts`, compared with the Go DTOs by a structural test.

## What only a person can settle

These were measured by hand on the reference machine on 2026-09-22, against a
build run with a sandboxed `APPDATA`. They need doing again when the window
changes.

| Check | Last result |
|---|---|
| The window opens and follows the Windows light or dark mode. | Dark, matching Windows. |
| Wails' IPC works under the page's Content-Security-Policy. | Works: the severities shown came from Go. |
| `window.print()` opens the Windows print dialog. | It does, offering Save as PDF and Microsoft Print to PDF. |
| The printed page carries the receipt alone. | It does: no band, no status line, no controls. |
| Recording, editing, deleting, exporting and importing through the real window. | All worked; the import skipped the event already held. |
| The Guide and About open, scroll and close. | They do. |
| The Guide and About hold still for five seconds, then read themselves down; they step aside the moment the reader scrolls. | Not yet run in the real window. The cycle is covered by the suite and the pane's ring was measured in Chromium against the built stylesheet; neither is a reading of the running application. |
| Tab reaches the Guide's text without drawing a ring round it; Close rings when it is reached. | Measured in Chromium against the built stylesheet: the pane draws nothing at rest, hovered, clicked or Tab-focused, while Close draws the 2px ring. Not yet confirmed in the real window. |
| The three-state ring: nothing at rest, green on hover or focus, permanent red while disabled. | Measured in Chromium against the built stylesheet, both modes, with `:hover` and `:focus-visible` confirmed each time. Dark: at rest `3px none` on the band, Show and an enabled Print; hovered or Tab-focused `2px solid rgb(52,211,153)`; disabled Print and Delete `2px solid rgb(255,138,128)` on a `rgb(42,48,58)` fill. Light: hovered `2px solid rgb(4,120,87)`, disabled `2px solid rgb(168,35,26)` on `rgb(228,232,237)`. The disabled Print is skipped by Tab. Not yet confirmed in the real window. |
| The window opens with nothing ringed; the first Tab enters the band. | Not yet run in the real window. The sink is measured as holding focus while being no stop; the ring's own stepping is covered by the suite. |
| Setup installs, with the options opening on what the machine already holds. | It does: files, registry entry, shortcuts, plus the application launched from the install folder. |
| Setup reopens on the manage screen when the versions match; a shortcut box applies immediately. | It does. |
| Setup started with `-uninstall`, as the Apps list starts it, opens on the removal screen. | It does. |
| Uninstall removes the program, its shortcuts, its log and the window's WebView2 folder, while keeping the record. | It does; the record was still there afterwards. |
| The Apps list's Uninstall and Modify point at a path that exists. | They do now. The first install wrote doubled separators; fixed and covered by a test. |
| The install folder is removed after setup exits. | Measured with a probe: the ported command removed nothing, so it was rewritten. Still to check by hand on a real install. |
| Uninstall with "also delete my symptom record" ticked. | Not yet run. |
| The Flatpak builds on a Linux machine, the window opens and the record lands under `~/.var/app/uk.codecrafter.SymChit`. | Not yet run. The manifest was checked by generating it and parsing it: valid YAML, no network permission on the finished application, plus every one of the eight icons it installs present in the tree. That says nothing about whether it builds. |
| The DMG builds, signs, notarizes and staples on an Apple Silicon Mac. | Not yet run. Nothing about this script has been measured beyond its syntax. |

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): what the structural tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [REQUIREMENTS.md](REQUIREMENTS.md): each requirement names the test that
  verifies it.
