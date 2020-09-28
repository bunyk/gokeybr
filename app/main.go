package app

import (
	"time"

	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/view"
	"github.com/nsf/termbox-go"
)

type App struct {
	state State
}

// Parameters define arguments with which program started
type Parameters struct {
	Sourcefile   string // From where to read training text
	Sourcetext   string // Training text itself (optional)
	Mode         string // Treat training text as paragraphs, or set of words to create random texts
	PhraseLength int    // default lenght for generated phrase
}

func New(params Parameters) (*App, error) {
	a := &App{}
	var err error
	generator, err := phrase.NewGenerator(
		params.Sourcefile, params.Sourcetext, params.Mode, params.PhraseLength,
	)
	if err != nil {
		return a, err
	}
	a.state = newState(generator)
	if err != nil {
		return a, err
	}
	err = initTermbox()
	if err != nil {
		return a, err
	}
	return a, nil
}

func initTermbox() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	termbox.SetInputMode(termbox.InputEsc)
	// This is done to update rerender timer values, etc.
	go func() {
		for range time.Tick(100 * time.Millisecond) {
			termbox.Interrupt()
			// Interrupt an in-progress call to PollEvent
			// by causing it to return EventInterrupt.
		}
	}()
	return nil
}

func (a *App) Run() {
	for {
		view.Render(a.state.ToDisplay())
		for _, msg := range waitForEvent() {
			a.state = reduceEvent(a.state, msg, time.Now())
		}
	}
}

func waitForEvent() []termbox.Event {
	ev := termbox.PollEvent()
	switch ev.Type {
	case termbox.EventKey:
		return []termbox.Event{ev}
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventInterrupt:
	}

	return []termbox.Event{}
}
