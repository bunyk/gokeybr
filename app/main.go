package app

import (
	"log"
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
	Codelines    bool   // Treat training text as code?
	PhraseLength int    // default lenght for generated phrase
}

func New(params Parameters) (*App, error) {
	a := &App{}

	a.state = InitState(params)
	if err := InitTermbox(); err != nil {
		return a, err
	}
	return a, nil
}

func InitTermbox() error {
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

func InitState(params Parameters) State {
	state := *NewState(phrase.DefaultGenerator)

	if len(params.Sourcetext) > 0 {
		state.PhraseGenerator = phrase.StaticGenerator{Text: params.Sourcetext}
		state = resetPhrase(state, false)
		state.Header = params.Sourcetext
	} else {
		state.Header = params.Sourcefile
		items, err := phrase.ReadFileLines(params.Sourcefile)
		if err != nil {
			log.Fatal(err)
		}
		// state = reduceDatasource(state, data, params.Codelines)
		if params.Codelines {
			items = phrase.FilterWords(items, `^[^/][^/]`, 80)
			state.PhraseGenerator = &phrase.SequentialLineGenerator{Lines: items}
		} else {
			items = phrase.FilterWords(items, `^[a-z]+$`, 8)
			state.PhraseGenerator = phrase.NewRandomGenerator(items, params.PhraseLength)
		}

		if len(items) == 0 {
			log.Fatal("datafile contains no usable data")
		}

		state = resetPhrase(state, false)
	}

	return state
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
