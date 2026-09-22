# SymChit: architecture

SymChit records observations and retrieves them. Everything below follows from
one rule: **record what happened; do not decide what it means.**

## The invariants

Each is enforced by a test, named beside it. Every one was proved to bite by
planting a violation and reading the exit code.

| # | Invariant | Enforced by |
|---|---|---|
| 1 | The domain depends on nothing: no other layer, no framework. | `tests/structural/boundary_test.go::TestDomainHasNoOutwardImports` |
| 2 | The domain performs no IO and never reads the wall clock. Time arrives as an argument. | `TestDomainIsPure` |
| 3 | The application layer depends on the domain and on the ports it declares, never on infrastructure or Wails. | `TestApplicationDoesNotImportInfrastructure` |
| 4 | Only `main.go` wires infrastructure to the application. | `TestCompositionRootIsWhitelisted` |
| 5 | No source file exceeds 400 lines; none sits in the 381 to 399 danger band. | `TestNoFileExceedsLineLimit`, `TestNoFileInDangerBand` |
| 6 | SymChit opens no network connection: no `net` package anywhere, `connect-src 'none'` on the page. | `TestNoNetworkImports`, `TestThePageOpensNoConnection` |
| 7 | The wire is stated twice, in Go and in TypeScript; the two agree field for field. | `tests/structural/wire_test.go` |
| 8 | Every colour lives in `frontend/src/theme.css`; every text pairing meets WCAG 2.2 AA in both modes. | `tests/structural/colours_test.go` |
| 9 | Every exported type carries a doc comment. | `TestEveryExportedTypeIsDocumented` |
| 10 | The ring belongs to a control: no container carries a ring rule or a tabindex; a surface made to scroll turns the engine's own ring off. | `tests/structural/focus_test.go::TestNoRingRuleNamesAContainer`, `TestNoContainerTakesFocus`, `TestEveryScrollingSurfaceSuppressesTheNativeRing` |
| 11 | A dialog body that scrolls pins its action row beneath it and wears the self-reading cycle. | `TestEveryScrollingDialogPinsItsActionsAndReadsItself` |
| 12 | A control that rings while it is usable says so while it is not: the ring is green on hover or focus, permanently red while disabled. | `TestEveryRingedControlSaysWhenItIsInert` |
| 13 | Each ring colour meets the 3:1 WCAG asks of a non-text indicator, in both modes. | `tests/structural/colours_test.go::TestEveryRingIsVisibleAgainstWhatItIsDrawnOn` |

## The layers

```text
UI (root package, frontend/)  →  Application  →  Domain  ←  Infrastructure
```

### Domain: `internal/domain`

The rules, as pure types and functions. No IO, no clock, no framework.

| File | What it owns |
|---|---|
| `symptom.go` | The symptom label and its comparison key: how a typed name is matched to one already held, without ever changing the label. |
| `severity.go` | The three severities and the absence of one. Names live here once. |
| `event.go` | The event, the future-occurrence rule and what makes two events the same observation. |
| `definition.go` | Reusable symptoms: matching, suggestion order and the rename checks. |
| `dates.go` | Civil dates, the two wire time formats and the display format. |
| `filter.go` | The history filter and the newest-first order. |
| `receipt.go` | The receipt: grouping, ordering and the exact lines that print. |

The receipt is domain code because its wording is a rule, not decoration: it is
what stops a printed record saying anything the user did not record. The page
receives lines with a kind and draws them; it composes no sentences of its own.

### Application: `internal/application`

One service per user-visible action, over the ports in `ports.go`.

| File | What it owns |
|---|---|
| `ports.go` | `Store`, `RecordFile` and `Clock`: everything the layer needs from outside. |
| `record.go` | Recording, including which instant an event gets. |
| `edit.go` | Editing, the deletion confirmation's wording and deletion. |
| `history.go` | Listing, suggesting, the symptom list, renaming and building the receipt. |
| `transfer.go` | Export; import too, skipping what is already held. |
| `about.go` | The About content and the credits. |
| `errors.go` | The refusals, each wrapped round the reason it happened. |

