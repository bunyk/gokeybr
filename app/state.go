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
	Text       string
	Input      string
	StartedAt  time.Time
	FailedAt   time.Time
	FinishedAt time.Time
	Errors     int
}

type State struct {
	Header          string
	PhraseGenerator phrase.Generator
	Phrase          Phrase
}

func (s State) ToDisplay() view.DisplayableData {
	done, wrong, todo := compareInput(s.Phrase.Text, s.Phrase.Input)
	return view.DisplayableData{
		Header:    s.Header,
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

func reduceEvent(s State, ev termbox.Event, now time.Time) State {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		Exit(0, "bye!")
		return s
	}

	if s.Phrase.StartedAt.IsZero() {
		s.Phrase.StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		return reduceBackspace(s)
	case termbox.KeyCtrlF:
		s = resetPhrase(s, true)
	case termbox.KeyEnter, termbox.KeyCtrlJ:
		return reduceEnter(s, now)
	default:
		return reduceCharInput(s, ev, now)
	}

	return s
}

func reduceBackspace(s State) State {
	if len(s.Phrase.Input) == 0 {
		return s
	}

	_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
	s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
	return s
}

func reduceEnter(s State, now time.Time) State {
	if s.Phrase.Input != s.Phrase.Text {
		return s
	}

	s.Phrase.FinishedAt = now

	s = resetPhrase(s, false)

	return s
}

func reduceCharInput(s State, ev termbox.Event, now time.Time) State {
	var ch rune
	if ev.Key == termbox.KeySpace {
		ch = ' '
	} else {
		ch = ev.Ch
	}

	if ch == 0 {
		return s
	}

	exp := s.Phrase.expected()
	if ch == exp {
		s.Phrase.Input += string(ch)
		return s
	}

	s.Phrase.Errors++
	s.Phrase.FailedAt = now

	// normal mode
	s.Phrase.Input += string(ch)
	return s
}

func resetPhrase(state State, forceNext bool) State {
	if forceNext {
		state.PhraseGenerator.Phrase() // Just to update seed
	}
	phrase := state.PhraseGenerator.Phrase()
	state.Phrase = *NewPhrase(phrase)

	return state
}

func NewState(phraseGenerator phrase.Generator) *State {
	s := resetPhrase(State{
		PhraseGenerator: phraseGenerator,
	}, false)

	return &s
}

func NewPhrase(text string) *Phrase {
	return &Phrase{
		Text: text,
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
	termbox.Close()
	fmt.Println(message)
	os.Exit(status)
}
