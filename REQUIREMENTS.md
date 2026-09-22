# SymChit: Software Requirements Specification

Status: baselined 2026-09-22. Derived the same day from
`health-symptom-tracker-requirements.md` (the source document, section 1.5).
Changes from here arrive as numbered amendments with a reason (section 6).
One question remains open (Appendix B); it gates neither the domain nor the
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
macOS is a signed and notarized DMG for Apple Silicon. Both have now been built
on their own platform and run: the Flatpak's window opens and the DMG is
notarized. The gate adds to that on every run by compiling and vetting the whole
module for all three, which settles that nothing Windows-only can be written
without the next run saying so. What each platform run did and did not settle is
in TESTING.md.

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
| A-4 | WebView2 in a Wails window can print the receipt through the Windows print dialog, including Microsoft Print to PDF. | Claude | Measured and true (Appendix A, M-1) |

## 3. Requirements

Priorities use MoSCoW. Every requirement names the test or check that verifies
it. Each name was read off the tree rather than written from intent: a reference
that does not resolve is a requirement nobody can check.

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
- Verified by: `internal/infrastructure/store/store_test.go::TestNoteAndSymptomRoundTripByteForByte`

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
- Verified by: `edit_test.go::TestEditLeavesRecordedAtAlone`

**FR-008 Form after recording**
- Priority: Must
- Requirement: When an event has been recorded, the recording form shall clear
  its symptom, note, severity and time, then place the keyboard focus on the
  symptom field.
- Rationale: The next observation starts from a clean form (source 3).
- Verified by: `frontend/src/RecordPane.test.tsx`, "records what was entered and clears afterwards"

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
- Verified by: `frontend/src/RecordPane.test.tsx`, "records what was entered and clears afterwards", which types a symptom no definition holds

**FR-013 Labels kept as typed**
- Priority: Must
- Requirement: The application shall display every symptom exactly as the
  user typed it; no code path shall change its case, spelling or wording.
- Verified by: `store_test.go::TestNoteAndSymptomRoundTripByteForByte`

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
- Verified by: `internal/application/edit_test.go::TestEditKeepsOccurredAt`, plus `frontend/src/EditRow.test.tsx`, "opens with what the event holds"

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
- Verified by: `internal/application/edit_test.go::TestDeleteNeedsConfirmation`,
  `frontend/src/HistoryPane.test.tsx`, "asks before deleting, naming the event; deletes only when confirmed"

**FR-026 Delete many**
- Priority: Could
- Requirement: Where several events are selected, the delete confirmation shall
  state how many events will be removed.
- Verified by: `frontend/src/HistoryPane.test.tsx`, "counts the events chosen for a bulk deletion"

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
- Verified by: `internal/application/history_test.go::TestRenameReachesEveryEvent`

### 3.4 Functional requirements: the receipt

**FR-040 Print a receipt**
- Priority: Must
- Requirement: When the user asks for a receipt over a local date range, the
  receipt service shall produce a page for the Windows print dialog, which
  also offers saving as PDF.
- Verified by: `internal/application/history_test.go::TestReceiptFromTheStore` plus a
  manual print on the reference machine (A-4, measured).

**FR-041 Receipt contents**
- Priority: Must
- Requirement: The receipt shall hold, in order: the title `SYMPTOM RECORD`;
  the date range; then, for each symptom recorded in the range, a heading of
  the symptom and its number of events, followed by each event's local date
  and time, severity where given and note where given. The fixed framing of
  FR-045 sits around and within this order, which it does not otherwise
  disturb.
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
- Requirement: The receipt shall contain only what the user recorded, counts of
  events and the fixed framing of FR-045; it shall contain no interpretation,
  trend, comparison or suggestion.
- The framing is fixed text that names no event and changes with no event, so
  it states nothing about what was recorded. Anything that varies with the
  record belongs to FR-041 and is held by this requirement.
- Verified by: `receipt_test.go::TestReceiptHoldsNoOtherText`, which asserts
  every line of a generated receipt is a title, a range, a heading, a count, a
  field of a recorded event or one of the two framing lines.

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

**FR-045 Printed framing**
- Priority: Must
- Requirement: The receipt shall carry, above the title and again below the
  last event, a line naming the program that produced it and its home address.
  Immediately below the date range it shall carry a statement that the record
  is not a diagnosis but the person's own notes, printed for a healthcare
  professional.
