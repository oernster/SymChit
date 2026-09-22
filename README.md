# <img width="128" height="128" alt="application-icon" src="https://github.com/user-attachments/assets/6558296b-5218-4165-8a3a-0ca71ea7f0e7" /> SymChit

A health tracker.  It doesn't diagnose.  It's just a little record of symptoms.

> **Commercial licences available.** SymChit is free and open source under the
> GPL-3.0. If those terms do not suit what you are building, such as a
> closed-source product, a commercial licence can be bought from me separately.
> It covers my own code; third-party libraries keep their own licences. See
> [commercial licensing](https://ernster.dev/commercial-licensing.html).

When you notice a symptom, record it. When you see your doctor, take the record
with you.

SymChit is a local-first symptom recorder for Windows, Linux and macOS. It keeps what you
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
  'none'`. The Donate button is not an exception: it hands an address to
  Windows and your browser does the asking, so SymChit still fetches nothing.
- It does not encrypt your record. The file is protected by your Windows
  account, as your documents are.

## What it does

- Records a symptom in a few seconds, with an optional note and severity.
- Timestamps it for you; you can also enter something you remembered later.
- Keeps each symptom you name, exactly as you typed it, then offers it next time.
- Shows the history newest first, filtered by dates, symptom or severity.
- Corrects and deletes entries, asking first and naming what will go.
- Prints a symptom record for a date range, which is also a PDF if you choose
  Save as PDF. The sheet names the program that produced it and says, above the
  first entry, that it is not a diagnosis: it is your own notes, printed for a
  healthcare professional to read.
- Exports the whole record to a JSON file you own; reads one back too.
- Opens dark, with a button in the bar that moves it to light and remembers
  which you chose.

## Built with

| Piece | What |
|---|---|
| Language | Go 1.26, cgo disabled |
| Window | Wails v2 with WebView2 |
| Page | React 18 and TypeScript, built by Vite |
| Record | SQLite through `modernc.org/sqlite` (pure Go), WAL, synchronous FULL |
| Layering | `internal/{domain,application,infrastructure,ui}`, enforced by tests |

## Running it

Download the executable on Windows, the Flatpak bundle on Linux or the DMG on
macOS. To build any of them from source, see [DEVELOPMENT.md](DEVELOPMENT.md).

Your record and the run log go wherever the platform keeps such things:

| | The record | The run log |
|---|---|---|
| Windows | `%APPDATA%\SymChit\symchit.db` | `%LOCALAPPDATA%\SymChit\SymChit.log` |
| Linux | `$XDG_CONFIG_HOME/SymChit/symchit.db` | `$XDG_STATE_HOME/SymChit/SymChit.log` |
| macOS | `~/Library/Application Support/SymChit/symchit.db` | `~/Library/Logs/SymChit/SymChit.log` |

Inside the Flatpak both land under `~/.var/app/uk.codecrafter.SymChit`.

All three have been built and run. Windows is the platform SymChit has been used
on; the Flatpak and the DMG have each been built on their own machine and the
window opened, with the DMG notarized. The bespoke setup program stays
Windows-only: a Flatpak and a DMG are how those platforms install things.

## Testing

```powershell
./test.ps1
```

What it runs and what each floor means is in [TESTING.md](TESTING.md).

## Building

```powershell
./build.ps1
```

The gate runs first. `-Fast` skips it for a working loop and says so; a release
is never cut that way.

## Documentation

- [REQUIREMENTS.md](REQUIREMENTS.md): what SymChit must do, requirement by
  requirement, with the acceptance criteria and the tests that verify them.
- [ARCHITECTURE.md](ARCHITECTURE.md): the invariants and the tests that enforce
  them.
- [TESTING.md](TESTING.md): the gate, the floors and what only a person can
  check.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks like debt.

## Supporting the project

SymChit is free and stays free. There is no paid tier, no licence key and no
feature held back behind a donation. If it has saved you an afternoon or made an
appointment go better, the Donate button at the right of the bar opens a
contribution page in your browser. SymChit sends nothing itself: it hands the
address to Windows and your browser does the rest.

<a href="https://www.paypal.com/ncp/payment/4XP3AYNMPQGUC"><img src="frontend/src/assets/icons/donate.png" alt="Donate to SymChit" width="120"></a>

## What version 1 promises

A first release is where a promise starts, so here is what this one is.

Your exported file stays readable. Every export SymChit has ever written will
be read by every later SymChit: the format carries its own version and old
readers are kept beside new ones, with each old version's real bytes frozen in
the test suite. A file you own is worth nothing if next year's version refuses
it.

The printed sheet's wording is a public claim, not decoration. The line saying
the record is not a diagnosis is held by a test that asserts it whole.

What is not promised: the window's layout, the wording on screen, the log or
the setup program's screens. Those improve whenever they can.

## Licence

GPL-3.0. See [LICENSE](LICENSE). For terms that suit a closed-source product,
see [commercial licensing](https://ernster.dev/commercial-licensing.html).

SymChit is not a medical device and gives no medical advice.
