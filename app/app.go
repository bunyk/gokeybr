package app

import (
	"fmt"
	"time"

	"github.com/bunyk/gokeybr/fs"
	"github.com/bunyk/gokeybr/view"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

// used for testing
// j - type, k - untype
const cheating = false

const InitialLife = 10 * time.Second

// App holds whole app state
type App struct {
	Text          []rune
	Timeline      []float64
	InputPosition int
	ErrorInput    []rune
	StartedAt     time.Time
	Offset        int

	Zen  bool
	Mute bool

	// Minimal permitted speed
	MinSpeed              int
	LastLifeReductionTime time.Time
	// For how long you could have your speed under speed limit and still continue typing
	RemainingLife time.Duration

	scr tcell.Screen
}

func New(text string) (*App, error) {
	a := &App{}
	a.ErrorInput = make([]rune, 0, 20)
	a.Text = []rune(text)
	a.Timeline = make([]float64, len(a.Text))
	a.RemainingLife = InitialLife

	encoding.Register()
	var err error
	if a.scr, err = tcell.NewScreen(); err != nil {
		return a, err
	}
	if err = a.scr.Init(); err != nil {
		return a, err
	}
	return a, nil
}

// tick will implement tcell.Event, and be used for updating timers on screen
type tick struct {
}

func (t tick) When() time.Time {
	return time.Time{} // no need to know real time yet
}

func (a *App) Run() error {
	defer a.scr.Fini()
	events := make(chan tcell.Event)
	go func() {
		for {
			ev := a.scr.PollEvent()
			events <- ev
		}
	}()
	if !a.Zen {
		go func() {
			t := time.NewTicker(100 * time.Millisecond)
			for {
				<-t.C
				events <- tick{}
			}
		}()
	}

	for {
		view.Render(a.scr, a.ToDisplay())
		if a.RemainingLife <= 0 {
			return nil
		}
		ev := <-events
		switch event := ev.(type) {
		case *tcell.EventKey:
			if !a.processKey(event) {
				if cheating {
					a.InputPosition = 0
				}
				return nil
			}
		case *tcell.EventResize:
			a.scr.Sync()
		}
	}
}

func log(v interface{}) {
	fs.AppendJSONLine("debug.jsonl", v)
}

// wordsPerChar is used for computing WPM.
// Word is considered to be in average 5 characters long.
const wordsPerChar = 0.2
const WPMWindow = 20

func (a *App) CheckWPM() float64 {
	wpm := 0.0
	seconds := time.Since(a.StartedAt).Seconds()
	if a.InputPosition > 1 {
		secondsPerWindow := seconds - a.Timeline[max(a.InputPosition-WPMWindow, 0)]
		wpm = wordsPerChar * float64(min(WPMWindow, a.InputPosition)) / secondsPerWindow * 60.0

		if a.MinSpeed > 0 { // need to check speed limits
			if wpm < float64(a.MinSpeed) { // speed below limit
				if !a.LastLifeReductionTime.IsZero() { // speed was already below limit
					diff := time.Since(a.LastLifeReductionTime)
					a.RemainingLife -= diff
				}
				a.LastLifeReductionTime = time.Now()
			} else { // speed above limit, stop reductions
				a.LastLifeReductionTime = time.Time{}
			}
		}
	}
	return wpm
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *App) ToDisplay() view.DisplayableData {
	wpm := a.CheckWPM()
	life := 0.0
	if a.MinSpeed > 0 {
		life = float64(a.RemainingLife) / float64(InitialLife)
	}
	return view.DisplayableData{
		DoneText:  a.Text[:a.InputPosition],
		WrongText: a.ErrorInput,
		TODOText:  a.Text[a.InputPosition:],
		StartedAt: a.StartedAt,
		WPM:       wpm,
		Life:      life,
		Zen:       a.Zen,
		Offset:    a.Offset,
	}
}

func (a App) Summary() string {
	if a.InputPosition == 0 {
		return "Typed nothing"
	}
	elapsed := a.Timeline[a.InputPosition-1]
	if elapsed == 0 {
		return "Speed of light! (actually, probably some error with timer)"
	}
	return fmt.Sprintf(
		"Typed %d characters in %4.1f seconds. Speed: %4.1f wpm\n",
		a.InputPosition, elapsed, float64(a.InputPosition)/elapsed*60.0/5.0,
	)
}

// Compute number of typed lines
func (a App) LinesTyped() int {
	lt := 0
	for _, c := range a.Text[:a.InputPosition] {
		if c == '\n' {
			lt++
		}
	}
	if a.InputPosition == len(a.Text) {
		lt++
	}
	return lt
}

// Return true when should continue loop
func (a *App) processKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		return false
	}

	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		a.processBackspace()
	default:
		return a.processCharInput(ev)
	}
	return true
}

func (a *App) processBackspace() {
	if len(a.ErrorInput) == 0 {
		return
	}
	a.ErrorInput = a.ErrorInput[:len(a.ErrorInput)-1]
}

// Return true when should continue loop
func (a *App) processCharInput(ev *tcell.EventKey) bool {
	var ch rune
	if ev.Key() == tcell.KeyRune {
		ch = ev.Rune()
	} else if ev.Key() == tcell.KeyEnter || ev.Key() == tcell.KeyCtrlJ {
		ch = '\n'
	}
	if ch == 0 {
		return true
	}
	if a.StartedAt.IsZero() {
		a.StartedAt = ev.When()
	}

	if cheating { // always type correct :)
		if ch == 'j' {
			a.InputPosition += 3
		}
		if ch == 'k' {
			a.InputPosition -= 3
		}
		if a.InputPosition < 0 {
			a.InputPosition = 0
		}
		return a.InputPosition < len(a.Text)
	}
	if ch == a.Text[a.InputPosition] && len(a.ErrorInput) == 0 { // correct
		a.Timeline[a.InputPosition] = ev.When().Sub(a.StartedAt).Seconds()
		a.InputPosition++
	} else { // wrong
		a.ErrorInput = append(a.ErrorInput, ch)
		if !a.Mute {
			a.scr.Beep()
		}
	}
	return a.InputPosition < len(a.Text)
}