- Acceptance: Given the receipt of FR-041's example, its first line reads
  `Generated by SymChit (https://symchit.com)`, its fourth line reads `This
  record is not a diagnosis. It is one person's own notes of what they observed
  and when, printed for a healthcare professional to read.` and its last line
  repeats the first.
- The wording avoids saying the record is guidance for a professional to use.
  MHRA guidance v1.10f treats software that provides information to help a
  healthcare professional reach a clinical decision as a separate category
  (page 12); a printed sheet that calls itself an input to that decision is
  describing a purpose SymChit does not have. Notes read by a professional are
  what section 4.1 confirmed SymChit to be, so that is what the sheet says.
- The words and the address live in `internal/product`, which is the one home
  for the facts every layer may name. The domain is handed them rather than
  reading them, so no layer boundary is crossed to print a line.
- Verified by: `internal/domain/receipt_test.go::TestTheFramingWrapsTheRecord`,
  `TestTheFramingIsTheSameWhateverTheRecord`

### 3.5 Functional requirements: export

**FR-050 Export**
- Priority: Must
- Requirement: When the user asks to export, the export service shall write the
  whole record to a JSON file at a path the user chooses in a save dialog. The
  dialog shall open in the user's Downloads folder, which the user may leave
  (Amendment 4).
- Verified by: `internal/infrastructure/export/export_test.go::TestExportWritesEveryEventAndReadsBack`

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
- Verified by: `internal/application/transfer_test.go::TestExportThenImportRoundTrip`, `TestImportSkipsRepeatsWithinTheFile`

### 3.6 Functional requirements: storage and lifecycle

**FR-060 Local store**
- Priority: Must
- Requirement: The application shall keep the record in one SQLite file, in the
  folder the platform keeps a program's own configuration in: on Windows
  `%APPDATA%\SymChit\symchit.db` (Amendment 7 names the other two in 2.3).
- Verified by: `internal/infrastructure/store/unavailable_test.go::TestOpensAtTheGivenPath`, which reads
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
- Verified by: `store_test.go::TestNewerSchemaRefused` and `TestReopenKeepsEverything`. The migration list holds one entry, so no file older than the current schema exists to upgrade from; the loop that would do it runs on every open and is exercised whenever a fresh file is made

**FR-064 Single instance**
- Priority: Must
- Requirement: When SymChit starts while another instance is running for the
  same user, the new instance shall bring the running window forward and exit.
- Verified by: a manual check with two copies started; the lock is Wails' own and needs two real processes, so no test covers it (TESTING.md lists it)

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
- Verified by: `runlog_test.go`, including `TestEveryPlatformsLogFolder`, plus
  `tests/structural/logscan_test.go::TestOnlyTheKnownPlacesWriteToTheLog`,
  which holds the log's writers to a declared list. Nothing about a write to
  standard error says whether its arguments came from the record, so the
  promise is kept by there being few enough writers to read, all of them known
  and each a constant sentence plus an error or a stack. A new one fails the
  test and has to be declared, which is the moment to ask what it puts in the
  file.

**FR-066 About**
- Priority: Must
- Requirement: The application shall offer an About dialog stating the
  product name, the version, the author, the copyright notice and the
  open-source works it is built with, each with its licence, plus the statement
  in FR-067. The notice shall read `© Oliver Ernster 2026` (Amendment 8): the
  symbol says it, so the word beside it says it twice.
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
over its scrolling body. The window shall hand the page the keyboard as it
opens, so the first key pressed reaches the ring without the page being clicked
first (Amendment 11). Verified by `frontend/src/ring.test.ts`,
`useRing.test.tsx` and `window_test.go::TestReadyAsksTheWindowForTheKeyboard`,
plus a manual pass.

**NFR-USE-003 Contrast**: Text shall meet WCAG 2.2 AA contrast (4.5:1 for body
text) in the theme in use; each ring colour shall meet the 3:1 WCAG 2.2 asks
of a non-text indicator against the surfaces it is drawn on. Verified by a test
over the colour tokens.

