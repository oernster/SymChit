// The symptoms held, each renamable (FR-028). A rename reaches every event
// using the symptom.

import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api, type Definition, type Refused } from './api'

interface Props {
  refused: Refused
  tell: (text: string) => void
}

export function SymptomsPane({ refused, tell }: Props) {
  const [symptoms, setSymptoms] = useState<Definition[]>([])
  const [renaming, setRenaming] = useState<number | null>(null)
  const [label, setLabel] = useState('')

  const load = useCallback(async () => {
    const held = await api.symptoms(refused)
    if (held) setSymptoms(held)
  }, [refused])

  useEffect(() => {
    void load()
  }, [load])

  const save = async (event: FormEvent, symptom: Definition) => {
    event.preventDefault()
    if (await api.rename(symptom.id, label, refused)) {
      tell(`Renamed ${symptom.label} to ${label}.`)
      setRenaming(null)
      await load()
    }
  }

  return (
    <section className="pane symptoms" aria-labelledby="symptoms-title">
      <h1 id="symptoms-title">Symptoms</h1>
      {symptoms.length === 0 && <p>No symptoms yet. Each one you record is kept here.</p>}
      <ul className="symptom-list">
        {symptoms.map((symptom) => (
          <li key={symptom.id}>
            {renaming === symptom.id ? (
              <form onSubmit={(event) => save(event, symptom)} className="rename">
                <label htmlFor={`rename-${symptom.id}`}>New name for {symptom.label}</label>
                <input id={`rename-${symptom.id}`} value={label} autoFocus
                  onChange={(event) => setLabel(event.target.value)} />
                <button type="button" onClick={() => setRenaming(null)}>Cancel</button>
                <button type="submit" className="primary">Save</button>
              </form>
            ) : (
              <>
                <span>{symptom.label}</span>
                <button type="button" onClick={() => {
                  setLabel(symptom.label)
                  setRenaming(symptom.id)
                }}>
                  Rename
                </button>
              </>
            )}
          </li>
        ))}
      </ul>
    </section>
  )
}
