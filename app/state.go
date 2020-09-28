// only pure code in this file (no side effects)
package app

import (
	"fmt"
	"os"
	"time"

	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/view"

	"github.com/nsf/termbox-go"
)

type State struct {
	PhraseGenerator phrase.Generator
	Text            []rune
	Input           []rune
	StartedAt       time.Time
}

func newState(generator phrase.Generator) State {
	state := State{PhraseGenerator: generator}
	state.resetPhrase()

	return state
}

func (s State) ToDisplay() view.DisplayableData {
	done, wrong, todo := compareInput(s.Text, s.Input)
	return view.DisplayableData{
		Header:    "Type the text below:", // TODO: add more data here
		DoneText:  done,
		WrongText: wrong,
		TODOText:  todo,
		StartedAt: s.StartedAt,
	}
}

// compareInput with required text and return
// properly typed part, wrongly typed part, and part that is left to type
func compareInput(text, input []rune) (done, wrong, todo []rune) {
	li := len(input)
	for i, tr := range text {
		if i >= li {
			return input, nil, text[i:]
		}
		if tr != input[i] {
			done = text[:i]
			wrong = input[i:]
			if li < len(text) {
				todo = text[li:]
			}
			return
		}
	}
	done = text
	if li > len(text) {
		wrong = input[len(text):]
	}
	return
}

func (s *State) finish() {
	Exit(0, "bye!")
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
	if len(s.Input) == 0 {
		return
	}
	s.Input = s.Input[:len(s.Input)-1]
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

	s.Input = append(s.Input, ch)
}

func (s *State) resetPhrase() {
	phrase := s.PhraseGenerator.Phrase()
	s.Text = []rune(phrase)
	s.Input = nil
	s.StartedAt = time.Time{}
}

func Exit(status int, message string) {
	if termbox.IsInit {
		termbox.Close()
	}
	fmt.Println(message)
	os.Exit(status)
}
