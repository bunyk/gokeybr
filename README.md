# Gokeybr

A fork of [gotypist](https://github.com/pb-/gotypist) touch-typing tutor, that tries to improve on [keybr.com](https://www.keybr.com/)

Work in progress, not expect it to work yet. :D

This project is done to improve my typing speed, and to figure out how to use [termbox-go](https://github.com/nsf/termbox-go), because I found no tutorials for that library or gocell.

## Installation

### From source

```bash
go get github.com/bunyk/gokeybr
```

## Usage

    gotypist [-f FILE] [-s] [-n PROB] [-c] [WORD]...

    WORD...     Explicitly specify a phrase
    -f FILE     Use FILE instead of /usr/share/dict/words as data source
    -n PROB     Sprinkle in random numbers with probability 0 <= PROB <= 1
    -c          Tread -f FILE as code and go sequenntially through the lines
    -d          Run in demo mode to take a screenshot

## Key bindings

    ESC   quit
    C-F   skip forward to the next phrase
    C-R   toggle repeat phrase mode

## Code organization

The code loosely follows an [Elm-like architecture](https://guide.elm-lang.org/architecture/). In a nutshell that means all interesting and Gotypist-specific code resides within pure functions. This is quite experimental and some corners were cut since Go is not primarily a functional language, but it still enjoys a lot of the benefits of this architectural style!
