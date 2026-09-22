# SymChit
A health tracker.  It doesn't contain diagnose.  It's just a little record of symptoms.

When you notice a symptom, record it. When you see your doctor, take the record
with you.

SymChit is a local-first symptom recorder for Windows. It keeps what you
observed and when you observed it, then prints a short factual record for an
appointment. It is a recorder, not a diagnostician.

Instead of telling a doctor "I have been tired a lot recently", you can show
them:

```text
SYMPTOM RECORD
23 August - 22 September 2026

Tired - 9 recorded events

28 Aug 07:45
  Felt unusually tired shortly after getting up.

22 Sep 17:12
  Only been awake for about 10 minutes.
```

## Who it is for

For anyone who wants an accurate record of their own symptoms to discuss with a healthcare professional; someone who would rather keep that record on their own computer.

Not for anyone wanting advice about what their symptoms mean, a wellness score,
a medication tracker or anything to sync between devices. SymChit does none of
those and is not intended to.

## What it does not do

Stated plainly, because a health application is usually assumed to do some of
these:

- It does not diagnose, suggest conditions or recommend treatment.
- It does not interpret your observations, score them or claim that one event
  led to another. The receipt counts your entries; that is the only arithmetic.
- It does not rewrite what you wrote into medical terminology.
- It does not remind or nag you to record anything.
- It opens no network connection at all. There is no account, no cloud service,
  no advertising and no telemetry. A test forbids the networking packages
  outright and the page carries a Content-Security-Policy of `connect-src
  'none'`.
- It does not encrypt your record. The file is protected by your Windows
  account, as your documents are.

## What it does

- Records a symptom in a few seconds, with an optional note and severity.
- Timestamps it for you; you can also enter something you remembered later.
- Keeps each symptom you name, exactly as you typed it, then offers it next time.
- Shows the history newest first, filtered by dates, symptom or severity.
- Corrects and deletes entries, asking first and naming what will go.
- Prints a symptom record for a date range, which is also a PDF if you choose
  Save as PDF.
- Exports the whole record to a JSON file you own; reads one back too.

## Built with

| Piece | What |
|---|---|
| Language | Go 1.26, cgo disabled |
| Window | Wails v2 with WebView2 |
| Page | React 18 and TypeScript, built by Vite |
| Record | SQLite through `modernc.org/sqlite` (pure Go), WAL, synchronous FULL |
| Layering | `internal/{domain,application,infrastructure,ui}`, enforced by tests |

## Running it

Download the executable. To build from source, see
[DEVELOPMENT.md](DEVELOPMENT.md).

Your record lives at `%APPDATA%\SymChit\symchit.db` and the run log at
`%LOCALAPPDATA%\SymChit\SymChit.log`.

## Testing

```powershell
./test.ps1
```

What it runs and what each floor means is in [TESTING.md](TESTING.md).

## Building

```powershell
./build.ps1
```

The gate runs first and cannot be skipped.

## Documentation

- [REQUIREMENTS.md](REQUIREMENTS.md): what SymChit must do, requirement by
  requirement, with the acceptance criteria and the tests that verify them.
- [ARCHITECTURE.md](ARCHITECTURE.md): the invariants and the tests that enforce
  them.
- [TESTING.md](TESTING.md): the gate, the floors and what only a person can
  check.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.

## Licence

GPL-3.0. See [LICENSE](LICENSE).

SymChit is not a medical device and gives no medical advice.
