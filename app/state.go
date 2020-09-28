// only pure code in this file (no side effects)
package app

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/view"

	"github.com/nsf/termbox-go"
)

type Phrase struct {
	Text      string
	Input     string
	StartedAt time.Time
}

type State struct {
	PhraseGenerator phrase.Generator
	Phrase          Phrase
}

func newState(generator phrase.Generator) State {
	state := State{PhraseGenerator: generator}
	state.resetPhrase()

	return state
}

func (s State) ToDisplay() view.DisplayableData {
	done, wrong, todo := compareInput(s.Phrase.Text, s.Phrase.Input)
	return view.DisplayableData{
		Header:    "Type the text below", // TODO: add more data here
		DoneText:  done,
		WrongText: wrong,
		TODOText:  todo,
		StartedAt: s.Phrase.StartedAt,
	}
}

// compareInput with required text and return
// properly typed part, wrongly typed part, and part that is left to type
func compareInput(text, input string) (done, wrong, todo []rune) {
	ri := []rune(input)
	li := len(ri)
	rt := []rune(text)
	for i, tr := range rt {
		if i >= li {
			return ri, nil, rt[i:]
		}
		if tr != ri[i] {
			done = rt[:i]
			wrong = ri[i:]
			if li < len(rt) {
				todo = rt[li:]
			}
			return
		}
	}
	done = rt
	if li > len(rt) {
		wrong = ri[len(rt):]
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

	if s.Phrase.StartedAt.IsZero() {
		s.Phrase.StartedAt = time.Now()
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		s.reduceBackspace()
	case termbox.KeyCtrlF:
		s.resetPhrase()
		s.reduceEnter()
	default:
		s.reduceCharInput(ev)
	}
}

func (s *State) reduceBackspace() {
	if len(s.Phrase.Input) == 0 {
		return
	}
	_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
	s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
}

func (s *State) reduceEnter() {
	if s.Phrase.Input != s.Phrase.Text {
		return
	}
	s.resetPhrase()
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

	exp := s.Phrase.expected()
	if ch == exp {
		s.Phrase.Input += string(ch)
		return
	}

	// normal mode
	s.Phrase.Input += string(ch)
}

func (s *State) resetPhrase() {
	phrase := s.PhraseGenerator.Phrase()
	s.Phrase = Phrase{
		Text: phrase,
	}
}

func (p *Phrase) expected() rune {
	if len(p.Input) >= len(p.Text) {
		return 0
	}

	expected, _ := utf8.DecodeRuneInString(p.Text[len(p.Input):])
	return expected
}

func Exit(status int, message string) {
	if termbox.IsInit {
		termbox.Close()
	}
	fmt.Println(message)
	os.Exit(status)
}
