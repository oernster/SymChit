# SymChit: Technical Debt

A standing reference to the project's outstanding technical debt. It records what is still open, weighs whether each item is worth doing and gives the rationale. Every item is a behaviour-preserving internal concern: nothing here proposes reverting a feature or changing any UI or UX behaviour. Scope is the whole repository (the Go core, the React front end, the setup program with its hand-written page and the delivery scripts) read against `ARCHITECTURE.md`, `TESTING.md` and the structural tests in `tests/structural`.

The sections below the open items are the standing record of what was weighed and deliberately left alone, so the same ground is not covered twice. They carry no numbers, because a number here means an open item.

---

## 1. Nothing checks the setup program's page beyond its syntax

`installer/frontend/dist` holds four hand-written files (`index.html`, `setup.css`, `setup-routes.js`, `setup-shell.js`) that are the setup program's entire user interface. They have no build step, so nothing compiles them, nothing lints them and no test drives them. The application's page has all three: `npm run build` runs `eslint`, then `tsc --noEmit`, then Vite; `vitest` runs 108 tests over it.

Three checks do reach the setup page today; they are worth naming so the gap is read accurately rather than as "unchecked":

- `tests/structural/boundary_test.go` holds it to the 400-line module limit.
- `tests/structural/cssscan_test.go` holds it to the ring rules and the scrolling rules.
- `test.ps1` runs `node --check` over each script, which parses it without running it. Proved to bite by planting a syntax error: the gate fails naming the file.

What none of them can see is a name that does not match. The page reaches into the DOM by string (`$('licence-open')`, `$('record-path')`) and reads fields off a state object Go marshals (`state.recordFile`, `state.installed`, `state.mode`). Rename an element id in the HTML or a field in the Go DTO and every check above still passes while the user gets a screen that does nothing. The house has already been bitten by exactly this shape elsewhere: a product name written into a setup page in sixteen places survived a rename that reached every other surface; setup went on announcing a product that no longer existed.

The cost is bounded by how small the page is and by the fact that the whole of it is exercised by hand every time an install is tested. It is not a defect; it is a class of defect that nothing would catch first.

Two ways to close it, not yet chosen. Either point the frontend's already-installed `eslint` at the directory with a small flat config, which catches an undeclared identifier and a typo in a name the page itself defines but not a mismatch against the HTML or the Go DTO; or give the setup page the same Vite build the application's page has, which costs a build step in the distribution's one-file story and buys type checking of the DTOs across the wire. The second is the fuller answer and the larger change.

## 2. The `installer` package has no tests

`go test ./...` reports `?   github.com/oernster/symchit/installer   [no test files]`. `App.Install`, `App.Uninstall`, `App.Survey` and the progress reporting around them are covered by nothing.

Most of what those methods decide has already been pushed down into `internal/infrastructure/setup`, which is where the tests are: the payload fence, the paths, the version comparison that picks the route, the shortcut writing and (since this pass) the folders an uninstall clears and the scheduled removal of the install folder. So the individual decisions are held. What is not held is the sequence: that an uninstall removes the shortcuts before the registry entry, that it refuses outright while the application is running, that the record goes after everything else and only when asked and that each step reports the progress the page is waiting on.

**Blocked on a seam that does not exist yet.** The package is `main`, the progress calls go through the Wails runtime and the methods call the Windows half of `setup` directly, so there is nothing to substitute. Closing it means the same move the rest of the repository has already made: the sequence stated over an interface, with the Wails calls behind it. That is a real refactor of the setup program's facade rather than a test to be written, which is why it is recorded rather than done.

## Looks like debt, not worth touching

**The setup page living under `dist` with no bundler.** It reads as generated output and is not; the structural tests name it explicitly for that reason. Keeping it hand written is what makes the setup program a single file with no build step of its own, which is the whole of the distribution story. Item 1 above is about checking it, not about bundling it.

**The setup package's coverage floor at 62.** The uncovered half is the registry writes, the process work and the Win32 handle calls, which change the machine they run on. A test double for them would assert that the double was called; the gate would rise on nothing. A real install exercises them and `TESTING.md` lists what a person has to look at.

**`internal/infrastructure/windowfocus` having two tests that only prove it returns.** It is `EnumWindows`, `AttachThreadInput` and `SetFocus` against a window the suite does not own. The tests exist to pin the one property a harness can see, that a best-effort function stays best effort when there is no window to find. The behaviour that matters is settled by opening the built window and pressing Tab.

## Not debt (do not "fix" these)

**`Unavailable` repeating the same refusal across every `Store` method.** It looks like a dozen copies of one line. Each one is a different method of the interface answering with the reason the record could not be opened, which is what keeps the window open and every action honest about why it cannot act (FR-062). Collapsing it would mean a smaller interface, not less code.

**`quotedPath` and `literal` both quoting a path.** They are not a duplicate pair. `quotedPath` writes plain double quotes for the registry, because that is what every other entry in the Apps list carries. `literal` writes a PowerShell single-quoted string, doubling any apostrophe, because the shortcut is created through a PowerShell one-liner. Go's `%q` was used for both once and shipped a broken install: it escapes the separators, so every backslash reached the registry and the shortcut doubled. Two quoting rules need two functions.

**The frozen export samples, one per format version.** `internal/infrastructure/export` keeps a committed sample for every version ever written and reads each one on every run. They are near-identical files by design: the point is that a change to the reader that breaks an old export fails here rather than at a user who is importing a record they made last year.

**`appVersion` being a `var` rather than a `const`.** `VERSION` is read by `build.ps1` into `-ldflags -X` against a `var` and by `builddmg.sh` into the bundle's plist. A `const` would take the flag silently and keep the old value, which is why it is a `var` and why it looks like a mutable global that should be a constant.
