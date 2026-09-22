import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { SymptomInput } from './SymptomInput'
import { installBridge } from './bridge-fake'

/** Host holds the value, as the forms do. */
function Host({ refused }: { refused: () => void }) {
  const [value, setValue] = useState('')
  return <SymptomInput id="s" value={value} onChange={setValue} refused={refused} />
}

const held = [
  { id: 1, label: 'Tired' },
  { id: 2, label: 'Headache' },
]

describe('the symptom field', () => {
  it('suggests held symptoms and takes one with the keyboard', async () => {
    installBridge({ Suggest: vi.fn(() => Promise.resolve(held)) })
    render(<Host refused={vi.fn()} />)
    const field = screen.getByRole('combobox')

    fireEvent.change(field, { target: { value: 'ti' } })
    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(2))
    expect(field).toHaveAttribute('aria-expanded', 'true')

    fireEvent.keyDown(field, { key: 'ArrowDown' })
    fireEvent.keyDown(field, { key: 'ArrowDown' })
    expect(screen.getAllByRole('option')[1]).toHaveAttribute('aria-selected', 'true')
    fireEvent.keyDown(field, { key: 'ArrowUp' })
    fireEvent.keyDown(field, { key: 'Enter' })

    await waitFor(() => expect(field).toHaveValue('Tired'))
    expect(screen.queryByRole('option')).toBeNull()
  })

  it('takes one with the mouse', async () => {
    installBridge({ Suggest: vi.fn(() => Promise.resolve(held)) })
    render(<Host refused={vi.fn()} />)
    const field = screen.getByRole('combobox')
    fireEvent.change(field, { target: { value: 'h' } })
    fireEvent.mouseDown(await screen.findByText('Headache'))
    await waitFor(() => expect(field).toHaveValue('Headache'))
  })

  it('closes the list on Escape and on leaving the field', async () => {
    installBridge({ Suggest: vi.fn(() => Promise.resolve(held)) })
    render(<Host refused={vi.fn()} />)
    const field = screen.getByRole('combobox')

    fireEvent.change(field, { target: { value: 't' } })
    await screen.findAllByRole('option')
    fireEvent.keyDown(field, { key: 'Escape' })
    expect(screen.queryByRole('option')).toBeNull()

    fireEvent.keyDown(field, { key: 'ArrowDown' })
    await screen.findAllByRole('option')
    fireEvent.blur(field)
    expect(screen.queryByRole('option')).toBeNull()
  })

  it('leaves Enter to the form when nothing is highlighted', async () => {
    installBridge({ Suggest: vi.fn(() => Promise.resolve(held)) })
    const submit = vi.fn((event: { preventDefault: () => void }) => event.preventDefault())
    render(
      <form onSubmit={submit}>
        <Host refused={vi.fn()} />
      </form>,
    )
    const field = screen.getByRole('combobox')
    fireEvent.change(field, { target: { value: 'Tired' } })
    await screen.findAllByRole('option')
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(field).toHaveValue('Tired')
  })

  it('says nothing and shows nothing when the suggestions are refused', async () => {
    const refused = vi.fn()
    installBridge({ Suggest: vi.fn(() => Promise.reject(new Error('the record could not be read'))) })
    render(<Host refused={refused} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'ti' } })
    await waitFor(() => expect(refused).toHaveBeenCalled())
    expect(screen.queryByRole('option')).toBeNull()
  })
})