### Infrastructure: `internal/infrastructure`

| Package | What it implements |
|---|---|
| `store` | The SQLite record, one transaction per write, plus `Unavailable`, which stands in when the file could not be opened so that every action still answers. |
| `export` | The JSON export file: a versioned format, an atomic write and a distrustful read. |
| `runlog` | The run log, plus pointing the process's error output at it before anything can fail. Ported from WhatDay. |
| `setup` | The per-user install policy: the paths, the fenced payload extraction, the version comparison, the registry entry, the shortcuts and the process work. The setup program is a facade over it and owns no install logic. |

### UI: the root package and `frontend/`

`app.go` and `actions.go` are the facade: one bound method per action, each
converting shapes, calling one service and converting back. `dto.go` holds the
wire shapes. `window.go` holds the file dialogs and the single-instance lock.
`main.go` is the composition root.

The page in `frontend/src` is a client of the facade through `api.ts` and
nothing else.

### The setup program: `installer/`

A second Wails application in the same module, carrying the built application
as an embedded zip, so one file is the whole distribution. `installer/main.go`
is its composition root and `installer/app.go` a facade over
`internal/infrastructure/setup`; its page is hand-written and holds no product
name of its own, because a page has no build step to catch a stale one.

It follows the house setup model: work moves to a progress screen rather than
greying the options in place, the footer is rebuilt per screen, the progress
screen offers nothing, one reading of the machine decides the route; every path ends in a verdict. Setup follows the Windows light or dark setting with no
toggle, as the application does.

## Decisions and why

| Decision | Why | What it costs |
|---|---|---|
| The receipt's lines are built in the domain, not the page. | The one thing SymChit must never do is add words to a medical record. A test asserts every line is a title, a range, a heading, a count or a recorded field. | The page cannot reflow a line; it styles by kind. |
| The store resolves a symptom by key, creating it when absent. | Recording an event and creating its symptom is one transaction, so a failure leaves neither. | The store holds a key column the domain computes. |
| An edit sends no time unless the user changed it. | An empty time means "keep what is held", so an edit cannot move an occurrence to the moment of the edit. The rule is structural rather than remembered. | The edit form compares before sending. |
| A record that will not open becomes `store.Unavailable`. | The window opens and says what is wrong, instead of a program that never appears. Every action answers with the same reason. | Eleven one-line methods that refuse. |
| Times are stored as RFC 3339 with their offset. | The instant reads back as the same instant; an export carries the offset it was written in. | Ordering happens in Go rather than in SQL. |
| No encryption at rest. | A passphrase is a thing to lose; the Windows account already guards the file. Stated in the README rather than assumed. | Anyone who can sign in as the user can read the record. |
| Light and dark follow Windows, with no switch. | One fewer setting; the palette is held to AA in both modes by a test. | No preference for a user who wants the other one. |
| The Guide and About read themselves down, gently, until the reader takes over. | Long help holds still on open, descends a pixel every second tick, holds at the tail and rewinds; any wheel, press, key or focus arrival suspends it for 2.5 seconds and it then resumes from wherever the reader left it. The pace belongs to the application rather than to either dialog. | A timer per open dialog, plus a pure state machine to keep the pacing testable without waiting. |
| The focus ring is answered by the page, not left to the browser. | The browser has an opinion about Tab and none about the arrows, so the house model (Tab and Right forward, Shift+Tab and Left back, wrapping at both ends) has to be stated. It is split in two: the rules are a pure module under test; one listener drives them against the page. | One key listener at the shell, plus a text field that has to be asked for its arrows back rather than assumed. |
| Three ring states and no more. | Nothing at rest, so the window is quiet until it is used; green while a control is hovered or focused, because both say "you can use this" and a reader should not have to learn two colours for one fact; permanently red while disabled, because the red IS the state and a ring that waited for the mouse would leave Print looking like a button nobody had pressed yet. The accent is data meaning and never a ring. | A disabled control has to give up its fill as well; otherwise the ring it is meant to show disappears into it. |
| Only the run log knows which platform it is on. | Everything else was already portable: the record's folder comes from `os.UserConfigDir`, the page is a page; SQLite is pure Go. The run log answers a Windows-only failure, a windowed run handed a standard error handle of 0, so that half sits behind a build tag; its folder rule is a pure function taking the platform as an argument, so all three answers are exercised wherever the suite runs. | Two small files instead of one, plus a rule stated rather than read from the machine it runs on. |
| The Flatpak is given no network permission. | SymChit opens no connection; the sandbox is where that claim stops being a claim: an application that started talking to something would fail at run time rather than quietly working. The build gets the network, because it fetches Go modules and npm packages. | The manifest has two permission lists that must not be confused for each other. |
| The donate address lives in Go and the page never names one. | The page asks for the donation page; Go holds the only copy of the address and hands it to the desktop. Nothing arrives from the page, so there is no address to validate before opening; the no-network guarantee is untouched because SymChit fetches nothing. | One more bound method, plus a seam over Wails' opener so no test opens a browser. |
| The Donate button takes a seat in the bar rather than a band of its own. | The window already has a tray of icon buttons and no footer, so a second strip carrying one control costs more than it buys. It is drawn at its neighbours' height: a member sized smaller than the row it sits in reads as a mistake. | The mark keeps its own width, so one rule sits beside the band's square icons. |
| A scrolling dialog body stays a keyboard stop and paints nothing. | It carries no controls of its own, so a reader who never touches the mouse must be able to reach it and scroll it; a ring round a whole page of words marks nothing to act on. Measured in Chromium: an overflowing container is focusable with no tabindex and drew the engine's own ring, so the ring is turned off explicitly. | One suppression rule, held by a test, rather than the absence of a rule. |

