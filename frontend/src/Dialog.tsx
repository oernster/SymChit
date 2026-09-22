// The confirmation shown before anything is deleted (FR-025) and About
// (FR-066). Both are drawn in the shared Modal.

import { useEffect, useRef } from 'react'
import type { About } from './api'
import { Modal } from './Modal'
import { useAutoScroll } from './useAutoScroll'
import crest from './assets/icons/application-icon.png'

interface ConfirmProps {
  text: string
  confirmLabel: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({ text, confirmLabel, onConfirm, onCancel }: ConfirmProps) {
  const cancel = useRef<HTMLButtonElement>(null)
  // Cancel takes the focus: the safe answer is the one a stray Enter reaches.
  useEffect(() => cancel.current?.focus(), [])
  return (
    <Modal labelId="confirm-text" role="alertdialog" onClose={onCancel}>
      <p id="confirm-text">{text}</p>
      <div className="actions">
        <button type="button" ref={cancel} onClick={onCancel}>
          Cancel
        </button>
        <button type="button" className="danger" onClick={onConfirm}>
          {confirmLabel}
        </button>
      </div>
    </Modal>
  )
}

interface AboutProps {
  about: About
  onClose: () => void
}

export function AboutDialog({ about, onClose }: AboutProps) {
  // The credits run past the dialog's height on a short window, so the body is
  // the scroller and reads itself down gently, with Close pinned beneath it.
  const autoScroll = useAutoScroll()
  return (
    <Modal labelId="about-title" role="dialog" onClose={onClose} focusAction pinnedActions>
      <div className="dialog-body" ref={autoScroll}>
        <img className="crest" src={crest} alt="" />
        <h2 id="about-title">
          {about.name} {about.version}
        </h2>
        <p>By {about.author}</p>
        <p className="statement">{about.statement}</p>
        <h3>Built with</h3>
        <ul className="credits">
          {about.credits.map((credit) => (
            <li key={credit.work}>
              {credit.work}: {credit.licence}, {credit.holder}
            </li>
          ))}
        </ul>
      </div>
      <div className="actions">
        <button type="button" onClick={onClose}>
          Close
        </button>
      </div>
    </Modal>
  )
}
