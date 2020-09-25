package app

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/bunyk/gokeybr/models"
	"github.com/nsf/termbox-go"
)

type App struct {
	state State
}

func New(params models.Parameters) (*App, error) {
	a := &App{}

	a.state = InitState(params)
	if err := InitTermbox(); err != nil {
		return a, err
	}
	return a, nil
}

func InitTermbox() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	termbox.SetInputMode(termbox.InputEsc)
	// This is done to update rerender timer values, etc.
	go func() {
		for range time.Tick(100 * time.Millisecond) {
			termbox.Interrupt()
			// Interrupt an in-progress call to PollEvent
			// by causing it to return EventInterrupt.
		}
	}()
	return nil
}

func InitState(params models.Parameters) State {
	state := *NewState(0, DefaultPhrase)

	if len(params.Sourcetext) > 0 {
		state.PhraseGenerator = StaticPhrase(params.Sourcetext)
		state = resetPhrase(state, false)
	} else {
		data, err := readFile(params.Sourcefile)
		if err != nil {
			log.Fatal(err)
		}
		state = reduceDatasource(state, data, params.Codelines)
	}

	return state
}

func (a *App) Run() {
	for {
		render(a.state)
		a.state = reduceMessages(a.state, waitForEvent(), time.Now())
	}
}

func readFile(filename string) (content []byte, err error) {
	if filename == "-" {
		content, err = ioutil.ReadAll(os.Stdin)
	} else {
		content, err = ioutil.ReadFile(filename)
	}
	return
}

func reduceDatasource(state State, data []byte, codelines bool) State {
	var generator func([]string) PhraseFunc

	items := readLines(data)
	if codelines {
		items = filterWords(items, `^[^/][^/]`, 80)
		generator = SequentialLine
	} else {
		items = filterWords(items, `^[a-z]+$`, 8)
		generator = func(words []string) PhraseFunc { return RandomPhrase(words, PhraseLength) }
		state.Seed = time.Now().UnixNano()
	}

	if len(items) == 0 {
		log.Fatal("datafile contains no usable data")
	}

	state.PhraseGenerator = generator(items)

	return resetPhrase(state, false)
}

func runCommands(state State, commands []Command) State {
	for _, command := range commands {
		state = reduceMessages(state, RunCommand(command), time.Now())
	}

	return state
}

func reduceMessages(state State, messages []Message, now time.Time) State {
	for _, message := range messages {
		newState, commands := reduce(state, message, time.Now())
		state = runCommands(newState, commands)
	}

	return state
}

func waitForEvent() []Message {
	ev := termbox.PollEvent()
	switch ev.Type {
	case termbox.EventKey:
		return []Message{ev}
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventInterrupt:
	}

	return []Message{}
}
