import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { SymptomsPane } from './SymptomsPane'
import { installBridge, tired } from './bridge-fake'

describe('the symptoms screen', () => {
  it('says so when nothing has been recorded yet', async () => {
    installBridge()
    render(<SymptomsPane refused={vi.fn()} tell={vi.fn()} />)
    expect(await screen.findByText(/No symptoms yet/)).toBeInTheDocument()
  })

  it('renames a symptom', async () => {
    const tell = vi.fn()
    const bridge = installBridge({
      Symptoms: vi.fn(() => Promise.resolve([tired])),
      Rename: vi.fn(() => Promise.resolve()),
    })
    render(<SymptomsPane refused={vi.fn()} tell={tell} />)

    fireEvent.click(await screen.findByRole('button', { name: 'Rename' }))
    const field = screen.getByLabelText('New name for Tired')
    expect(field).toHaveValue('Tired')
    fireEvent.change(field, { target: { value: 'Exhausted' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => expect(bridge.Rename).toHaveBeenCalledWith(tired.id, 'Exhausted'))
    expect(tell).toHaveBeenCalledWith('Renamed Tired to Exhausted.')
  })

  it('keeps the name when the rename was refused; the same on cancel', async () => {
    const tell = vi.fn()
    const refused = vi.fn()
    installBridge({
      Symptoms: vi.fn(() => Promise.resolve([tired])),
      Rename: vi.fn(() => Promise.reject(new Error('another symptom already has that name'))),
    })
    render(<SymptomsPane refused={refused} tell={tell} />)

    fireEvent.click(await screen.findByRole('button', { name: 'Rename' }))
    fireEvent.change(screen.getByLabelText('New name for Tired'), { target: { value: 'Headache' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    await waitFor(() => expect(refused).toHaveBeenCalled())
    expect(tell).not.toHaveBeenCalled()

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.getByText('Tired')).toBeInTheDocument()
  })
})
