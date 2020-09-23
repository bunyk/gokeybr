package main

import (
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	if len(os.Args) == 2 && os.Args[1] == "-d" {
		runDemo()
	} else {
		loop(os.Args)
	}
}

func loop(args []string) {
	state, cmds := Init(args)
	state = runCommands(state, cmds)

	for {
		render(state, time.Now())
		state = reduceMessages(state, waitForEvent(), time.Now())
	}
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
