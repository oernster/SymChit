// The receipt: the factual record for a date range, printed through the
// Windows print dialog, which also offers saving as PDF (FR-040 to FR-044).
// Every word on it comes from the backend's receipt; the page only lays it out.

import { useEffect, useState } from 'react'
import { api, type ReceiptLine, type Refused } from './api'
import crest from './assets/icons/application-icon.png'

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

/**
 * Whether a line is the sheet's letterhead: the first line, a provenance one.
 * The position is part of the test and not only the kind, so a mark can never
 * arrive anywhere other than the head of the sheet.
 */
function isLetterhead(line: ReceiptLine, index: number): boolean {
  return index === 0 && line.kind === 'provenance'
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
          {/*
            The sheet is a table so that it can be printed with no page margin
            at all, which is the only way to leave the browser nowhere to draw
            its own header and footer (FR-045, Amendment 14). A print engine
            repeats a thead and a tfoot on every page it lays out, so those two
            empty rows reserve the band at the top and the foot of every sheet.
            Padding cannot do it: padding applies once to the element, not once
            per page, so it leaves page two starting at the edge of the paper.
            It holds no data and announces nothing, hence the presentation role.
          */}
          <table className="sheet" role="presentation">
            <thead>
              <tr>
                <td className="gutter" />
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  {lines.map((line, index) => (
                    <p key={index} className={`line ${line.kind}`}>
                      {/*
                        The mark sits with the words rather than above them,
                        because the pair is the letterhead: a picture and the
                        address it belongs to. It is decorative, so it carries
                        no alt text; the line beside it already names the
                        program; a screen reader saying it twice helps
                        nobody.
                      */}
                      {isLetterhead(line, index) && (
                        <img className="mark" src={crest} alt="" />
                      )}
                      {line.text}
                    </p>
                  ))}
                </td>
              </tr>
            </tbody>
            <tfoot>
              <tr>
                <td className="gutter" />
              </tr>
            </tfoot>
          </table>
        </article>
      )}
    </section>
  )
}
