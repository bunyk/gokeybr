package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/stats"
	"github.com/bunyk/gokeybr/view"
)

type State struct {
	PhraseGenerator phrase.Generator
	Text            []rune
	Timeline        []float64
	InputPosition   int
	ErrorInput      []rune
	StartedAt       time.Time
}

func newState(generator phrase.Generator) State {
	state := State{
		PhraseGenerator: generator,
		ErrorInput:      make([]rune, 0, 20),
	}
	state.resetPhrase()

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
		if err := stats.SaveSession(s.StartedAt, s.Text[:s.InputPosition], s.Timeline[:s.InputPosition]); err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
}

func (s *State) reduceEvent(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		s.finish()
	}

	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now()
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		s.reduceBackspace()
	case termbox.KeyCtrlF:
		s.resetPhrase()
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

	if ch == s.Text[s.InputPosition] { // correct
		s.Timeline[s.InputPosition] = time.Since(s.StartedAt).Seconds()
		s.InputPosition++
	} else { // wrong
		s.ErrorInput = append(s.ErrorInput, ch)
	}
	if s.InputPosition >= len(s.Text) {
		s.finish()
	}
}

func (s *State) resetPhrase() {
	phrase := s.PhraseGenerator.Phrase()
	s.Text = []rune(phrase)
	s.Timeline = make([]float64, len(s.Text))
	s.InputPosition = 0
	s.ErrorInput = s.ErrorInput[:0] // Clear errors
	s.StartedAt = time.Time{}
}
