// contains all commands (side effects)
package app

import (
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

type Message interface{}

func Exit(status int, message string) {
	termbox.Close()

	if message != "" {
		fmt.Println(message)
	}

	os.Exit(status)
}
