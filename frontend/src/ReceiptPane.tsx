// The receipt: the factual record for a date range, printed through the
// Windows print dialog, which also offers saving as PDF (FR-040 to FR-044).
// Every word on it comes from the backend's receipt; the page only lays it out.

import { useEffect, useState } from 'react'
import { api, type ReceiptLine, type Refused } from './api'

interface Props {
  refused: Refused
}

/** defaultDays is how far back the range starts when the pane opens. */
const defaultDays = 30

/** daysBefore answers the ISO date a number of days before an ISO date. */
export function daysBefore(iso: string, days: number): string {
  const [year, month, day] = iso.split('-').map(Number)
  const date = new Date(Date.UTC(year, month - 1, day - days))
  return date.toISOString().slice(0, 10)
}

export function ReceiptPane({ refused }: Props) {
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
          onClick={() => window.print()}>
          Print
        </button>
      </div>
      {lines.length > 0 && (
        <article className="receipt" aria-label="The symptom record">
          {lines.map((line, index) => (
            <p key={index} className={`line ${line.kind}`}>{line.text}</p>
          ))}
        </article>
      )}
    </section>
  )
}
