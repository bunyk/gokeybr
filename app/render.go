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

	byteOffset, runeOffset := errorOffset(s.Phrase.Text, s.Phrase.Input)

	x := (w / 2) - (utf8.RuneCountInString(s.Phrase.Text) / 2)
	write(text(s.Phrase.Text + string('⏎')).X(x).Y(h / 2).Fg(white))

	write(text(spaced(s.Phrase.Input[:byteOffset])).
		X(x).Y(h / 2).Fg(green))
	write(text(spaced(s.Phrase.Input[byteOffset:])).
		X(x + runeOffset).Y(h / 2).Fg(black).Bg(red))

	seconds := time.Since(s.Phrase.StartedAt).Seconds()
	errorsText := text("%3d errors", s.Phrase.Errors).
		Y(h/2 + 4).Fg(termbox.ColorDefault)
	secondsText := text("%4.1f seconds", seconds).
		Y(h/2 + 4)

	write(errorsText.X(w/2 - 1).Align(Right))
	write(secondsText.X(w/2 + 1))
}

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

func spaced(s string) string {
	return strings.Replace(s, " ", "␣", -1)
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
