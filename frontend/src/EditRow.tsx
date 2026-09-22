// One event opened for editing (FR-023, FR-024).
//
// The time is sent only when the user changed it. Left alone it is sent empty,
// which the backend reads as "keep the time held", so saving a corrected note
// can never move the event to the moment of the edit.

import { useState, type FormEvent } from 'react'
import { api, type EventEntry, type Refused } from './api'
import { EventFields, type EventValues } from './EventFields'

interface Props {
  event: EventEntry
  severities: string[]
  refused: Refused
  onSaved: (saved: EventEntry) => void
  onCancel: () => void
}

export function EditRow({ event, severities, refused, onSaved, onCancel }: Props) {
  const [values, setValues] = useState<EventValues>({
    symptom: event.symptom,
    severity: event.severity,
    note: event.note,
  })
  const [time, setTime] = useState(event.occurredAt)
  const prefix = `edit-${event.id}`

  const submit = async (form: FormEvent) => {
    form.preventDefault()
    const occurredAt = time === event.occurredAt ? '' : time
    const saved = await api.edit(event.id, { ...values, occurredAt }, refused)
    if (saved) onSaved(saved)
  }

  return (
    <form className="edit" onSubmit={submit} aria-label={`Edit the ${event.symptom} event`}>
      <EventFields
        idPrefix={prefix}
        values={values}
        onChange={setValues}
        severities={severities}
        refused={refused}
      />
      <div className="field">
        <label htmlFor={`${prefix}-time`}>Time</label>
        <input
          id={`${prefix}-time`}
          type="datetime-local"
          value={time}
          onChange={(change) => setTime(change.target.value)}
        />
      </div>
      <div className="actions">
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="primary">
          Save
        </button>
      </div>
    </form>
  )
}
