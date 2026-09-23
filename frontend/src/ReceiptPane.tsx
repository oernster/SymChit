// The receipt: the factual record for a date range, shown here and saved as a
// PDF the reader keeps (FR-040 to FR-044). Every word on it comes from the
// backend's receipt; the page only lays it out on screen, while the document
// itself is drawn by Go so that it reads the same on every desktop.

import { useEffect, useState } from 'react'
import { api, type ReceiptLine, type Refused } from './api'
import crest from './assets/icons/application-icon.png'

interface Props {
  refused: Refused
  /** saved says where the document went, once it has gone there. */
  saved: (message: string) => void
}

/** defaultDays is how far back the range starts when the pane opens. */
const defaultDays = 30

/** daysBefore answers the ISO date a number of days before an ISO date. */
export function daysBefore(iso: string, days: number): string {
  const [year, month, day] = iso.split('-').map(Number)
  const date = new Date(Date.UTC(year, month - 1, day - days))
  return date.toISOString().slice(0, 10)
}

/**
 * Whether a line is the sheet's letterhead: the first line, a provenance one.
 * The position is part of the test and not only the kind, so a mark can never
 * arrive anywhere other than the head of the sheet.
 */
function isLetterhead(line: ReceiptLine, index: number): boolean {
  return index === 0 && line.kind === 'provenance'
}

export function ReceiptPane({ refused, saved }: Props) {
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [lines, setLines] = useState<ReceiptLine[]>([])

  useEffect(() => {
    void api.now(refused).then((now) => {
      if (!now) return
      const today = now.slice(0, 10)
      setTo(today)
      setFrom(daysBefore(today, defaultDays))
    })
  }, [refused])

  const show = async () => {
    setLines([])
    const found = await api.receipt(from, to, refused)
    if (found) setLines(found)
  }

  const savePDF = async () => {
    const path = await api.savePDF(from, to, refused)
    // An empty path is a cancelled dialog, which is not worth announcing.
    if (path) saved(`Your symptom record was saved to ${path}.`)
  }

  return (
    <section className="pane receipt-pane" aria-labelledby="receipt-title">
      <h1 id="receipt-title">Symptom record</h1>
      <div className="receipt-controls">
        <label>
          From <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        </label>
        <label>
          To <input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        </label>
        <button type="button" onClick={show}>Show the record</button>
        <button type="button" className="primary" disabled={lines.length === 0}
          onClick={() => void savePDF()}>
          Save PDF
        </button>
      </div>
      {lines.length > 0 && (
        <article className="receipt" aria-label="The symptom record">
          {lines.map((line, index) => (
            <p key={index} className={`line ${line.kind}`}>
              {/*
                The mark sits with the words rather than above them, because
                the pair is the letterhead: a picture and the address it
                belongs to. It is decorative, so it carries no alt text; the
                line beside it already names the program.
              */}
              {isLetterhead(line, index) && <img className="mark" src={crest} alt="" />}
              {line.text}
            </p>
          ))}
        </article>
      )}
    </section>
  )
}
