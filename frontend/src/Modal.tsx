// The shell every dialog is drawn in: a backdrop, a labelled box and Escape to
// close. The dialogs themselves say what goes inside.

import { useEffect, useRef, type ReactNode } from 'react'

interface Props {
  labelId: string
  role: 'dialog' | 'alertdialog'
  onClose: () => void
  /** Whether the first button inside should take the focus when it opens. */
  focusAction?: boolean
  /**
   * Whether the dialog holds a scrolling body. Its action row is then pinned
   * beneath that body, so Close stays where the reader expects it however tall
   * the content grows and never drifts as the body reads itself.
   */
  pinnedActions?: boolean
  children: ReactNode
}

export function Modal({ labelId, role, onClose, focusAction, pinnedActions, children }: Props) {
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])
  useEffect(() => {
    if (focusAction) box.current?.querySelector('button')?.focus()
  }, [focusAction])
  return (
    <div className="backdrop">
      <div
        className={pinnedActions ? 'dialog pinned-actions' : 'dialog'}
        role={role}
        aria-modal="true"
        aria-labelledby={labelId}
        ref={box}
      >
        {children}
      </div>
    </div>
  )
}
