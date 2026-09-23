package application

import (
	"time"

	"github.com/oernster/symdiary/internal/domain"
)

// RecordForm is what the user entered on the recording form.
type RecordForm struct {
	Symptom string
	// OccurredAt is the time the user set; nil means now (FR-001, FR-003).
	OccurredAt *time.Time
	Severity   domain.Severity
	Note       string
}

// Recorder records new events.
type Recorder struct {
	store Store
	clock Clock
}

// NewRecorder answers a recorder over a store and a clock.
func NewRecorder(store Store, clock Clock) Recorder {
	return Recorder{store: store, clock: clock}
}

// Record stores one event. The recording instant is read once, so the default
// occurrence and the recorded-at are the same instant.
func (r Recorder) Record(form RecordForm) (domain.Event, error) {
	if err := domain.CheckLabel(form.Symptom); err != nil {
		return domain.Event{}, err
	}
	now := r.clock.Now()
	occurred := now
	if form.OccurredAt != nil {
		occurred = *form.OccurredAt
	}
	if err := domain.CheckOccurrence(occurred, now); err != nil {
		return domain.Event{}, err
	}
	saved, err := r.store.AddEvent(EventInput{
		Label:      form.Symptom,
		Key:        domain.KeyOf(form.Symptom),
		OccurredAt: occurred,
		RecordedAt: now,
		Severity:   form.Severity,
		Note:       form.Note,
	})
	if err != nil {
		return domain.Event{}, because(ErrNotSaved, err)
	}
	return saved, nil
}
