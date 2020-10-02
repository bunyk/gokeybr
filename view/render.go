package view

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/nsf/termbox-go"
)

const wordsPerChar = 0.2 // In computing WPM word is considered to be in avearge 5 characters long

var doneStyle tcell.Style = tcell.StyleDefault.
	Background(tcell.ColorBlack).
	Foreground(tcell.ColorGreen)
var errorStyle = tcell.StyleDefault.
	Background(tcell.ColorRed).
	Foreground(tcell.ColorBlack)

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

func Render(s tcell.Screen, dd DisplayableData) {
	s.Clear()
	w, _ := s.Size()

	// write(text(dd.Header).X(1).Y(0).Bg(white).Fg(black))

	write3colors(s, dd.DoneText, dd.WrongText, dd.TODOText, 2, 2, w-5)

	// Stats:
	/*
		seconds := 0.0
		wpm := 0.0
		if !dd.StartedAt.IsZero() {
			seconds = time.Since(dd.StartedAt).Seconds()
			wpm = wordsPerChar * float64(len(dd.DoneText)) / seconds * 60.0
		}
		write(text("%4.1f sec, %4.1f wpm", seconds, wpm).X(w/2 + 1).Y(h - 1).Align(Center))
	*/
	s.Show()
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

func write3colors(scr tcell.Screen, done, wrong, todo []rune, x, y, w int) {
	cursorX := x
	cursorY := y
	putS := func(s []rune, style tcell.Style) {
		for _, c := range s {
			if c == '\n' {
				scr.SetContent(cursorX, cursorY, '⏎', nil, style)
				cursorX = x
				cursorY++
				continue
			}
			if c == ' ' {
				c = '␣'
			}
			scr.SetContent(cursorX, cursorY, c, nil, style)
			cursorX++
			if cursorX >= x+w {
				cursorX = x
				cursorY++
			}
		}
	}

	putS(done, doneStyle)
	putS(wrong, errorStyle)
	scr.ShowCursor(cursorX, cursorY)
	putS(todo, tcell.StyleDefault)
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
