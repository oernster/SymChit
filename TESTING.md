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

`./test.ps1 -SkipFrontend` runs the Go half alone while working on it.

`build.ps1` runs the whole gate first and stops on a failure. Its one escape is
`-Fast`, which skips the gate for a working loop and prints that it has done so;
a release is never cut with it.

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
| `internal/infrastructure/setup` | 59 | The portable half and the shortcut writing are tested. The registry writes, the process work and the scheduled deletion change the machine, so a real install exercises them instead. It was 56 until the Windows theme read went: setup opens dark and carries its own toggle, so a registry lookup no test could exercise is gone rather than sitting there lowering the number. |
| root package (the facade) | 76 | The facade itself is covered. `main`, the log handover, the file dialogs, the browser opener and the single-instance lock need a real window. The floor was 77 until the donate button landed: the opener is one more line of Wails runtime no test can reach, so the blend fell by half a point and the floor was re-measured rather than the facade going untested. |

Not gated at all: `internal/product` holds constants and no behaviour, though the
two that reach paper are asserted word for word from the facade's own tests;
`tests/structural` is itself the guard.

## What each suite proves

| Suite | What it settles |
|---|---|
| `internal/domain` | The rules: blank and future occurrences refused, case variants matched, filters combined, the receipt's grouping, ordering and exact lines, including a range that crosses a year. The printed framing too: it wraps the record, the statement prints once and under the range, then two different records are shown to carry framing identical word for word. |
| `internal/application` | The use cases over fakes: recording reads the clock once, an edit leaves recorded-at alone, a failed write loses nothing the user typed, an import skips what is held. |
| `internal/infrastructure/store` | Real SQLite in a temporary folder: notes and symptoms read back byte for byte, a deletion is all or nothing, a garbage file is never overwritten, a newer schema is refused, the durability pragmas are set, plus a planted trigger proving a failed write leaves nothing behind. |
| `internal/infrastructure/export` | The format against a committed sample, a refusal for anything that is not a SymChit export, plus a size cap read before the file is. The version contract as well: every version ever written still reads its own frozen sample, no reader claims a version nothing wrote, a file naming no version is refused rather than read as the current one, plus a gap in the table refusing rather than guessing. |
| `internal/infrastructure/runlog` | Crash reporting, by starting this test binary again as a child and making it panic, plus each platform's rule for where the log lives, all three exercised wherever the suite runs. |
| root package | The facade end to end over a real record: every conversion, every refusal, plus a panic in a bound method becoming an error rather than a dead window. `window_test.go` holds the window's own seams: the Downloads folder both dialogs open in, plus the keyboard being asked for once on DOM ready, including the guard that keeps a focus asked for too early from ending the process. `receipt_test.go` holds the sheet on its own, with the framing's exact words written out a second time, so an edit to the line saying the record is not a diagnosis fails rather than ships. |
| `internal/infrastructure/setup` | The install policy: the payload fence refusing an entry that climbs out of the install folder, the paths, the version comparison that picks the route, plus real shortcuts written into temporary folders with plain paths rather than doubled separators. |
| `tests/structural` | The invariants in ARCHITECTURE.md. |
| `frontend` | The page over a fake facade: the keyboard path to a recorded event, the form keeping everything when a save is refused, an edit sending no time unless it changed, the confirmation naming what will go, plus the receipt showing exactly the lines it was given. The self-reading cycle is covered twice: the pure machine tick by tick, then the hook under jsdom for what suspends it and what freezes it. The focus ring is covered the same way: the rules alone, then the ring against a real page, where Tab and Right agree, Shift+Tab and Left agree, both wrap, a disabled control is passed over and a text field keeps its arrows. |

## What the tests never do

- They never reach the network. Nothing in the suite may.
- They never touch your real record: every store test opens a file under `t.TempDir()`; the facade tests do the same.
- They never mock SQLite. The store is tested against the real engine.
- They never assert against Wails' generated bindings. The wire contract is
  `frontend/src/api.ts`, compared with the Go DTOs by a structural test; its
  receipt-line kinds are compared with the domain's by a second one. The first
  sees that a line carries a kind; only the second sees which kinds exist. Both
  arms were proved by planting a violation: a kind dropped from the page's union
  and a kind the page was never told about each failed by name, with the tree
  put back afterwards.

## What only a person can settle

These were measured by hand on the reference machine on 2026-09-22, against a
build run with a sandboxed `APPDATA`. They need doing again when the window
changes.

