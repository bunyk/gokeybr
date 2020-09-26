package app

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const (
	ScoreHighlightDuration = time.Second * 3
)

const (
	black = termbox.ColorBlack
	red   = termbox.ColorRed
	green = termbox.ColorGreen
	white = termbox.ColorWhite
)

type Align int

const (
	Left Align = iota
	Center
	Right
)

type printSpec struct {
	text  string
	x     int
	y     int
	fg    termbox.Attribute
	bg    termbox.Attribute
	align Align
}

func render(s State) {
	_ = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	done, wrong, todo := compareInput(s.Phrase.Text, s.Phrase.Input)
	write3colors(done, wrong, todo, 2, 2, w-5)

	// Stats:
	secondsText := text("Go!")
	if !s.Phrase.StartedAt.IsZero() {
		seconds := time.Since(s.Phrase.StartedAt).Seconds()
		secondsText = text("%4.1f seconds", seconds)
	}
	write(secondsText.X(w/2 + 1).Y(h - 1).Align(Center))

	write(text("%3d errors", s.Phrase.Errors).
		X(w - 1).
		Y(h - 1).
		Align(Right).
		Fg(termbox.ColorDefault),
	)
}

// Compare input with required text and return properly typed part, wrongly typed, and part to type
func compareInput(text, input string) (done, wrong, todo string) {
	ri := []rune(input)
	li := len(ri)
	rt := []rune(text)
	for i, tr := range rt {
		if i >= li {
			return input, "", string(rt[i:])
		}
		if tr != ri[i] {
			done = string(rt[:i])
			wrong = string(ri[i:])
			if li < len(rt) {
				todo = string(rt[li:])
			}
			return
		}
	}
	done = text
	if li > len(rt) {
		wrong = string(ri[len(rt):])
	}
	return
}

// errorOffset returns position in bytes and then in runes, where

func text(t string, args ...interface{}) *printSpec {
	s := &printSpec{}
	if len(args) > 0 {
		s.text = fmt.Sprintf(t, args...)
	} else {
		s.text = t
	}
	return s
}

func write(spec *printSpec) {
	if spec == nil {
		return
	}
	var x int
	switch spec.align {
	case Left:
		x = spec.x
	case Center:
		x = spec.x - utf8.RuneCountInString(spec.text)/2
	case Right:
		x = spec.x - utf8.RuneCountInString(spec.text)
	}

	for _, c := range spec.text {
		termbox.SetCell(x, spec.y, c, spec.fg, spec.bg)
		x++
	}
}

func write3colors(done, wrong, todo string, x, y, w int) {
	cursorX := x
	cursorY := y
	putS := func(s string, fg, bg termbox.Attribute) {
		for _, c := range s {
			termbox.SetCell(cursorX, cursorY, c, fg, bg)
			cursorX++
			if cursorX >= x+w {
				cursorX = x
				cursorY++
			}
		}
	}

	putS(spaced(done), green, 0)
	putS(spaced(wrong), black, red)
	putS(todo, white, 0)
}

func spaced(s string) string {
	return strings.ReplaceAll(s, " ", "␣")
}

func (p *printSpec) Align(align Align) *printSpec {
	p.align = align
	return p
}

func (p *printSpec) X(x int) *printSpec {
	p.x = x
	return p
}

func (p *printSpec) Y(y int) *printSpec {
	p.y = y
	return p
}

func (p *printSpec) Fg(fg termbox.Attribute) *printSpec {
	p.fg = fg
	return p
}

func (p *printSpec) Bg(bg termbox.Attribute) *printSpec {
	p.bg = bg
	return p
}
