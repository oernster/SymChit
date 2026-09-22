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