**NFR-USE-004 Focus ring** (Amendment 8): A control shall show no ring at rest,
the ring colour while it is hovered or keyboard-focused, then the danger colour
permanently while it is disabled; a disabled control's fill shall be muted so
the ring reads against it. No container shall take focus or paint a ring. The
accent colour shall never be used as a ring. Verified by
`tests/structural/focus_test.go`, each assertion proved by planting.

**NFR-PERF-001 Startup**: The recording form shall accept input within 2 s of
process start on the reference machine. UNMEASURED: the log carries the run's
start line and no ready line, so nothing in the product times this. Measuring it
means adding a second line at the moment the page reports itself ready; until
then the number is a target rather than a reading.

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
be held at 100% statement coverage by `test.ps1`, which `build.ps1` runs before
it builds anything. `build.ps1 -Fast` skips the gate for a working loop and
prints that it has; a release is never cut with it.

**NFR-MAINT-002 Structure**: A structural test shall enforce the layering
(C-4), domain purity (no `os`, `time.Now` or `database/sql` in the domain), the
400-line module cap and its 381 to 399 danger band.

**NFR-MAINT-003 Checks**: `gofmt`, `go vet`, `staticcheck`, `tsc --noEmit`,
`eslint` and Vitest shall report nothing.

**NFR-MAINT-004 Docs**: The repository shall carry `README.md` (with who it is
for and not for), this specification, `ARCHITECTURE.md`, `TESTING.md`,
`DEVELOPMENT.md` and `VERSION`.

**NFR-MAINT-005 Wire**: A structural test shall compare the Go DTOs with the
hand-written TypeScript interfaces, field for field; a second shall compare
the receipt's line kinds with the union the page admits. The first sees that a
line carries a kind; only the second sees which kinds exist.

### 3.9 Won't this time

| Item | Reason |
|---|---|
| Other event types (medication, meal, sleep) | Source 14: version 1 stays on symptoms. The event table carries a kind column so they can follow without a migration of existing rows. |
| Phone or web version | Platform decided 2026-09-22: the desktop, which Amendment 7 widened to all three desktops rather than to a phone. |
| Synchronisation between machines | Source 11: no cloud. The export is the portability path. |
| Reminders | Source 10: no nagging. |
| Charts and trends | Source 10: no interpretation. |
| Edit history (audit trail) | Source 13 allows one; left out of version 1 (Q-7). |
| Update check | No network (NFR-PRIV-001). |

## 4. Other requirements

### 4.1 Legal and regulatory

GPL-3.0. SymChit makes no medical claim and performs no clinical function; the
design keeps it outside what UK MHRA guidance treats as software as a medical
device, which turns on intended medical purpose such as diagnosis, prevention,
prediction, prognosis, treatment or alleviation.

Owner's ruling, 2026-09-22 (Q-8): of those purposes only monitoring could be
argued at all, while SymChit is a note-taking application. It stores what the user
typed and prints it back. None of the other purposes apply.

The ruling is what the product is built to stay inside, so it is a constraint on
every later change rather than a note about this version. The line is
interpretation: the moment SymChit scores, trends, alerts, predicts or says two
events are related, the argument that it only takes notes is gone. Section 1.3
already puts every one of those out of scope; FR-042 holds the receipt to the
record alone.

Confirmed 2026-09-22 against the published text, which is what a regulator reads
rather than this document. Two sources, both read in full at the passages that
bear on SymChit:

- The Medical Devices Regulations 2002 (SI 2002/618), regulation 2, as
  published on legislation.gov.uk and stated there to be current to 21
  September 2026. "Software" is named in the definition of a medical device;
  the qualifying purposes are diagnosis, prevention, monitoring, treatment or
  alleviation of disease, the injury and handicap limbs, investigation or
  modification of the anatomy and control of conception. "Intended purpose" is
  what the manufacturer supplies "on the labelling, the instructions for use
  and/or the promotional materials", so the README, the Guide, About and the
  store listings are the text that settles it, not this document.
- MHRA, "Medical device stand-alone software including apps (including
  IVDMDs)" v1.10f, last updated 1 July 2023. Page 21 (Monitoring) lists, among
  the examples unlikely to be devices, software that simply replaces a written
  diary or log of symptoms that can be used when consulting with the patient's
  doctor. That is SymChit's intended purpose. Page 10 puts software that stores
  medical data without change on the non-medical side of the medical-purpose
  chart; it states as well that an electronic patient record simply replacing a
  paper file does not meet the definition. Appendix 1 (symptom checkers)
  governs software that matches entered symptoms to conditions, which SymChit
  does not do, so it does not apply.

