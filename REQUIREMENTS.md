# SymChit: Software Requirements Specification

Status: baselined 2026-09-22. Derived the same day from
`health-symptom-tracker-requirements.md` (the source document, section 1.5).
Changes from here arrive as numbered amendments with a reason (section 6).
Two questions remain open (Appendix B); neither gates the domain or the
application layer.

## 1. Introduction

### 1.1 Purpose

SymChit records symptoms at the moment they are noticed. Later, it prints a
short factual record to take to a doctor. It is a recorder, not a
diagnostician: it keeps what the user observed and when; interpreting that is
for the user and their doctor.

> When you notice a symptom, record it. When you see your doctor, take the
> record with you.

### 1.2 Intended audience

The owner (Oliver Ernster), who is the developer and the first user, plus any
AI assistant working on the code. The repository is public under GPL-3.0.

### 1.3 Scope

In scope for version 1:

- Recording a symptom event: symptom, time, optional severity, optional note.
- Reusable symptom definitions, offered by autocomplete.
- A history of events, filtered by date range, symptom and severity.
- Editing and deleting events.
- A printable symptom receipt over a chosen date range.
- A machine-readable export of the whole record.
- A Windows desktop application in Go with Wails, plus the house setup program.
- The same application on Linux and macOS, packaged as a Flatpak and a signed
  DMG (Amendment 7).

Out of scope (decided by the source document, sections 10 and 15):

- Any diagnosis, suggested diagnosis or statement of what a symptom may mean.
- Any recommendation about medication or treatment.
- Any AI, generated text or interpretation of the user's observations.
- Health scores, risk scores, trend statements and colour-coded health ratings.
- Any claim that two events are related.
- Rewriting, normalising or translating the user's words into medical terms.
- Reminders, notifications or any prompt to record.
- Accounts, sign-in, cloud storage and synchronisation between machines.
- Advertising, analytics and telemetry.
- Event types other than symptoms (source section 14; see 3.8).
- A phone or web version in version 1 (platform decided 2026-09-22).
- A single-file setup program for Linux or macOS. Those two are packaged the way
  each platform expects, a Flatpak and a signed DMG; the bespoke setup program
  stays Windows-only (Amendment 7).

### 1.4 Definitions

| Term | Meaning |
|---|---|
| Symptom | A label the user supplies for what they noticed, such as `Tired`. Stored exactly as typed. |
| Symptom definition | A symptom kept for reuse, so it is offered the next time. |
| Event | One observation: a symptom at a point in time, with optional severity and note. The fundamental unit of data. |
| Occurred at | When the user says the symptom happened. Defaults to the recording instant; editable. |
| Recorded at | When SymChit created the event. Set once, never edited. |
| Severity | An optional user choice from a fixed list (FR-004). Never computed. |
| Note | Optional free text attached to an event, kept byte for byte as entered. |
| History | The list of events in the application, newest first. |
| Receipt | The printable symptom record for a date range (section 3.4). |
| Export | The machine-readable file holding the whole record (section 3.5). |
| Record | Everything SymChit stores: definitions plus events. |
| Local time | The time in the zone Windows is set to. |
| Reference machine | Oliver's desktop, Windows 11 build 26200. |

### 1.5 References

- Source document: `C:\Users\Oliver\Downloads\health-symptom-tracker-requirements.md`
  (read 2026-09-22). Section numbers cited as "source N".
- `C:\Users\Oliver\Development\WhatDay\REQUIREMENTS.md`: the shape of this document.
- `C:\Users\Oliver\Development\ed-voyage-companion`: the Go + Wails delivery
  reference (build, gate, setup program).
- WCAG 2.2 level AA.

## 2. Overall description

### 2.1 Product perspective

A new, standalone Windows desktop application. It has no server, no network
access and no data other than the record and its own settings.

### 2.2 User classes

One: the person whose symptoms are recorded, using their own Windows account.
A doctor reads the printed receipt and never uses the application. No
administrator rights are needed at any point.

### 2.3 Operating environment

Windows 11 on x64, with the WebView2 runtime that Windows 11 ships. Built and
tested on the reference machine.

