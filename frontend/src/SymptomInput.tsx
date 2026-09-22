// The symptom field: free text with the held symptoms suggested as the user
// types (FR-011, FR-012).
//
// The keys, so recording a held symptom takes a few strokes (NFR-USE-001): type
// a few letters, Down to the suggestion, Enter to take it, Enter again to record.
// Enter with no suggestion highlighted is left to the form, which records.

import { useEffect, useState, type KeyboardEvent, type Ref } from 'react'
import { api, type Definition, type Refused } from './api'

interface Props {
  id: string
  value: string
  onChange: (value: string) => void
  refused: Refused
  inputRef?: Ref<HTMLInputElement>
}

export function SymptomInput({ id, value, onChange, refused, inputRef }: Props) {
  const [options, setOptions] = useState<Definition[]>([])
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(-1)
  const listId = `${id}-suggestions`

  useEffect(() => {
    if (!open) return
    let live = true
    void api.suggest(value, refused).then((found) => {
      if (live && found) {
        setOptions(found)
        setActive(-1)
      }
    })
    return () => {
      live = false
    }
  }, [value, open, refused])

  const pick = (definition: Definition) => {
    onChange(definition.label)
    setOpen(false)
  }

  const showing = open && options.length > 0

  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      if (!open) {
        setOpen(true)
        return
      }
      setActive((at) => Math.min(at + 1, options.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActive((at) => Math.max(at - 1, -1))
    } else if (event.key === 'Enter' && showing && active >= 0) {
      event.preventDefault()
      pick(options[active])
    } else if (event.key === 'Escape' && showing) {
      event.preventDefault()
      setOpen(false)
    }
  }

  return (
    <div className="combo">
      <input
        id={id}
        ref={inputRef}
        type="text"
        autoComplete="off"
        role="combobox"
        aria-autocomplete="list"
        aria-expanded={showing}
        aria-controls={listId}
        aria-activedescendant={showing && active >= 0 ? `${listId}-${active}` : undefined}
        value={value}
        onChange={(event) => {
          onChange(event.target.value)
          setOpen(true)
        }}
        onKeyDown={onKeyDown}
        onBlur={() => setOpen(false)}
      />
      {showing && (
        <ul id={listId} role="listbox" className="suggestions">
          {options.map((definition, index) => (
            <li
              key={definition.id}
              id={`${listId}-${index}`}
              role="option"
              aria-selected={index === active}
              onMouseDown={(event) => {
                event.preventDefault()
                pick(definition)
              }}
            >
              {definition.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