The ruling therefore stands on the nearest published example there is. The
constraint on every later change is the guidance's own caveat, in its words:
"the addition of features that enhance the data presented may bring it into the
remit of the UK MDR 2002". That is broader than the list above, because it
reaches presentation and not only inference.

Two further readings taken from the same text and held as standing rules:

- A disclaimer carries no weight on its own. Page 11: a general disclaimer such
  as "this product is not a medical device" is not acceptable where medical
  claims are made elsewhere in the labelling or promotional literature. The
  README's closing line is accurate because nothing else makes a claim; it is
  not what keeps SymChit outside the definition.
- The store category is promotional material. Page 11 names the app store
  description and category among the materials that fix intended purpose, so
  SymChit is never listed under a medical or health category. The Flatpak
  desktop entry is `Utility;Office;` and the macOS bundle sets no category.

This is a documented reading of the published text, not regulatory advice. A
borderline determination can be put to the MHRA directly if one is ever wanted.

### 4.2 Internationalisation

English only in version 1. Dates print in the day-month order of the source
example (`22 Sep 17:12`).

### 4.3 Risk

Two harms matter: loss or silent corruption of the record; a receipt that
misstates it to a doctor. Rather than a full FMEA, three
requirements carry it: FR-024 (time never silently changed), FR-042 (receipt
holds nothing but the record) and NFR-REL-001 (durability). Judged
proportionate for a single-user record keeper.

### 4.4 What version 1 commits to

A first release is where a promise starts, so the promises are named rather
than left to be inferred from the code as it happens to stand.

- **The export format is a contract, not a snapshot.** Every file SymChit has
  ever written stays readable by every later SymChit. The envelope (`format`
  and an integer `version`) never changes shape; a new version is added beside
  the old readers rather than in place of them, with the old version's real
  bytes frozen under `testdata` and read by the suite on every run. This is
  what makes an export a file the user owns: it is worth nothing if the program
  that wrote it is the only one that will ever read it, then worth nothing again
  if next year's SymChit refuses last year's file.
- **The printed sheet's wording is a public claim** (FR-045). It goes into
  filing cabinets and is read by people who will never see the application, so
  the line saying the record is not a diagnosis is held by a test asserting it
  whole; changing it is a decision about what SymChit tells a clinician rather
  than a wording tidy-up.
- **The record's own schema** is versioned by the store and a newer schema is
  refused rather than guessed at, so a record written by a later SymChit is
  never silently misread by this one.

What version 1 does NOT commit to: the window's layout, the wording of anything on
screen other than the sheet, the log's shape, the setup program's screens or
any internal structure. Those change whenever they are improved.

## 5. Appendices

### Appendix A: Feasibility measurements

| ID | Question | Measured |
|---|---|---|
| M-1 | Does `window.print()` in a Wails v2.12 window open the Windows print dialog with Microsoft Print to PDF offered? | Yes. Measured in the built window on 2026-09-22: the dialog opened, offering Save as PDF and Microsoft Print to PDF; the printed page carried the receipt alone. |

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
| Q-8 | Does the MHRA reading in 4.1 hold? | Yes. Ruled 2026-09-22, then confirmed the same day against SI 2002/618 reg. 2 and MHRA guidance v1.10f, page 21, which lists a replacement for a written symptom diary among the examples unlikely to be devices. | 4.1 |
| Q-10 | Identifying details on the receipt | None. | FR-041 |

Still open:

