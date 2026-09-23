import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ReceiptPane, daysBefore } from './ReceiptPane'
import { aReceipt, installBridge } from './bridge-fake'

describe('the receipt', () => {
  it('opens on the last thirty days, ending today', async () => {
    const bridge = installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} saved={vi.fn()} />)

    await waitFor(() => expect(screen.getByLabelText(/To/)).toHaveValue('2026-09-22'))
    expect(screen.getByLabelText(/From/)).toHaveValue('2026-08-23')

    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await waitFor(() => expect(bridge.Receipt).toHaveBeenCalledWith('2026-08-23', '2026-09-22'))
  })

  it('shows every line the backend gave it; nothing else', async () => {
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} saved={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    const shown = Array.from(record.querySelectorAll('p')).map((line) => line.textContent)
    expect(shown).toEqual(aReceipt.map((line) => line.text))
  })

  it('sets the framing apart from the record it frames', async () => {
    // The framing (FR-045) is styled by its kind, so a line that loses its kind
    // loses the separation on paper while still reading correctly here.
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} saved={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))

    const record = await screen.findByLabelText('The symptom record')
    const lines = Array.from(record.querySelectorAll('p'))
    const framed = lines.filter((line) => line.classList.contains('provenance'))
    expect(framed).toHaveLength(1)
    expect(framed[0]).toBe(lines[0])
    expect(record.querySelectorAll('p.statement')).toHaveLength(1)
  })

  it('marks the sheet once, beside the address at the top', async () => {
    // The letterhead: the application's own icon beside the line naming the
    // program and where it lives, so a sheet on a desk of paper says what
    // produced it. It belongs to the opening line alone, so its place is
    // asserted rather than only its presence.
    installBridge({ Receipt: vi.fn(() => Promise.resolve(aReceipt)) })
    render(<ReceiptPane refused={vi.fn()} saved={vi.fn()} />)
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

  it('cannot be saved until there is something to save', async () => {
    const bridge = installBridge({
      Receipt: vi.fn(() => Promise.resolve(aReceipt)),
      SavePDF: vi.fn(() => Promise.resolve('C:/Users/x/Downloads/record.pdf')),
    })
    const saved = vi.fn()
    render(<ReceiptPane refused={vi.fn()} saved={saved} />)
    expect(screen.getByRole('button', { name: 'Save PDF' })).toBeDisabled()

    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Save PDF' })).toBeEnabled())

    // Go draws the document, not the browser: three engines print a page three
    // different ways; the record is what the product is for.
    fireEvent.click(screen.getByRole('button', { name: 'Save PDF' }))
    await waitFor(() =>
      expect(bridge.SavePDF).toHaveBeenCalledWith('2026-08-23', '2026-09-22'))
    await waitFor(() =>
      expect(saved).toHaveBeenCalledWith(
        'Your symptom record was saved to C:/Users/x/Downloads/record.pdf.'))
  })

  it('says nothing when the reader cancels the save dialog', async () => {
    // An empty path is a cancelled dialog. Announcing it would tell the reader
    // something happened when they had just decided it should not.
    installBridge({
      Receipt: vi.fn(() => Promise.resolve(aReceipt)),
      SavePDF: vi.fn(() => Promise.resolve('')),
    })
    const saved = vi.fn()
    render(<ReceiptPane refused={vi.fn()} saved={saved} />)
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Save PDF' })).toBeEnabled())

    fireEvent.click(screen.getByRole('button', { name: 'Save PDF' }))
    await waitFor(() => expect(saved).not.toHaveBeenCalled())
  })

  it('shows nothing when the range holds no events', async () => {
    const refused = vi.fn()
    installBridge({
      Receipt: vi.fn(() => Promise.reject(new Error('no events were recorded in that range'))),
    })
    render(<ReceiptPane refused={refused} saved={vi.fn()} />)
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