## The export format

The file a user exports is their own copy of their record, so SymChit has to go
on reading it after the format has moved on. That is a promise about every
version ever written, not only the current one.

Every file carries an envelope that never changes shape: `format`, always
`symchit-record`, then `version`, an integer. The envelope is read on its own
first; the version chooses which reader reads the rest. Reading the whole
file into one struct and hoping it fits is the thing a versioned format exists
to avoid.

| The file says | What happens |
|---|---|
| A different `format` (or nothing that parses as JSON) | Refused: not a SymChit export |
| No `version` (or one below 1) | Refused. An unversioned file is not one SymChit wrote; a record is not the thing to be generous about |
| A version SymChit knows | Read by that version's own reader |
| A version above the current one | Refused, naming the version, so the user knows a newer SymChit wrote it |

Adding a version is three things, all of which the suite refuses the change
until you have done: raise `formatVersion`, add a reader to the table in
`versions.go`, then commit a sample of the **old** version under `testdata`. The
sample is the bytes that version really wrote, frozen and never edited again: a
reader tested against the current struct is only tested against itself.

Both halves of that guard were proved by planting. Raising the version alone
fails with "format 2 can be written but not read"; adding the reader without the
sample fails naming the file to commit.

## Execution

1. `main` points the error output at the log, so a crash leaves a record.
2. It opens the record, else carries the failure as a problem the window shows.
3. It builds the four services over the store, the export format and the clock.
4. It hands the facade to Wails, with the single-instance lock.
5. Every page action is one bound method: convert, call one service, convert
   back. A panic inside one becomes an error the page shows, with the stack in
   the log.
6. On shutdown the record is closed.

## Further reading

- [REQUIREMENTS.md](REQUIREMENTS.md): the requirements these invariants serve.
- [TESTING.md](TESTING.md): how it is verified.
- [DEVELOPMENT.md](DEVELOPMENT.md): how it is built.
