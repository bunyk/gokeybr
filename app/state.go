package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/bunyk/gokeybr/stats"
	"github.com/bunyk/gokeybr/view"
)

type State struct {
	Text          []rune
	Timeline      []float64
	InputPosition int
	ErrorInput    []rune
	StartedAt     time.Time
	IsTraining    bool
}

func newState(text string, isTraining bool) State {
	state := State{
		ErrorInput: make([]rune, 0, 20),
		Text:       []rune(text),
	}
	state.Timeline = make([]float64, len(state.Text))
	return state
}

func (s State) ToDisplay() view.DisplayableData {
	return view.DisplayableData{
		Header:    "Type the text below:", // TODO: add more data here
		DoneText:  s.Text[:s.InputPosition],
		WrongText: s.ErrorInput,
		TODOText:  s.Text[s.InputPosition:],
		StartedAt: s.StartedAt,
	}
}

func (s State) finish() {
	termbox.Close()
	if s.InputPosition > 0 {
		elapsed := s.Timeline[s.InputPosition-1]
		fmt.Printf(
			"Typed %d characters in %4.1f seconds. Speed: %4.1f wpm\n",
			s.InputPosition, elapsed, float64(s.InputPosition)/elapsed*60.0/5.0,
		)
		if err := stats.SaveSession(
			s.StartedAt,
			s.Text[:s.InputPosition],
			s.Timeline[:s.InputPosition],
			s.IsTraining,
		); err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
}

func (s *State) reduceEvent(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		s.finish()
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		s.reduceBackspace()
	default:
		s.reduceCharInput(ev)
	}
}

func (s *State) reduceBackspace() {
	if len(s.ErrorInput) == 0 {
		return
	}
	s.ErrorInput = s.ErrorInput[:len(s.ErrorInput)-1]
}

func (s *State) reduceCharInput(ev termbox.Event) {
	var ch rune
	if ev.Key == termbox.KeySpace {
		ch = ' '
	} else if ev.Key == termbox.KeyEnter || ev.Key == termbox.KeyCtrlJ {
		ch = '\n'
	} else {
		ch = ev.Ch
	}

	if ch == 0 {
		return
	}
	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now()
	}

	if ch == s.Text[s.InputPosition] && len(s.ErrorInput) == 0 { // correct
		s.Timeline[s.InputPosition] = time.Since(s.StartedAt).Seconds()
		s.InputPosition++
	} else { // wrong
		s.ErrorInput = append(s.ErrorInput, ch)
	}
	if s.InputPosition >= len(s.Text) {
		s.finish()
	}
}
