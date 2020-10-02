package app

import (
	"fmt"
	"time"

	"github.com/bunyk/gokeybr/view"
	"github.com/nsf/termbox-go"
)

// App holds whole app state
type App struct {
	Text          []rune
	Timeline      []float64
	InputPosition int
	ErrorInput    []rune
	StartedAt     time.Time
}

func New(text string) *App {
	a := &App{}
	a.ErrorInput = make([]rune, 0, 20)
	a.Text = []rune(text)
	a.Timeline = make([]float64, len(a.Text))
	return a
}

func initTermbox() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	termbox.SetInputMode(termbox.InputEsc)
	return nil
}

func (a *App) Run() error {
	if err := initTermbox(); err != nil {
		return err
	}
	defer termbox.Close()
	events := make(chan termbox.Event)
	go func() {
		for {
			ev := termbox.PollEvent()
			events <- ev
		}
	}()
	for {
		view.Render(a.ToDisplay())
		ev := <-events
		switch ev.Type {
		case termbox.EventKey:
			if !a.reduceEvent(ev) {
				return nil
			}
		case termbox.EventError:
			return ev.Err
		}
	}
}

func (a App) ToDisplay() view.DisplayableData {
	return view.DisplayableData{
		Header:    "Type the text below:", // TODO: add more data here
		DoneText:  a.Text[:a.InputPosition],
		WrongText: a.ErrorInput,
		TODOText:  a.Text[a.InputPosition:],
		StartedAt: a.StartedAt,
	}
}

func (a App) Summary() string {
	elapsed := a.Timeline[a.InputPosition-1]
	if elapsed > 0 {
		return fmt.Sprintf(
			"Typed %d characters in %4.1f seconds. Speed: %4.1f wpm\n",
			a.InputPosition, elapsed, float64(a.InputPosition)/elapsed*60.0/5.0,
		)
	}
	return ""
}

// Return true when should continue loop
func (a *App) reduceEvent(ev termbox.Event) bool {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		return false
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		a.reduceBackspace()
	default:
		return a.reduceCharInput(ev)
	}
	return true
}

func (a *App) reduceBackspace() {
	if len(a.ErrorInput) == 0 {
		return
	}
	a.ErrorInput = a.ErrorInput[:len(a.ErrorInput)-1]
}

// Return true when should continue loop
func (a *App) reduceCharInput(ev termbox.Event) bool {
	var ch rune
	if ev.Key == termbox.KeySpace {
		ch = ' '
	} else if ev.Key == termbox.KeyEnter || ev.Key == termbox.KeyCtrlJ {
		ch = '\n'
	} else {
		ch = ev.Ch
	}

	if ch == 0 {
		return true
	}
	if a.StartedAt.IsZero() {
		a.StartedAt = time.Now()
	}

	if ch == a.Text[a.InputPosition] && len(a.ErrorInput) == 0 { // correct
		a.Timeline[a.InputPosition] = time.Since(a.StartedAt).Seconds()
		a.InputPosition++
	} else { // wrong
		a.ErrorInput = append(a.ErrorInput, ch)
	}
	return a.InputPosition < len(a.Text)
}
