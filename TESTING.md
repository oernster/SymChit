# SymChit: testing

## The gate

```powershell
./test.ps1
```

It runs these in order, stopping at the first failure:

1. `gofmt -l` over the Go, ignoring the front end.
2. `go vet` over every package but `node_modules`.
3. `staticcheck`, fetched with `go run honnef.co/go/tools/cmd/staticcheck@latest`.
4. The whole Go suite.
5. Coverage of `internal/domain` and `internal/application`, which must be 100%.
6. Coverage of every other Go package against its measured floor.
7. The front end: `eslint`, `tsc --noEmit` and `vite build`.
8. The front-end suite: Vitest over jsdom.

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
| `internal/infrastructure/runlog` | 74 | The rest is Win32 standard-handle work, reachable only in a windowed process with no error output. |
| root package (the facade) | 77 | The facade itself is covered. `main`, the log handover, the file dialogs and the single-instance lock need a real window. |

Not gated at all: `internal/product` holds two constants; `tests/structural` is itself the guard.

## What each suite proves

| Suite | What it settles |
|---|---|
| `internal/domain` | The rules: blank and future occurrences refused, case variants matched, filters combined, the receipt's grouping, ordering and exact lines, including a range that crosses a year. |
| `internal/application` | The use cases over fakes: recording reads the clock once, an edit leaves recorded-at alone, a failed write loses nothing the user typed, an import skips what is held. |
| `internal/infrastructure/store` | Real SQLite in a temporary folder: notes and symptoms read back byte for byte, a deletion is all or nothing, a garbage file is never overwritten, a newer schema is refused, the durability pragmas are set, plus a planted trigger proving a failed write leaves nothing behind. |
| `internal/infrastructure/export` | The format against a committed sample, a refusal for anything that is not a SymChit export, plus a size cap read before the file is. |
| `internal/infrastructure/runlog` | Crash reporting, by starting this test binary again as a child and making it panic. |
| root package | The facade end to end over a real record: every conversion, every refusal, plus a panic in a bound method becoming an error rather than a dead window. |
| `tests/structural` | The invariants in ARCHITECTURE.md. |
| `frontend` | The page over a fake facade: the keyboard path to a recorded event, the form keeping everything when a save is refused, an edit sending no time unless it changed, the confirmation naming what will go, plus the receipt showing exactly the lines it was given. |

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

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): what the structural tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [REQUIREMENTS.md](REQUIREMENTS.md): each requirement names the test that
  verifies it.
