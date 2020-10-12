package app

import (
	"fmt"
	"time"

	"github.com/bunyk/gokeybr/view"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

// used for testing
// j - type, k - untype
const cheating = false

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

func (a *App) Run() error {
	encoding.Register()
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = scr.Init()
	if err != nil {
		return err
	}
	defer scr.Fini()
	events := make(chan tcell.Event)
	go func() {
		for {
			ev := scr.PollEvent()
			events <- ev
		}
	}()
	for {
		view.Render(scr, a.ToDisplay())
		ev := <-events
		switch event := ev.(type) {
		case *tcell.EventKey:
			if !a.reduceEvent(event) {
				if cheating {
					a.InputPosition = 0
				}
				return nil
			}
		case *tcell.EventResize:
			scr.Sync()
			view.Render(scr, a.ToDisplay())
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
	if a.InputPosition == 0 {
		return "Typed nothing"
	}
	elapsed := a.Timeline[a.InputPosition-1]
	if elapsed == 0 {
		return "Speed of light! (actually, probably some error with timer)"
	}
	return fmt.Sprintf(
		"Typed %d characters in %4.1f seconds. Speed: %4.1f wpm\n",
		a.InputPosition, elapsed, float64(a.InputPosition)/elapsed*60.0/5.0,
	)
}

// Compute number of typed lines
func (a App) LinesTyped() int {
	lt := 0
	for _, c := range a.Text[:a.InputPosition] {
		if c == '\n' {
			lt++
		}
	}
	if a.InputPosition == len(a.Text) {
		lt++
	}
	return lt
}

// Return true when should continue loop
func (a *App) reduceEvent(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		return false
	}

	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
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
func (a *App) reduceCharInput(ev *tcell.EventKey) bool {
	var ch rune
	if ev.Key() == tcell.KeyRune {
		ch = ev.Rune()
	} else if ev.Key() == tcell.KeyEnter || ev.Key() == tcell.KeyCtrlJ {
		ch = '\n'
	}
	if ch == 0 {
		return true
	}
	if a.StartedAt.IsZero() {
		a.StartedAt = time.Now()
	}

	if cheating { // always type correct :)
		if ch == 'j' {
			a.InputPosition += 3
		}
		if ch == 'k' {
			a.InputPosition -= 3
		}
		if a.InputPosition < 0 {
			a.InputPosition = 0
		}
		return a.InputPosition < len(a.Text)
	}
	if ch == a.Text[a.InputPosition] && len(a.ErrorInput) == 0 { // correct
		a.Timeline[a.InputPosition] = time.Since(a.StartedAt).Seconds()
		a.InputPosition++
	} else { // wrong
		a.ErrorInput = append(a.ErrorInput, ch)
	}
	return a.InputPosition < len(a.Text)
}