Linux and macOS carry the same application (Amendment 7). Linux is a Flatpak on
the GNOME runtime, which supplies the webkit2gtk-4.1 that Wails renders through;
macOS is a signed and notarized DMG for Apple Silicon. Neither has been built on
its own platform yet: both are held by the gate compiling and vetting the whole
module for them on every run, which settles that nothing is Windows-only and
settles nothing else. TESTING.md says so in those words.

Where each platform keeps the two files it owns:

| | The record | The run log |
|---|---|---|
| Windows | `%APPDATA%\SymChit\symchit.db` | `%LOCALAPPDATA%\SymChit\SymChit.log` |
| Linux | `$XDG_CONFIG_HOME/SymChit/symchit.db` | `$XDG_STATE_HOME/SymChit/SymChit.log` |
| macOS | `~/Library/Application Support/SymChit/symchit.db` | `~/Library/Logs/SymChit/SymChit.log` |

Inside the Flatpak both variables already point at the sandbox, so the same
rules land under `~/.var/app/uk.codecrafter.SymChit`.

### 2.4 Constraints

- C-1 Language: Go 1.26 with cgo disabled; frontend React with TypeScript on Vite.
- C-2 Toolkit: Wails v2 (v2.12.0 measured on the reference machine).
- C-3 Storage: SQLite through `modernc.org/sqlite` (pure Go).
- C-4 Layering: `internal/{domain,application,infrastructure,ui}` with an
  explicit composition root; enforced by a structural test.
