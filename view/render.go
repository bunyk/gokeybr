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

	write3colors(s, dd.DoneText, dd.WrongText, dd.TODOText, 2, 2, w-5, h-4)

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

func write3colors(scr tcell.Screen, done, wrong, todo []rune, x, y, w, h int) {
	var cursorX, cursorY int
	var style tcell.Style
	var blank bool // turns off printing for computing cursor position

	// put character on screen
	putC := func(r rune) {
		if blank {
			return // this is just trial run
		}
		scr.SetContent(cursorX, cursorY, r, nil, style)
	}
	putS := func(s []rune) {
		for _, c := range s {
			if !blank && cursorY > y+h {
				break // Do not type below allowed window
			}
			if !blank && cursorY == y+h {
				c = '↡' // If we are on a lower border - show that there will be more text
			}
			if c == '\n' {
				putC('⏎')
				// move cursor to new line
				cursorX = x
				cursorY++
				continue
			}
			// displayable spaces
			if c == ' ' {
				c = '␣'
			}
			putC(c)
			cursorX++
			if cursorX >= x+w { // line wrap
				cursorX = x
				cursorY++
			}
		}
	}

	cursorX = x
	cursorY = y
	blank = true

	putS(done)
	putS(wrong)

	// cursor will be in current position if we won't scroll

	// but we will scroll following number of lines
	scroll := cursorY - y - h/2

	// TODO: maybe move this out
	if scroll > 0 {
		scrolledLines := 0
		i := 0
		var c rune
		for i, c = range done {
			if c == '\n' {
				scrolledLines++
			}
			if scrolledLines >= scroll {
				i++
				break
			}
		}
		done = done[i:]
		if len(done) == 0 && scrolledLines < scroll {
			for i, c = range wrong {
				if c == '\n' {
					scrolledLines++
				}
				if scrolledLines >= scroll {
					i++
					break
				}
			}
			wrong = wrong[i:]
		}
	}

	cursorX = x
	cursorY = y
	blank = false

	style = doneStyle
	putS(done)

	style = errorStyle
	putS(wrong)

	scr.ShowCursor(cursorX, cursorY)

	style = tcell.StyleDefault
	putS(todo)
}
