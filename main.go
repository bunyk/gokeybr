package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	loop()
}

func loop() {
	state, cmds := Init()
	state = runCommands(state, cmds)

	for {
		render(state)
		state = reduceMessages(state, waitForEvent(), time.Now())
	}
}

func Init() (State, []Command) {
	args := os.Args
	state := *NewState(0, DefaultPhrase)

	commandLine := flag.NewFlagSet(args[0], flag.ContinueOnError)
	datafile := commandLine.String("f", "/usr/share/dict/words", "load word list from `FILE`. \"-\" for stdin.")
	commandLine.BoolVar(&state.Codelines, "c", false, "treat -f FILE as lines of code")

	err := commandLine.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			buf := new(bytes.Buffer)
			commandLine.SetOutput(buf)
			commandLine.PrintDefaults()
			log.Fatal(buf.String())
			return State{}, []Command{}
		}
		log.Fatal(err.Error())
		return State{}, []Command{}
	}

	go func() {
		for range time.Tick(100 * time.Millisecond) {
			termbox.Interrupt()
			// Interrupt an in-progress call to PollEvent
			// by causing it to return EventInterrupt.
		}
	}()

	commands := []Command{}

	if len(commandLine.Args()) > 0 {
		state.PhraseGenerator = StaticPhrase(strings.Join(commandLine.Args(), " "))
		state = resetPhrase(state, false)
	} else {
		data, err := readFile(*datafile)
		if err != nil {
			log.Fatal(err)
		}
		state = reduceDatasource(state, data)
	}

	return state, commands
}

func readFile(filename string) (content []byte, err error) {
	if filename == "-" {
		content, err = ioutil.ReadAll(os.Stdin)
	} else {
		content, err = ioutil.ReadFile(filename)
	}
	return
}

func reduceDatasource(state State, data []byte) State {
	var generator func([]string) PhraseFunc

	items := readLines(data)
	if state.Codelines {
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
