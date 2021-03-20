package view

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/bunyk/gokeybr/stats"
	"github.com/gdamore/tcell/v2"
)

// wordsPerChar is used for computing WPM.
// Word is considered to be in average 5 characters long.
const wordsPerChar = 0.2

var doneStyle tcell.Style = tcell.StyleDefault.
	Foreground(tcell.ColorGreen)

var redBar = tcell.StyleDefault.
	Background(tcell.ColorRed)

var greenBar = tcell.StyleDefault.
	Background(tcell.ColorGreen)

var errorStyle = redBar.
	Foreground(tcell.ColorBlack)

type DisplayableData struct {
	Header    string
	DoneText  []rune
	WrongText []rune
	TODOText  []rune
	Timeline  []float64
	StartedAt time.Time
	Zen       bool
	Offset    int
}

func Render(s tcell.Screen, dd DisplayableData) {
	s.Clear()
	w, h := s.Size()

	write3colors(s, dd.DoneText, dd.WrongText, dd.TODOText, 2, 2, w-5, h-4)

	if !dd.Zen {
		write(s, dd.Header, 1, 0, tcell.StyleDefault)

		// Stats:
		timer := "Go!"
		wpm := 0.0
		if !dd.StartedAt.IsZero() {
			seconds := time.Since(dd.StartedAt).Seconds()
			timer = fmt.Sprintf("%.1f sec", seconds)
			if len(dd.DoneText) > 10 {
				tenCharsSeconds := seconds - dd.Timeline[len(dd.DoneText)-10]
				wpm = wordsPerChar * 10 / tenCharsSeconds * 60.0
			}
		}

		// Show timer
		x := (w - utf8.RuneCountInString(timer)) / 2
		write(s, timer, x, h-1, tcell.StyleDefault)

		// Show wpm
		if wpm > 0 {
			speedometer := wpm / stats.AverageWPM() // compute speed improvement relative to average
			speedStyle := redBar                    // show slow speeds in red
			if speedometer >= 0.90 {                // Keeping in range of 90% of average speed is good
				speedStyle = greenBar
			}
			speedometer = speedometer / 2.0 // so average speed is displayed at the middle of speedometer
			if speedometer > 1.0 {
				speedometer = 1.0
			}
			vBar(s, 0, 0, int(float64(h)*speedometer), speedStyle)

			write(s, fmt.Sprintf("%.0f wpm", wpm), 0, h-1, tcell.StyleDefault)
		}

		// Show progress
		done := float64(len(dd.DoneText)) + float64(dd.Offset)
		progress := done / (done + float64(len(dd.TODOText)+len(dd.WrongText)))
		vBar(s, w-1, 0, int(float64(h)*progress), greenBar)
		progressIndicator := fmt.Sprintf("%.1f%%", progress*100)
		x = w - utf8.RuneCountInString(progressIndicator)
		write(s, progressIndicator, x, h-1, tcell.StyleDefault)
	}
	s.Show()
}

func vBar(scr tcell.Screen, x, y, h int, style tcell.Style) {
	for i := 0; i < h; i++ {
		scr.SetContent(x, y+i, ' ', nil, style)
	}
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
