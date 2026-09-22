package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/oernster/symchit/internal/application"
	"github.com/oernster/symchit/internal/domain"
	"github.com/oernster/symchit/internal/product"
)

// errInternal is what the page is told when a bound method panics. The stack
// goes to the log; the page gets a sentence it can show.
var errInternal = errors.New(product.Name + " hit an internal fault; the details are in its log")

// fileChooser asks the user where to write or read an export. An empty path
// with no error means the user cancelled.
type fileChooser interface {
	SavePath(suggested string) (string, error)
	OpenPath() (string, error)
}

// Services is what the facade drives: one application service per concern.
type Services struct {
	Recorder application.Recorder
	Editor   application.Editor
	History  application.History
	Transfer application.Transfer
}

// App is the facade the window calls. It converts between the page's shapes
// and the application's; it owns nothing else, since every rule lives below it.
type App struct {
	ctx      context.Context
	zone     *time.Location
	clock    application.Clock
	version  string
	problem  string
	services Services
	chooser  fileChooser
	close    func() error
}

// newApp answers the facade. problem is the reason the record could not be
// opened, empty when it opened; the page shows it (FR-062).
func newApp(services Services, clock application.Clock, zone *time.Location,
	version, problem string, chooser fileChooser, close func() error) *App {
	return &App{
		zone: zone, clock: clock, version: version, problem: problem,
		services: services, chooser: chooser, close: close,
	}
}

// startup keeps the window's context for the dialogs.
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// shutdown closes the record.
func (a *App) shutdown(context.Context) {
	if err := a.close(); err != nil {
		fmt.Fprintf(os.Stderr, "closing the record: %v\n", err)
	}
}

// guard turns a panic in a bound method into errInternal, writing the stack to
// the log. A bound method runs on a goroutine Wails owns, so the recover has to
// sit here, in the method itself.
func guard(err *error) {
	if recovered := recover(); recovered != nil {
		fmt.Fprintf(os.Stderr, "panic in a bound method: %v\n%s", recovered, debug.Stack())
		*err = errInternal
	}
}

// State answers what the page needs to start.
func (a *App) State() (state StateDTO, err error) {
	defer guard(&err)
	severities := []string{}
	for _, severity := range domain.Severities() {
		severities = append(severities, severity.String())
	}
	return StateDTO{
		Name: product.Name, Version: a.version, Problem: a.problem, Severities: severities,
	}, nil
}

// Now answers the current local time in the form the time field takes.
func (a *App) Now() (now string, err error) {
	defer guard(&err)
	return domain.FormatLocal(a.clock.Now(), a.zone), nil
}

// About answers the About dialog.
func (a *App) About() (about AboutDTO, err error) {
	defer guard(&err)
	return aboutDTO(application.NewAbout(a.version)), nil
}

// occurrence reads an occurrence time from the page; nil when none was given.
func (a *App) occurrence(text string) (*time.Time, error) {
	if text == "" {
		return nil, nil
	}
	parsed, err := domain.ParseLocal(text, a.zone)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// optionalDate reads a date from the page; the zero date when none was given.
func optionalDate(text string) (domain.Date, error) {
	if text == "" {
		return domain.Date{}, nil
	}
	return domain.ParseDate(text)
}

// eventIDs converts ids from the page.
func eventIDs(ids []int64) []domain.EventID {
	out := make([]domain.EventID, 0, len(ids))
	for _, id := range ids {
		out = append(out, domain.EventID(id))
	}
	return out
}

// eventDTOs converts events for the page.
func (a *App) eventDTOs(events []domain.Event) []EventDTO {
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, a.eventDTO(event))
	}
	return out
}
