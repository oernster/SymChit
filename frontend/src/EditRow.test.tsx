import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { EditRow } from './EditRow'
import { anEvent, installBridge, severities } from './bridge-fake'

/** editing renders the edit row over a facade that answers an edit. */
function editing() {
  const edit = vi.fn(() => Promise.resolve(anEvent))
  const saved = vi.fn()
  installBridge({ Edit: edit })
  render(
    <EditRow
      event={anEvent}
      severities={severities}
      refused={vi.fn()}
      onSaved={saved}
      onCancel={vi.fn()}
    />,
  )
  return { edit, saved }
}

describe('editing an event', () => {
  it('sends no time when the time was not touched', async () => {
    const { edit, saved } = editing()
    fireEvent.change(screen.getByLabelText('Note'), { target: { value: 'Went away.' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(edit).toHaveBeenCalled())
    expect(edit).toHaveBeenCalledWith(anEvent.id, {
      symptom: anEvent.symptom,
      severity: anEvent.severity,
      note: 'Went away.',
      occurredAt: '',
    })
    expect(saved).toHaveBeenCalledWith(anEvent)
  })

  it('sends the time when it was changed', async () => {
    const { edit } = editing()
    fireEvent.change(screen.getByLabelText('Time'), { target: { value: '2026-09-22T12:30' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(edit).toHaveBeenCalled())
    expect(edit).toHaveBeenCalledWith(
      anEvent.id,
      expect.objectContaining({ occurredAt: '2026-09-22T12:30' }),
    )
  })

  it('opens with what the event holds', () => {
    editing()
    expect(screen.getByLabelText('Symptom')).toHaveValue(anEvent.symptom)
    expect(screen.getByLabelText('Time')).toHaveValue(anEvent.occurredAt)
    expect(screen.getByLabelText(anEvent.severity)).toBeChecked()
  })
})
