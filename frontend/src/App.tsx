// The shell: the nav band, the pane it chooses, the status line and the
// problem banner shown when the record could not be opened (FR-062).

import { useCallback, useEffect, useState } from 'react'
import { api, type About, type State } from './api'
import { AboutDialog } from './Dialog'
import { GuideDialog } from './GuideDialog'
import { HistoryPane } from './HistoryPane'
import { ReceiptPane } from './ReceiptPane'
import { RecordPane } from './RecordPane'
import { SymptomsPane } from './SymptomsPane'
import { counted } from './words'
import recordIcon from './assets/icons/record.png'
import historyIcon from './assets/icons/history.png'
import receiptIcon from './assets/icons/receipt.png'
import symptomsIcon from './assets/icons/symptoms.png'
import exportIcon from './assets/icons/export.png'
import importIcon from './assets/icons/import.png'
import guideIcon from './assets/icons/help-info.png'
import aboutIcon from './assets/icons/application-icon.png'

type Pane = 'record' | 'history' | 'receipt' | 'symptoms'

const panes: { pane: Pane; label: string; icon: string }[] = [
  { pane: 'record', label: 'Record', icon: recordIcon },
  { pane: 'history', label: 'History', icon: historyIcon },
  { pane: 'receipt', label: 'Receipt', icon: receiptIcon },
  { pane: 'symptoms', label: 'Symptoms', icon: symptomsIcon },
]

interface Message {
  tone: 'info' | 'error'
  text: string
}

interface BandButtonProps {
  label: string
  icon: string
  pressed?: boolean
  onClick: () => void
}

function BandButton({ label, icon, pressed, onClick }: BandButtonProps) {
  return (
    <button type="button" className="band-button" aria-pressed={pressed} onClick={onClick}>
      <img src={icon} alt="" />
      <span>{label}</span>
    </button>
  )
}

export function App() {
  const [state, setState] = useState<State | null>(null)
  const [pane, setPane] = useState<Pane>('record')
  const [message, setMessage] = useState<Message | null>(null)
  const [about, setAbout] = useState<About | null>(null)
  const [guide, setGuide] = useState(false)

  const refused = useCallback((text: string) => setMessage({ tone: 'error', text }), [])
  const tell = useCallback((text: string) => setMessage({ tone: 'info', text }), [])

  useEffect(() => {
    void api.state(refused).then((found) => found && setState(found))
  }, [refused])

  const exportRecord = async () => {
    const path = await api.exportRecord(refused)
    if (path) tell(`Your record was exported to ${path}.`)
  }

  const importRecord = async () => {
    const result = await api.importRecord(refused)
    if (!result?.chosen) return
    const skipped =
      result.skipped > 0 ? ` Skipped ${counted(result.skipped, 'event')} already in your record.` : ''
    tell(`Imported ${counted(result.added, 'event')}.${skipped}`)
  }

  const severities = state?.severities ?? []

  return (
    <div className="shell">
      <nav className="band" aria-label={state?.name ?? 'SymChit'}>
        <div className="band-group">
          {panes.map((entry) => (
            <BandButton key={entry.pane} label={entry.label} icon={entry.icon}
              pressed={pane === entry.pane} onClick={() => setPane(entry.pane)} />
          ))}
        </div>
        <div className="band-group">
          <BandButton label="Export" icon={exportIcon} onClick={exportRecord} />
          <BandButton label="Import" icon={importIcon} onClick={importRecord} />
          <BandButton label="Guide" icon={guideIcon} onClick={() => setGuide(true)} />
          <BandButton label="About" icon={aboutIcon}
            onClick={() => void api.about(refused).then((found) => found && setAbout(found))} />
        </div>
      </nav>

      {state?.problem && <p className="problem" role="alert">{state.problem}</p>}

      <div className={`status ${message?.tone ?? ''}`} role="status" aria-live="polite">
        {message && (
          <>
            <span>{message.text}</span>
            <button type="button" className="dismiss" onClick={() => setMessage(null)}>
              Dismiss
            </button>
          </>
        )}
      </div>

      <main>
        {pane === 'record' && <RecordPane severities={severities} refused={refused} tell={tell} />}
        {pane === 'history' && <HistoryPane severities={severities} refused={refused} tell={tell} />}
        {pane === 'receipt' && <ReceiptPane refused={refused} />}
        {pane === 'symptoms' && <SymptomsPane refused={refused} tell={tell} />}
      </main>

      {guide && <GuideDialog onClose={() => setGuide(false)} />}
      {about && <AboutDialog about={about} onClose={() => setAbout(null)} />}
    </div>
  )
}