| ID | Question | Proposal | Owner | Gates |
|---|---|---|---|---|
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
| 11 | 2026-09-22 | NFR-USE-002 | The window hands the page the keyboard as it opens, through a focuser seam wired at the composition root and called on DOM ready. | A defect the owner found in the built window: no Tab ever stepped the ring, while a single click on the page fixed it for the rest of the run. WebView2 keeps DOM focus and keyboard focus apart, so the sink held the first while the webview held none of the second and no keydown reached the listener at all. Read from the Wails 2.12.0 source: its Windows frontend hands the webview focus only from WM_SETFOCUS on the main window, which nothing raises at startup; `runtime.Show` raises it. The page cannot fix this from its own side, since a DOM focus call cannot make the webview the thing keys are sent to. |
| 10 | 2026-09-22 | FR-041, FR-042, new FR-045 | The printed record is framed: the program and its address above the title and below the last event, then a statement under the date range saying it is not a diagnosis but the person's own notes for a healthcare professional. | Owner's request. It also answers the surface Amendment 9 identified as the one a regulator reads: a sheet that leaves the recording program's hand carries what it is and what it is not, rather than relying on a reader who has seen the application. The sheet says notes rather than guidance, since guidance names a purpose SymChit does not have (FR-045). |
| 9 | 2026-09-22 | 4.1, Q-8 | The MHRA reading is confirmed against SI 2002/618 reg. 2 and MHRA guidance v1.10f, so Q-8 closes. The guidance's own caveat about features that enhance the data presented replaces the narrower list as the standing constraint. Two further rules are recorded: a disclaimer carries no weight on its own; the store category is promotional material. | The ruling had been made against a summary of the guidance rather than its text, while intended purpose is fixed by what the manufacturer publishes, so the text was the only thing that could settle it. Measured: page 21 lists a replacement for a written symptom diary among the examples unlikely to be devices, which is SymChit's intended purpose. |
| 8 | 2026-09-22 | NFR-USE-002, new NFR-USE-004, FR-066 | The house keyboard model and its three-state focus ring are stated as requirements rather than left to the page; About names the copyright holder and year. | Owner's request. The ring was a single blue outline on keyboard focus alone, which said nothing about what could be used and nothing about what could not: Print sat inert beside the button that fills it with no way to tell it apart from a control waiting to be pressed. |
| 7 | 2026-09-22 | 1.3 scope, 2.3 operating environment, FR-060, FR-065 | Linux and macOS leave the deferred list and become part of this version: a Flatpak and a signed DMG, ported from PigeonPost's. Only the bespoke setup program stays Windows-only. | Owner's decision, superseding Amendment 3, which had made them a later version. Measured: one package stopped the module building elsewhere, the run log, whose Windows handle work is now behind a build tag and whose folder rule is a pure function taking the platform as an argument, so all three answers are exercised on every platform. The gate now builds and vets for Linux and macOS on every run, which is what keeps this true; a planted Windows-only import was refused by name. |
| 6 | 2026-09-22 | New FR-069 | The bar carries a Donate button, last in its right-hand group. | Owner's request. It takes a seat in the bar the window already has rather than a band of its own, as AudioDeck's does, since a whole new strip of chrome carrying one control costs more than it buys. The address lives once, in Go's product package; the page asks for the donation page rather than naming one, so nothing arriving from the page has to be checked before it is opened. |
| 5 | 2026-09-22 | New FR-068 | The application carries a Guide, reached from the bar. | Owner's request, in line with PigeonPost and ClearBudget. Its words are one document (`frontend/src/guideContent.ts`) and the dialog only draws them, as PigeonPost's does. |
| 4 | 2026-09-22 | FR-050, FR-053 | Both file dialogs open in the user's Downloads folder, which the user may leave. | Owner's request: it is where a person already looks for files they have saved. Measured: Wails' dialog options carry a default directory; the folder is checked before it is named, so a machine without one falls back to the dialog's own choice. |
| 3 | 2026-09-22 | 1.3 scope | Linux and macOS leave the permanent out-of-scope list and become planned for a later version: a Flatpak and a DMG builder, following PigeonPost's. | Owner's decision. Version 1 stays Windows only; nothing in the domain, application or page is Windows-specific, so the work is packaging plus the run log's standard-handle code. |
| 2 | 2026-09-22 | FR-041 | Receipt events name their year when the range crosses a year; they leave it out otherwise. | Found while writing the tests: the source example's `22 Sep 17:12` is unambiguous inside one year and ambiguous across two, which a doctor reading the record cannot resolve. |
| 1 | 2026-09-22 | FR-011 | The acceptance example's second case is corrected: typing `a` lists `Headache` and `Back pain`, not all three. | The baselined example was wrong: `Tired` holds no letter `a`. Found by the test that was written from it. |
