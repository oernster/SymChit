import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ReceiptPane, daysBefore } from './ReceiptPane'
import { aReceipt, installBridge } from './bridge-fake'

describe('the receipt', () => {
  it('opens on the last thirty days, ending today', async () => {
    const bridge = installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} />)

    await waitFor(() => expect(screen.getByLabelText(/To/)).toHaveValue('2026-09-22'))
    expect(screen.getByLabelText(/From/)).toHaveValue('2026-08-23')

    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await waitFor(() => expect(bridge.Receipt).toHaveBeenCalledWith('2026-08-23', '2026-09-22'))
  })

  it('shows every line the backend gave it; nothing else', async () => {
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    const shown = Array.from(record.querySelectorAll('p')).map((line) => line.textContent)
    expect(shown).toEqual(aReceipt.map((line) => line.text))
  })

  it('sets the framing apart from the record it frames', async () => {
    // The framing (FR-045) is styled by its kind, so a line that loses its kind
    // loses the separation on paper while still reading correctly here.
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    const lines = Array.from(record.querySelectorAll('p'))
    const framed = lines.filter((line) => line.classList.contains('provenance'))
    expect(framed).toHaveLength(1)
    expect(framed[0]).toBe(lines[0])
    expect(record.querySelectorAll('p.statement')).toHaveLength(1)
  })

  it('lays the record out with a head and a foot that a printer repeats', async () => {
    // The sheet prints with no page margin, so that the browser has nowhere to
    // draw its own header and footer. What then holds the record off the top and
    // bottom edges of every page is these two empty rows: a print engine lays a
    // thead and a tfoot out again on each page, while padding is applied once to
    // the element and leaves page two starting at the paper's edge.
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    expect(record.querySelectorAll('thead td.gutter')).toHaveLength(1)
    expect(record.querySelectorAll('tfoot td.gutter')).toHaveLength(1)
    // Every line of the record sits in the body, so none of it can be repeated
    // onto a page it does not belong to.
    expect(record.querySelectorAll('tbody p.line').length).toBe(aReceipt.length)
  })

  it('marks the sheet once, beside the address at the top', async () => {
    // The letterhead: the application's own icon beside the line naming the
    // program and where it lives, so a sheet on a desk of paper says what
    // produced it. It belongs to the opening line alone, so its place is
    // asserted rather than only its presence.
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    const marks = record.querySelectorAll('img.mark')
    expect(marks).toHaveLength(1)

    const lines = Array.from(record.querySelectorAll('p'))
    expect(lines[0]).toContainElement(marks[0] as HTMLElement)
    for (const line of lines.slice(1)) {
      expect(line.querySelector('img')).toBeNull()
    }
    // Decorative: the words beside it already name the program, so a reader
    // hearing the page read out should not be told twice.
    expect(marks[0]).toHaveAttribute('alt', '')
  })

  it('cannot be printed until there is something to print', async () => {
    const bridge = installBridge({
      Receipt: vi.fn(() => Promise.resolve(aReceipt)),
      Print: vi.fn(() => Promise.resolve()),
    })
    render(<ReceiptPane refused={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Print' })).toBeDisabled()

    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Print' })).toBeEnabled())

    // The window prints, not the page: WebKit on macOS does nothing with
    // window.print(), so a page that called it had a button that never answered.
    const print = vi.fn()
    window.print = print
    fireEvent.click(screen.getByRole('button', { name: 'Print' }))
    await waitFor(() => expect(bridge.Print).toHaveBeenCalledTimes(1))
    expect(print).not.toHaveBeenCalled()
  })

  it('shows nothing when the range holds no events', async () => {
    const refused = vi.fn()
    installBridge({
      Receipt: vi.fn(() => Promise.reject(new Error('no events were recorded in that range'))),
    })
    render(<ReceiptPane refused={refused} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    await waitFor(() => expect(refused).toHaveBeenCalledWith('No events were recorded in that range'))
    expect(screen.queryByLabelText('The symptom record')).toBeNull()
  })

  it('counts back across a month and a year', () => {
    expect(daysBefore('2026-09-22', 30)).toBe('2026-08-23')
    expect(daysBefore('2026-01-05', 30)).toBe('2025-12-06')
    expect(daysBefore('2026-03-01', 1)).toBe('2026-02-28')
  })
})
