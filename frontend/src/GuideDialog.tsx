// The Guide: what each control is, how to record quickly and the rules the
// window cannot state for itself. Its words live in guideContent.ts; this file
// only draws them.

import { guideSections } from './guideContent'
import { Modal } from './Modal'

interface Props {
  onClose: () => void
}

export function GuideDialog({ onClose }: Props) {
  return (
    <Modal labelId="guide-title" role="dialog" onClose={onClose} focusAction>
      <h2 id="guide-title">How SymChit works</h2>
      <div className="guide-body">
        {guideSections.map((section) => (
          <section className="guide-section" key={section.heading}>
            <h3>{section.heading}</h3>
            {section.intro && <p className="guide-intro">{section.intro}</p>}
            {section.paragraphs?.map((text) => (
              <p key={text}>{text}</p>
            ))}
            {section.entries?.map((entry) => (
              <p className="guide-entry" key={entry.name}>
                <img className="guide-icon" src={entry.icon} alt="" draggable={false} />
                <span>
                  <b>{entry.name}</b>: {entry.text}
                </span>
              </p>
            ))}
            {section.rules?.map((rule) => (
              <p className="guide-rule" key={rule.title}>
                <b>{rule.title}</b> {rule.text}
              </p>
            ))}
          </section>
        ))}
      </div>
      <div className="actions">
        <button type="button" className="close-guide" onClick={onClose}>
          Close
        </button>
      </div>
    </Modal>
  )
}
