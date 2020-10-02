package view

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

const wordsPerChar = 0.2 // In computing WPM word is considered to be in avearge 5 characters long

var doneStyle tcell.Style = tcell.StyleDefault.
	Foreground(tcell.ColorGreen)
var errorStyle = tcell.StyleDefault.
	Background(tcell.ColorRed).
	Foreground(tcell.ColorBlack)

type DisplayableData struct {
	Header    string
	DoneText  []rune
	WrongText []rune
	TODOText  []rune
	StartedAt time.Time
}

func Render(s tcell.Screen, dd DisplayableData) {
	s.Clear()
	w, h := s.Size()

	write(s, dd.Header, 1, 0, tcell.StyleDefault)

	write3colors(s, dd.DoneText, dd.WrongText, dd.TODOText, 2, 2, w-5)

	// Stats:
	seconds := 0.0
	wpm := 0.0
	if !dd.StartedAt.IsZero() {
		seconds = time.Since(dd.StartedAt).Seconds()
		wpm = wordsPerChar * float64(len(dd.DoneText)) / seconds * 60.0
	}
	stats := fmt.Sprintf("%.1f sec, %.1f wpm", seconds, wpm)
	x := (w - utf8.RuneCountInString(stats)) / 2
	write(s, stats, x, h-1, tcell.StyleDefault)
	s.Show()
}

func write(scr tcell.Screen, text string, x, y int, style tcell.Style) {
	for _, c := range text {
		scr.SetContent(x, y, c, nil, style)
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
