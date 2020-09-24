// only pure code in this file (no side effects)
package main

import (
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const PhraseLength = 100

type Typo struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type Round struct {
	StartedAt  time.Time
	FailedAt   time.Time
	FinishedAt time.Time
	Errors     int
	Typos      []Typo
}

type Phrase struct {
	Text   string
	Input  string
	Rounds [1]Round
}

type State struct {
	Codelines       bool
	Seed            int64
	PhraseGenerator PhraseFunc
	Phrase          Phrase
	Repeat          bool
}

func reduce(s State, msg Message, now time.Time) (State, []Command) {
	switch m := msg.(type) {
	case error:
		return s, []Command{Exit{GoodbyeMessage: m.Error()}}
	case Datasource:
		return reduceDatasource(s, m.Data, now)
	case termbox.Event:
		return reduceEvent(s, m, now)
	}

	return s, Noop
}

func reduceEvent(s State, ev termbox.Event, now time.Time) (State, []Command) {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		return s, []Command{Exit{GoodbyeMessage: "bye!"}}
	}

	if s.Phrase.CurrentRound().StartedAt.IsZero() {
		s.Phrase.CurrentRound().StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		return reduceBackspace(s)
	case termbox.KeyCtrlF:
		s = resetPhrase(s, true)
	case termbox.KeyCtrlR:
		s.Repeat = !s.Repeat
	case termbox.KeyEnter, termbox.KeyCtrlJ:
		return reduceEnter(s, now)
	default:
		return reduceCharInput(s, ev, now)
	}

	return s, Noop
}

func reduceBackspace(s State) (State, []Command) {
	if len(s.Phrase.Input) == 0 {
		return s, Noop
	}

	_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
	s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
	return s, Noop
}

func reduceEnter(s State, now time.Time) (State, []Command) {
	if s.Phrase.Input != s.Phrase.Text {
		return s, Noop
	}

	s.Phrase.CurrentRound().FinishedAt = now

	s = resetPhrase(s, false)

	return s, []Command{Interrupt{ScoreHighlightDuration}}
}

func reduceCharInput(s State, ev termbox.Event, now time.Time) (State, []Command) {
	var ch rune
	if ev.Key == termbox.KeySpace {
		ch = ' '
	} else {
		ch = ev.Ch
	}

	if ch == 0 {
		return s, Noop
	}

	exp := s.Phrase.expected()
	if ch == exp {
		s.Phrase.Input += string(ch)
		return s, Noop
	}

	if exp != 0 {
		s.Phrase.CurrentRound().Typos = append(
			s.Phrase.CurrentRound().Typos, Typo{
				Expected: string(exp),
				Actual:   string(ch),
			})
	}

	s.Phrase.CurrentRound().Errors++
	s.Phrase.CurrentRound().FailedAt = now

	// normal mode
	s.Phrase.Input += string(ch)
	return s, Noop
}

func reduceDatasource(state State, data []byte, now time.Time) (State, []Command) {
	var generator func([]string) PhraseFunc

	items := readLines(data)
	if state.Codelines {
		items = filterWords(items, `^[^/][^/]`, 80)
		generator = SequentialLine
	} else {
		items = filterWords(items, `^[a-z]+$`, 8)
		generator = func(words []string) PhraseFunc { return RandomPhrase(words, PhraseLength) }
		state.Seed = now.UnixNano()
	}

	if len(items) == 0 {
		return state, []Command{Exit{GoodbyeMessage: "datafile contains no usable data"}}
	}

	state.PhraseGenerator = generator(items)

	return resetPhrase(state, false), Noop
}

func resetPhrase(state State, forceNext bool) State {
	if !state.Repeat || forceNext {
		next, _ := state.PhraseGenerator(state.Seed)
		state.Seed = next
	}
	_, phrase := state.PhraseGenerator(state.Seed)
	state.Phrase = *NewPhrase(phrase)

	return state
}

func errorOffset(text string, input string) (int, int) {
	runeOffset := 0
	for i, tr := range text {
		if i >= len(input) {
			return len(input), runeOffset
		}

		ir, _ := utf8.DecodeRuneInString(input[i:])
		if ir != tr {
			return i, runeOffset
		}

		runeOffset++
	}

	return min(len(input), len(text)), runeOffset
}

func NewState(seed int64, phraseGenerator PhraseFunc) *State {
	s := resetPhrase(State{
		PhraseGenerator: phraseGenerator,
		Seed:            seed,
	}, false)

	return &s
}

func NewPhrase(text string) *Phrase {
	return &Phrase{
		Text: text,
	}
}

func (p *Phrase) CurrentRound() *Round {
	return &p.Rounds[0]
}

func (p *Phrase) expected() rune {
	if len(p.Input) >= len(p.Text) {
		return 0
	}

	expected, _ := utf8.DecodeRuneInString(p.Text[len(p.Input):])
	return expected
}
