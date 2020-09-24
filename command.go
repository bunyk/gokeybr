// contains all commands (side effects)
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

type Message interface{}

type Command interface{}

var Noop = []Command{}

type Interrupt struct {
	Delay time.Duration
}

type Exit struct {
	Status         int
	GoodbyeMessage string
}

func RunCommand(cmd Command) []Message {
	switch c := cmd.(type) {
	case Interrupt:
		return interrupt(c.Delay)
	case Exit:
		return exit(c.Status, c.GoodbyeMessage)
	}

	exit(1, fmt.Sprintf("Cannot handle command of type %T", cmd))
	return noMessages
}

var noMessages = []Message{}

func interrupt(d time.Duration) []Message {
	time.AfterFunc(d, termbox.Interrupt)
	return noMessages
}

func exit(status int, message string) []Message {
	termbox.Close()

	if message != "" {
		fmt.Println(message)
	}

	os.Exit(status)

	return noMessages
}
