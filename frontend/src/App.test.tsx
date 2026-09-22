import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { App } from './App'
import { aState, anImport, installBridge, noWindowShown } from './bridge-fake'

describe('the shell', () => {
  it('opens on the recording form and moves between the screens', async () => {
    installBridge()
    render(<App />)
    expect(await screen.findByRole('heading', { name: 'Record a symptom' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /History/ }))
    expect(await screen.findByRole('heading', { name: 'History' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Receipt/ }))
    expect(await screen.findByRole('heading', { name: 'Symptom record' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Symptoms/ }))
    expect(await screen.findByRole('heading', { name: 'Symptoms' })).toBeInTheDocument()
  })

  it('shows the problem when the record could not be opened', async () => {
    installBridge({
      State: vi.fn(() => Promise.resolve({ ...aState, problem: 'Your record could not be opened.' })),
    })
    render(<App />)
    expect(await screen.findByRole('alert')).toHaveTextContent('Your record could not be opened.')
  })

  it('says where the record went and what an import did', async () => {
    installBridge({
      Export: vi.fn(() => Promise.resolve('C:\\Users\\Oliver\\Downloads\\record.json')),
      Import: vi.fn(() => Promise.resolve(anImport)),
    })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Export/ }))
    expect(await screen.findByText(/Your record was exported to/)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Import/ }))
    expect(
      await screen.findByText('Imported 3 events. Skipped 1 event already in your record.'),
    ).toBeInTheDocument()
  })

  it('says nothing when a file dialog is cancelled', async () => {
    installBridge({
      Export: vi.fn(() => Promise.resolve('')),
      Import: vi.fn(() => Promise.resolve({ chosen: false, added: 0, skipped: 0 })),
    })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Export/ }))
    fireEvent.click(screen.getByRole('button', { name: /Import/ }))
    await waitFor(() => expect(screen.queryByText(/Imported/)).toBeNull())
    expect(screen.queryByText(/exported to/)).toBeNull()
  })

  it('opens the Guide and About, then closes them again', async () => {
    installBridge()
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Guide/ }))
    expect(await screen.findByRole('heading', { name: 'How SymChit works' })).toBeInTheDocument()
    expect(screen.getByText(/It does not diagnose\./)).toBeInTheDocument()
    fireEvent.keyDown(window, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())

    fireEvent.click(screen.getByRole('button', { name: /About/ }))
    expect(await screen.findByRole('heading', { name: 'SymChit 1.0.0' })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('offers the donation page last in the band, saying that the browser opens', async () => {
    const bridge = installBridge({ Donate: vi.fn(() => Promise.resolve()) })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    const donate = screen.getByRole('button', { name: /Donate/ })
    // The picture says nothing on its own about leaving the application.
    expect(donate).toHaveAttribute('title', 'Buy the author a drink (opens your browser)')
    // Last in the band: it belongs to nothing on screen, so nothing else is
    // reached by accident on the way to it.
    const band = screen.getByRole('navigation')
    const buttons = Array.from(band.querySelectorAll('button'))
    expect(buttons[buttons.length - 1]).toBe(donate)

    fireEvent.click(donate)
    // The page asks for the donation page and names no address: the one home
    // for that address is Go's product package.
    await waitFor(() => expect(bridge.Donate).toHaveBeenCalledTimes(1))
    expect(bridge.Donate).toHaveBeenCalledWith()
  })

  it('says so when the donation page could not be opened', async () => {
    installBridge({ Donate: vi.fn(() => Promise.reject(new Error('the desktop refused'))) })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Donate/ }))
    expect(await screen.findByText('The desktop refused')).toBeInTheDocument()
  })

  it('shows a refusal and lets it be dismissed', async () => {
    render(<App />)
    expect(await screen.findByText(noWindowShown)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(screen.queryByText(noWindowShown)).toBeNull()
  })
})
