package view

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const wordsPerChar = 0.2 // In computing WPM word is considered to be in avearge 5 characters long

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

type DisplayableData struct {
	Header    string
	DoneText  []rune
	WrongText []rune
	TODOText  []rune
	StartedAt time.Time
}

func Render(dd DisplayableData) {
	_ = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	write(text(dd.Header).X(1).Y(0).Bg(white).Fg(black))

	write3colors(dd.DoneText, dd.WrongText, dd.TODOText, 2, 2, w-5)

	// Stats:
	seconds := 0.0
	wpm := 0.0
	if !dd.StartedAt.IsZero() {
		seconds = time.Since(dd.StartedAt).Seconds()
		wpm = wordsPerChar * float64(len(dd.DoneText)) / seconds * 60.0
	}
	write(text("%4.1f sec, %4.1f wpm", seconds, wpm).X(w/2 + 1).Y(h - 1).Align(Center))
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

func write3colors(done, wrong, todo []rune, x, y, w int) {
	cursorX := x
	cursorY := y
	putS := func(s []rune, fg, bg termbox.Attribute) {
		for _, c := range s {
			if c == '\n' {
				termbox.SetCell(cursorX, cursorY, '⏎', fg, bg)
				cursorX = x
				cursorY++
				continue
			}
			if c == ' ' {
				c = '␣'
			}
			termbox.SetCell(cursorX, cursorY, c, fg, bg)
			cursorX++
			if cursorX >= x+w {
				cursorX = x
				cursorY++
			}
		}
	}

	putS(done, green, 0)
	putS(wrong, black, red)
	termbox.SetCursor(cursorX, cursorY)
	putS(todo, white, 0)
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