| Check | Last result |
|---|---|
| The window opens dark and the bar's button moves it to light and back, remembered across a restart. | Measured in Chromium against the built page on 2026-09-22: it opens dark with the sun and the words "Light mode", one press gives light with the moon and "Dark mode", then the choice is written to the window's storage. Not yet confirmed in the real window. |
| The setup program opens dark and its header toggle moves it to light and back. | Measured in Chromium against the setup page on 2026-09-22: dark with the sun, one press to light with the moon, the face and the words changing together and the choice stored. Not yet confirmed in the real setup window. |
| Wails' IPC works under the page's Content-Security-Policy. | Works: the severities shown came from Go. |
| `window.print()` opens the Windows print dialog. | It does, offering Save as PDF and Microsoft Print to PDF. |
| The printed page carries the receipt alone. | It does: no band, no status line, no controls. |
| The framing prints: the program and its address above the title and below the last event, with the statement under the range. | Not yet run in the real window. The lines are covered by the suite and by the page's own test; how WebView2 lays them out on paper, plus what a record spanning several pages does to the closing line, are readings only a print can give. They print once each, at the top and the bottom of the document, rather than repeating on every page. |
| Recording, editing, deleting, exporting and importing through the real window. | All worked; the import skipped the event already held. |
| The Guide and About open, scroll and close. | They do. |
| The Guide and About hold still for five seconds, then read themselves down; they step aside the moment the reader scrolls. | Not yet run in the real window. The cycle is covered by the suite and the pane's ring was measured in Chromium against the built stylesheet; neither is a reading of the running application. |
| Tab reaches the Guide's text without drawing a ring round it; Close rings when it is reached. | Measured in Chromium against the built stylesheet: the pane draws nothing at rest, hovered, clicked or Tab-focused, while Close draws the 2px ring. Not yet confirmed in the real window. |
| The three-state ring: nothing at rest, green on hover or focus, permanent red while disabled. | Measured in Chromium against the built stylesheet, both modes, with `:hover` and `:focus-visible` confirmed each time. Dark: at rest `3px none` on the band, Show and an enabled Print; hovered or Tab-focused `2px solid rgb(52,211,153)`; disabled Print and Delete `2px solid rgb(255,138,128)` on a `rgb(42,48,58)` fill. Light: hovered `2px solid rgb(4,120,87)`, disabled `2px solid rgb(168,35,26)` on `rgb(228,232,237)`. The disabled Print is skipped by Tab. Not yet confirmed in the real window. |
| The window opens with nothing ringed; the first Tab enters the band. | **Found broken 2026-09-22, cause measured, fix written, UNVERIFIED in the real window.** The owner reported that no Tab ever rang anything on a fresh build and install; one click on the page then made every Tab work. So the page never held the keyboard: WebView2 gives it DOM focus while the webview holds no keyboard focus, so no keydown reaches the listener. Chromium cannot show this and said so misleadingly: against the built page at 1100x760 the first Tab rang Record `rgb(52, 211, 153) solid 2px` correctly, while `document.hasFocus()` read false, which is the same defect wearing a pass. The first fix asked the Wails runtime to show the window and MISSED, because that focuses the main window while WebView2 hosts the page in a child window the keys follow. What is in the tree now is ported from PigeonPost: `internal/infrastructure/windowfocus` finds that child by its class and sets focus on it. **The next build is the measurement: open the window and press Tab without clicking first.** |
| Setup installs, with the options opening on what the machine already holds. | It does: files, registry entry, shortcuts, plus the application launched from the install folder. |
| Setup reopens on the manage screen when the versions match; a shortcut box applies immediately. | It does. |
| Setup started with `-uninstall`, as the Apps list starts it, opens on the removal screen. | It does. |
| Uninstall removes the program, its shortcuts, its log and the window's WebView2 folder, while keeping the record. | It does; the record was still there afterwards. |
| The Apps list's Uninstall and Modify point at a path that exists. | They do now. The first install wrote doubled separators; fixed and covered by a test. |
| The install folder is removed after setup exits. | Measured with a probe: the ported command removed nothing, so it was rewritten. Still to check by hand on a real install. |
| Uninstall with "also delete my symptom record" ticked. | Not yet run. |
| The Flatpak builds on a Linux machine and the window opens. | Built and run by the owner on 2026-09-22: it builds and the application runs. |
| The record lands under `~/.var/app/uk.codecrafter.SymChit` inside the Flatpak. | Not checked. The run says the window opens, not where the file went. |
| The DMG builds and notarizes on an Apple Silicon Mac. | Built and run by the owner on 2026-09-22, then notarized. |
| The notarisation ticket is stapled to the DMG. | Not separately checked. Confirm with `xcrun stapler validate` on the finished image; a notarized build with no ticket stapled still asks the network on a machine that is offline. |

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): what the structural tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [REQUIREMENTS.md](REQUIREMENTS.md): each requirement names the test that
  verifies it.