- C-5 The setup program is ported from ED Voyage Companion's `installer/`.
- C-6 Licence: GPL-3.0 (already in the repository).
- C-7 `VERSION` at the repository root is the single source of the version.
- C-8 The README header (the `# SymChit` title and the owner's opening line)
  stays verbatim; additions go beneath it.

### 2.5 Assumptions and dependencies

| ID | Assumption | Owner | Status |
|---|---|---|---|
| A-1 | One person's record is small: under 20,000 events over ten years (about five a day). Performance targets (3.6) are set against 20,000. | Oliver | To confirm |
| A-2 | Windows' clock is right; SymChit trusts it for the default occurrence time. | Oliver | To confirm |
| A-3 | Oliver supplies the application artwork as a master PNG with a transparent background. | Oliver | To confirm |
| A-4 | WebView2 in a Wails window can print the receipt through the Windows print dialog, including Microsoft Print to PDF. HYPOTHESIS: not yet measured (Appendix A, M-1). | Claude | Measure before FR-040 is built |

## 3. Requirements

Priorities use MoSCoW. Every requirement names the test or check that verifies
it; test names are provisional until the code exists.

### 3.1 Functional requirements: recording

**FR-001 Record now**
- Priority: Must
- Requirement: When the user presses `Record now` with a symptom entered, the
  recording service shall store one event whose occurrence time is the
  recording instant, unless the user has set another time (FR-003).
- Acceptance: Given the clock reads 2026-09-22T17:12:00+01:00, symptom `Tired`
  and note `Only been awake for about 10 minutes.`, when `Record now` is
  pressed, then the history's first entry reads 22 Sep 2026 17:12, `Tired`,
  with that note and no severity.
- Verified by: `internal/application/record_test.go::TestRecordNowUsesTheClock`

**FR-002 Symptom required**
- Priority: Must
- Requirement: If `Record now` is pressed with an empty symptom (nothing but
  white space), then the recording form shall store nothing and shall say that
  a symptom is needed.
- Verified by: `internal/domain/event_test.go::TestBlankSymptomRefused`

**FR-003 Set the occurrence time**
- Priority: Must
- Requirement: The recording form shall let the user replace the default
  occurrence time with another local date and time before recording.
- Rationale: Source 6: at 15:00 the user remembers a headache at about 12:30.
- Acceptance: Given the clock reads 15:00 and the user sets 12:30 the same
  day, when `Tired` is recorded, then the event's occurred-at is 12:30 and its
  recorded-at is 15:00.
- Verified by: `record_test.go::TestRetrospectiveTimeKept`

**FR-004 Optional severity**
- Priority: Must
- Requirement: The recording form shall offer severity as an optional choice
  of `Mild`, `Moderate` or `Severe`, with no choice selected by default.
- Verified by: `internal/domain/severity_test.go::TestSeverityIsOptional`

**FR-005 Optional note**
- Priority: Must
- Requirement: The recording service shall store the note exactly as entered,
  with no trimming, spelling correction or other change.
- Acceptance: A note of `  Much worse than yesterday.  ` (leading and trailing
  spaces) reads back identical, byte for byte.
- Verified by: `internal/infrastructure/store/store_test.go::TestNoteRoundTripsByteForByte`

**FR-006 Occurrence time never in the future**
- Priority: Must
- Requirement: If the user sets an occurrence time later than the current
  instant, then the recording form shall store nothing and shall say that an
  occurrence cannot be in the future.
- Verified by: `event_test.go::TestFutureOccurrenceRefused`

**FR-007 Recorded-at set once**
- Priority: Must
- Requirement: The recording service shall set recorded-at when the event is
  created; no later operation shall change it.
- Verified by: `record_test.go::TestEditLeavesRecordedAtAlone`

**FR-008 Form after recording**
- Priority: Must
- Requirement: When an event has been recorded, the recording form shall clear
  its symptom, note, severity and time, then place the keyboard focus on the
  symptom field.
- Rationale: The next observation starts from a clean form (source 3).
- Verified by: frontend test `RecordForm.test.tsx::clearsAfterRecord`

**FR-009 Recording failure**
- Priority: Must
- Requirement: If the store cannot write an event, then the recording form
  shall keep everything the user entered and shall say that the event was not
  saved, naming the reason.
- Verified by: `record_test.go::TestWriteFailureLosesNothing`

### 3.2 Functional requirements: symptom definitions

**FR-010 New symptom becomes a definition**
- Priority: Must
- Requirement: When an event is recorded with a symptom that matches no
  definition, the recording service shall create a definition holding the
  symptom exactly as typed.
- Verified by: `record_test.go::TestNewSymptomBecomesDefinition`

**FR-011 Autocomplete**
- Priority: Must
- Requirement: While the user types in the symptom field, the recording form
  shall list every definition containing the typed text, matched without
  regard to letter case, most recently used first.
- Acceptance: Given definitions `Tired`, `Headache` and `Back pain`, typing
  `he` lists `Headache` only; typing `a` lists `Headache` and `Back pain`
  (Amendment 1).
- Verified by: `internal/domain/definition_test.go::TestSuggestionsMatchIgnoringCase`

**FR-012 Free entry**
- Priority: Must
- Requirement: The recording form shall accept a symptom that is not in the
  list, without asking the user to confirm it.
- Verified by: `RecordForm.test.tsx::acceptsANewSymptom`

**FR-013 Labels kept as typed**
- Priority: Must
- Requirement: The application shall display every symptom exactly as the
  user typed it; no code path shall change its case, spelling or wording.
- Verified by: `store_test.go::TestSymptomRoundTripsByteForByte`

**FR-014 Near duplicates**
- Priority: Must
- Requirement: When the typed symptom differs from an existing definition only
  in letter case or surrounding spaces, the recording service shall use the
  existing definition, keeping the label as first typed (Q-2).
- Verified by: `definition_test.go::TestCaseVariantReusesDefinition`

### 3.3 Functional requirements: history, editing and deletion

**FR-020 History order**
- Priority: Must
- Requirement: The history shall list events by occurrence time, newest first,
  showing for each the local date and time, the symptom, the severity where
  given and the note where given.
- Verified by: `internal/application/history_test.go::TestNewestFirst`

**FR-021 Filters**
- Priority: Must
- Requirement: The history shall filter events by an inclusive local date
  range, by one or more symptoms and by one or more severities; filters
  combine, each narrowing the others.
- Acceptance: Given 12 events of `Tired` and 3 of `Headache` in September,
  filtering to `Headache` from 1 to 30 September lists 3.
- Verified by: `internal/domain/filter_test.go::TestFiltersCombine`

**FR-022 No analysis**
- Priority: Must
- Requirement: The history shall show events and counts only; it shall show no
  trend, average, score, chart or statement about what the events mean.
- Verified by: inspection against source 10.

**FR-023 Edit an event**
- Priority: Must
- Requirement: The application shall let the user change an event's symptom,
  occurrence time, severity and note.
- Verified by: `internal/application/edit_test.go::TestEditChangesOnlyWhatWasChanged`

**FR-024 Edit keeps the occurrence time**
- Priority: Must
- Requirement: When an event is edited without its occurrence time being
  changed, the edit service shall keep the occurrence time exactly as it was.
- Acceptance: Given an event occurring at 09:30, when its note is edited at
  14:00, then it still occurs at 09:30.
- Verified by: `edit_test.go::TestEditKeepsOccurredAt`

**FR-025 Delete an event**
- Priority: Must
- Requirement: When the user asks to delete an event, the application shall
  show a confirmation naming the event's symptom and local date and time; the
  event is removed only when the user confirms.
- Verified by: `internal/application/delete_test.go::TestDeleteNeedsConfirmation`,
  `History.test.tsx::confirmNamesTheEvent`

**FR-026 Delete many**
- Priority: Could
- Requirement: Where several events are selected, the delete confirmation shall
  state how many events will be removed.
- Verified by: `History.test.tsx::confirmStatesTheCount`

**FR-027 Edit and delete failure**
- Priority: Must
- Requirement: If the store cannot write an edit or a deletion, then the
  application shall leave the event as it was and shall say so, naming the
  reason.
- Verified by: `edit_test.go::TestFailedEditChangesNothing`

**FR-028 Rename a definition**
- Priority: Should
- Requirement: The application shall let the user rename a symptom definition;
  every event using it then shows the new name (Q-3).
- Verified by: `internal/application/definitions_test.go::TestRenameReachesEveryEvent`

### 3.4 Functional requirements: the receipt

**FR-040 Print a receipt**
- Priority: Must
- Requirement: When the user asks for a receipt over a local date range, the
  receipt service shall produce a page for the Windows print dialog, which
  also offers saving as PDF.
- Verified by: `internal/application/receipt_test.go` plus a manual print on
  the reference machine (A-4).

**FR-041 Receipt contents**
- Priority: Must
- Requirement: The receipt shall hold, in order: the title `SYMPTOM RECORD`;
  the date range; then, for each symptom recorded in the range, a heading of
  the symptom and its number of events, followed by each event's local date
  and time, severity where given and note where given.
- Acceptance: Given `Headache` on 25 Aug 12:30 and 2 Sep 08:10 plus `Tired`
  on 28 Aug 07:45 and 22 Sep 17:12, the receipt for 23 Aug to 22 Sep reads
  `Headache - 2 recorded events`, `25 Aug 12:30`, `02 Sep 08:10`, then
  `Tired - 2 recorded events`, `28 Aug 07:45`, `22 Sep 17:12`, each event
  followed by its severity and note where given.
- Event times leave out the year while the whole range lies in one year, as the
  source example does; a range crossing a year names the year on every event
  (Amendment 2).
- Verified by: `internal/domain/receipt_test.go::TestReceiptMatchesTheAcceptanceExample`,
  `TestReceiptAcrossAYearNamesTheYear`

**FR-042 Arithmetic only**
- Priority: Must
- Requirement: The receipt shall contain only what the user recorded plus
  counts of events; it shall contain no interpretation, trend, comparison or
  suggestion.
- Verified by: `receipt_test.go::TestReceiptHoldsNoOtherText`, which asserts
  every line of a generated receipt is a title, a range, a heading, a count or
  a field of a recorded event.

**FR-043 Empty range**
- Priority: Must
- Requirement: If the chosen range holds no events, then the receipt service
  shall say so and shall produce no receipt.
- Verified by: `receipt_test.go::TestEmptyRangeGivesNoReceipt`

**FR-044 Receipt order**
- Priority: Must
- Requirement: The receipt shall order symptom groups by each symptom's
  earliest event in the range, oldest first; within a group, events run oldest
  to newest. Names play no part in the order (Q-4).
- Rationale: The source list says "chronological" while its example runs
  newest first; the owner chose oldest first, never alphabetical.
- Verified by: `receipt_test.go::TestReceiptOrder`

### 3.5 Functional requirements: export

**FR-050 Export**
- Priority: Must
- Requirement: When the user asks to export, the export service shall write the
  whole record to a JSON file at a path the user chooses in a save dialog. The
  dialog shall open in the user's Downloads folder, which the user may leave
  (Amendment 4).
- Verified by: `internal/infrastructure/export/export_test.go::TestExportWritesEveryEvent`

**FR-051 Export format**
- Priority: Must
- Requirement: The export shall carry a format version plus every definition
  and every event with its symptom, occurred-at, recorded-at, severity and
  note; times are RFC 3339 with their UTC offset.
- Verified by: `export_test.go::TestFormatIsStable` against a committed sample
  file.

**FR-052 Export failure**
- Priority: Must
- Requirement: If the export file cannot be written, then the export service
  shall leave no partial file behind and shall say so, naming the path and the
  reason.
- Verified by: `export_test.go::TestFailedExportLeavesNothing`

**FR-053 Import**
- Priority: Should
- Requirement: The application shall read an export file back, adding its
  events to the record and skipping any event already present (Q-5).
- Verified by: `export_test.go::TestImportRoundTrip`

### 3.6 Functional requirements: storage and lifecycle

**FR-060 Local store**
- Priority: Must
- Requirement: The application shall keep the record in one SQLite file, in the
  folder the platform keeps a program's own configuration in: on Windows
  `%APPDATA%\SymChit\symchit.db` (Amendment 7 names the other two in 2.3).
- Verified by: `store/unavailable_test.go::TestOpensAtTheGivenPath`, which reads
  the folder from the platform through `os.UserConfigDir`

**FR-061 First run**
- Priority: Must
- Requirement: If no record file exists, then the application shall create an
  empty one without reporting a fault; the recording form is then ready.
- Verified by: `store_test.go::TestMissingFileGivesEmptyRecord`

**FR-062 Unreadable record**
- Priority: Must
- Requirement: If the record file exists but cannot be opened, then the
  application shall open its window, say that the record could not be read
  and name the file with the reason. It shall not overwrite the file.
- Rationale: House robustness rule 1: nothing before the window may end the run.
- Verified by: `store_test.go::TestCorruptFileIsNeverReplaced`

**FR-063 Schema versions**
- Priority: Must
- Requirement: The store shall record its schema version; when it opens a file
  of an older version, the store shall upgrade it in one transaction.
- Verified by: `store_test.go::TestUpgradeFromEveryEarlierVersion`

**FR-064 Single instance**
- Priority: Must
- Requirement: When SymChit starts while another instance is running for the
  same user, the new instance shall bring the running window forward and exit.
- Verified by: `internal/infrastructure/instance/lock_test.go`

**FR-065 Log**
- Priority: Must
- Requirement: The application shall point standard error at a log in the folder
  the platform keeps a program's own state in, as its first act: on Windows
  `%LOCALAPPDATA%\SymChit\SymChit.log` (Amendment 7 names the other two in 2.3).
  The log shall never contain a symptom, a note or any other part of the record.
- Note: pointing standard error at a file is Windows-only work, because the
  failure it answers is Windows-only: a windowed run started from a shortcut is
  handed a handle of 0 and everything written there is lost. Elsewhere the run
  keeps its own standard error and the crash report is copied to the log.
- Verified by: `runlog_test.go`, including
  `TestEveryPlatformsLogFolder`; `tests/structural/boundary_test.go::TestNoRecordFieldsLogged`

**FR-066 About**
- Priority: Must
- Requirement: The application shall offer an About dialog stating the
  product name, the version, the author, the copyright notice and the
  open-source works it is built with, each with its licence, plus the statement
  in FR-067. The notice shall read `Copyright © Oliver Ernster 2026`
  (Amendment 8).
- Note: the year is the year of the first release, not the year the program is
  run in. A notice that follows the clock claims a date nothing was published
  on.
- Verified by: `internal/application/about_test.go`;
  `frontend/src/App.test.tsx::opens the Guide and About`

**FR-068 Guide** (Amendment 5)
- Priority: Must
- Requirement: The application shall offer a Guide naming every control with
  the picture that control draws, stating how to record quickly, what SymChit
  will not do and how the record is kept.
- Rationale: The house guide, as PigeonPost and ClearBudget carry one.
- Verified by: `frontend/src/App.test.tsx::opens the Guide and About`

**FR-069 Donation** (Amendment 6)
- Priority: Should
- Requirement: The application shall offer a button in the bar that hands a
  donation address to the desktop for the user's browser to open. The
  application shall not itself fetch that page; the address shall be stated in
  one place in the application.
- Rationale: The house donate button. SymChit is free and stays free: nothing
  is withheld behind a donation, so the ask is a postscript rather than a
  prompt.
- Acceptance: Pressing Donate asks the desktop for exactly one address, that
  address is `https://www.paypal.com/ncp/payment/4XP3AYNMPQGUC`; no connection
  is opened by SymChit itself.
- Verified by: `donate_test.go`;
  `frontend/src/App.test.tsx::offers the donation page last in the band`;
  `tests/structural/boundary_test.go::TestNoNetworkImports`

**FR-067 Not medical advice**
- Priority: Must
- Requirement: The About dialog shall state that SymChit records observations
  only and gives no medical advice.
- Verified by: `about_test.go::TestAboutCarriesTheStatement`

### 3.7 Functional requirements: setup program

**FR-070 Per-user install**
- Priority: Must
- Requirement: The setup program shall install SymChit to
  `%LOCALAPPDATA%\Programs\SymChit` without requesting administrator rights.
- Verified by: manual install on the reference machine.

**FR-071 Routes**
- Priority: Must
- Requirement: The setup program shall offer install, update, going back to an
  older version, repair and uninstall, each on its own screen, registered in
  the Windows Apps list.
- Verified by: manual; each route exercised once.

**FR-072 Uninstall keeps the record**
- Priority: Must
- Requirement: When uninstalling, the setup program shall remove the program
  files, shortcuts and registry entries; it shall remove the record only where
  the user ticks a box naming the record file.
- Rationale: The record is the user's (source 12); losing it by removing the
  program would be accidental loss (source 13).
- Verified by: manual; folders inspected afterwards.

### 3.8 Non-functional requirements

**NFR-USE-001 Recording speed**: With SymChit's window open, recording an event
of an existing symptom with no note shall take no more than 4 keystrokes after
the first letters of the symptom: choose the suggestion, then press Enter to
record. Verified by a frontend test driving the keys.

**NFR-USE-002 Keyboard**: Every action shall be reachable by keyboard alone
(Amendment 8). Tab and Right shall step the focus ring forward and Shift+Tab and
Left shall step it back, both wrapping at the ends; Enter and Space shall fire
the focused control; Escape shall close an open dialog. A field holding text
keeps the horizontal arrows for its caret and is left with Tab. The window shall
open with nothing focused; a dialog shall open on its first control, passing
over its scrolling body. Verified by `frontend/src/ring.test.ts` and
`useRing.test.tsx`, plus a manual pass.

**NFR-USE-004 Focus ring** (Amendment 8): A control shall show no ring at rest,
the ring colour while it is hovered or keyboard-focused, then the danger colour
permanently while it is disabled; a disabled control's fill shall be muted so
the ring reads against it. No container shall take focus or paint a ring. The
accent colour shall never be used as a ring. Verified by
`tests/structural/focus_test.go`, each assertion proved by planting.

**NFR-USE-003 Contrast**: Text shall meet WCAG 2.2 AA contrast (4.5:1 for body
text) in the theme in use; each ring colour shall meet the 3:1 WCAG 2.2 asks
of a non-text indicator against the surfaces it is drawn on. Verified by a test
over the colour tokens.

**NFR-PERF-001 Startup**: The recording form shall accept input within 2 s of
process start on the reference machine, measured from the log's start line to
its ready line.

**NFR-PERF-002 History**: With 20,000 events (A-1), the history shall show its
first page within 300 ms of a filter change, at the 95th percentile over 100
changes, measured by a benchmark against a generated store.

**NFR-PERF-003 Receipt**: A receipt over 1,000 events shall reach the print
dialog within 2 s on the reference machine.

**NFR-PRIV-001 No network**: The application and the setup program shall open
no network connection. Verified by `tests/structural/boundary_test.go::TestNoNetworkImports`
(forbids `net` and every `net/` package in the repository's own Go code) plus a
Content-Security-Policy of `default-src 'self'; connect-src 'none'` on the
frontend, checked by a test reading `index.html`.

**NFR-PRIV-002 No telemetry**: No component shall collect or send usage data.
Verified by the same boundary test plus a dependency review in `go.mod` and
`package.json`.

**NFR-PRIV-003 Encryption at rest**: The record is not encrypted at rest; it is
protected by the Windows account only. This non-claim shall be stated in the
README (Q-6).

**NFR-REL-001 Crash safety**: An event reported as recorded shall survive a
power loss one second later. The store runs SQLite in WAL mode with
`synchronous=FULL`. Verified by `store_test.go::TestDurabilityPragmas`.

**NFR-MAINT-001 Coverage**: `internal/domain` and `internal/application` shall
be held at 100% statement coverage by `test.ps1`, which `build.ps1` runs first
and cannot skip.

**NFR-MAINT-002 Structure**: A structural test shall enforce the layering
(C-4), domain purity (no `os`, `time.Now` or `database/sql` in the domain), the
400-line module cap and its 381 to 399 danger band.

**NFR-MAINT-003 Checks**: `gofmt`, `go vet`, `staticcheck`, `tsc --noEmit`,
`eslint` and Vitest shall report nothing.

**NFR-MAINT-004 Docs**: The repository shall carry `README.md` (with who it is
for and not for), `ARCHITECTURE.md`, `TESTING.md`, `DEVELOPMENT.md` and
`VERSION`.

**NFR-MAINT-005 Wire**: A structural test shall compare the Go DTOs with the
hand-written TypeScript interfaces, field for field.

### 3.9 Won't this time

| Item | Reason |
|---|---|
| Other event types (medication, meal, sleep) | Source 14: version 1 stays on symptoms. The event table carries a kind column so they can follow without a migration of existing rows. |
| Phone or web version | Platform decided 2026-09-22: Windows desktop. |
| Synchronisation between machines | Source 11: no cloud. The export is the portability path. |
| Reminders | Source 10: no nagging. |
| Charts and trends | Source 10: no interpretation. |
| Edit history (audit trail) | Source 13 allows one; left out of version 1 (Q-7). |
| Update check | No network (NFR-PRIV-001). |

## 4. Other requirements

### 4.1 Legal and regulatory

GPL-3.0. SymChit makes no medical claim and performs no clinical function; the
design keeps it outside what UK MHRA guidance treats as software as a medical
device, which turns on intended medical purpose such as diagnosis or treatment.
HYPOTHESIS: this reading of the guidance has not been checked against the
current text; it is Q-8.

### 4.2 Internationalisation

English only in version 1. Dates print in the day-month order of the source
example (`22 Sep 17:12`).

### 4.3 Risk

Two harms matter: loss or silent corruption of the record; a receipt that
misstates it to a doctor. Rather than a full FMEA, three
requirements carry it: FR-024 (time never silently changed), FR-042 (receipt
holds nothing but the record) and NFR-REL-001 (durability). Judged
proportionate for a single-user record keeper.

## 5. Appendices

### Appendix A: Feasibility measurements

| ID | Question | Measured |
|---|---|---|
| M-1 | Does `window.print()` in a Wails v2.12 window open the Windows print dialog with Microsoft Print to PDF offered? | Not yet measured. A throwaway probe before FR-040 is built. |

### Appendix B: Open questions

Decided by the owner on 2026-09-22:

| ID | Question | Decision | Lives in |
|---|---|---|---|
| Q-1 | Severity list | `Mild`, `Moderate`, `Severe`. | FR-004 |
| Q-2 | `tired` typed when `Tired` exists | Same symptom; the existing label is kept. | FR-014 |
| Q-3 | Rename a definition | Yes, Should. | FR-028 |
| Q-4 | Receipt order | Groups by first occurrence, oldest first; events oldest to newest; never alphabetical. | FR-044 |
| Q-5 | Import in version 1 | Yes, Should. | FR-053 |
| Q-6 | Encryption at rest | No; the non-claim is stated. | NFR-PRIV-003 |
| Q-7 | Edit history | Not in version 1. | 3.9 |
| Q-10 | Identifying details on the receipt | None. | FR-041 |

Still open:

| ID | Question | Proposal | Owner | Gates |
|---|---|---|---|---|
| Q-8 | Does the MHRA reading in 4.1 hold? | Check the current guidance text before any public release. | Oliver | Release |
| Q-9 | Light, dark or following Windows? | Follow the Windows app mode. | Oliver | Build step 4 |

### Appendix C: Build order

1. Domain: event, symptom definition, severity, occurrence rules (FR-002,
   FR-006), suggestion matching, filters, receipt model.
2. Application: record, edit, delete, history, receipt, export, with fakes.
3. Infrastructure: SQLite store with schema versions, export file, run log,
   single-instance lock.
4. UI: Wails facade plus the React frontend; M-1 probe before the receipt page.
5. Setup program, ported from ED Voyage Companion.
6. Docs: README, ARCHITECTURE, TESTING, DEVELOPMENT.

## 6. Amendments

| No. | Date | Requirement | Change | Reason |
|---|---|---|---|---|
| 8 | 2026-09-22 | NFR-USE-002, new NFR-USE-004, FR-066 | The house keyboard model and its three-state focus ring are stated as requirements rather than left to the page; About names the copyright holder and year. | Owner's request. The ring was a single blue outline on keyboard focus alone, which said nothing about what could be used and nothing about what could not: Print sat inert beside the button that fills it with no way to tell it apart from a control waiting to be pressed. |
| 7 | 2026-09-22 | 1.3 scope, 2.3 operating environment, FR-060, FR-065 | Linux and macOS leave the deferred list and become part of this version: a Flatpak and a signed DMG, ported from PigeonPost's. Only the bespoke setup program stays Windows-only. | Owner's decision, superseding Amendment 3, which had made them a later version. Measured: one package stopped the module building elsewhere, the run log, whose Windows handle work is now behind a build tag and whose folder rule is a pure function taking the platform as an argument, so all three answers are exercised on every platform. The gate now builds and vets for Linux and macOS on every run, which is what keeps this true; a planted Windows-only import was refused by name. |
| 6 | 2026-09-22 | New FR-069 | The bar carries a Donate button, last in its right-hand group. | Owner's request. It takes a seat in the bar the window already has rather than a band of its own, as AudioDeck's does, since a whole new strip of chrome carrying one control costs more than it buys. The address lives once, in Go's product package; the page asks for the donation page rather than naming one, so nothing arriving from the page has to be checked before it is opened. |
| 5 | 2026-09-22 | New FR-068 | The application carries a Guide, reached from the bar. | Owner's request, in line with PigeonPost and ClearBudget. Its words are one document (`frontend/src/guideContent.ts`) and the dialog only draws them, as PigeonPost's does. |
| 4 | 2026-09-22 | FR-050, FR-053 | Both file dialogs open in the user's Downloads folder, which the user may leave. | Owner's request: it is where a person already looks for files they have saved. Measured: Wails' dialog options carry a default directory; the folder is checked before it is named, so a machine without one falls back to the dialog's own choice. |
| 3 | 2026-09-22 | 1.3 scope | Linux and macOS leave the permanent out-of-scope list and become planned for a later version: a Flatpak and a DMG builder, following PigeonPost's. | Owner's decision. Version 1 stays Windows only; nothing in the domain, application or page is Windows-specific, so the work is packaging plus the run log's standard-handle code. |
| 2 | 2026-09-22 | FR-041 | Receipt events name their year when the range crosses a year; they leave it out otherwise. | Found while writing the tests: the source example's `22 Sep 17:12` is unambiguous inside one year and ambiguous across two, which a doctor reading the record cannot resolve. |
| 1 | 2026-09-22 | FR-011 | The acceptance example's second case is corrected: typing `a` lists `Headache` and `Back pain`, not all three. | The baselined example was wrong: `Tired` holds no letter `a`. Found by the test that was written from it. |
