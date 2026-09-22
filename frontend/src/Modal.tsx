// The shell every dialog is drawn in: a backdrop, a labelled box and Escape to
// close. The dialogs themselves say what goes inside.

import { useEffect, useRef, type ReactNode } from 'react'

interface Props {
  labelId: string
  role: 'dialog' | 'alertdialog'
  onClose: () => void
  /** Whether the first button inside should take the focus when it opens. */
  focusAction?: boolean
  children: ReactNode
}

export function Modal({ labelId, role, onClose, focusAction, children }: Props) {
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
      <div className="dialog" role={role} aria-modal="true" aria-labelledby={labelId} ref={box}>
        {children}
      </div>
    </div>
  )
}
