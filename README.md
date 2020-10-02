# Gokeybr

A fork of [gotypist](https://github.com/pb-/gotypist) touch-typing tutor, that tries to improve on [keybr.com](https://www.keybr.com/) by generating training lessons not only from words, but with curly brackets, or any other type of text you would want to train on.

Rewritten to use [tcell](https://github.com/gdamore/tcell/) instead of [termbox-go](https://github.com/nsf/termbox-go).

## Installation

### From source

```bash
go get github.com/bunyk/gokeybr
```

## Usage

`gokeybr --help` will give you the latest & most true information with which parameters this could be started.

## Code organization
Code is split in following packages:

- `cmd/` - is entry point of the program, handles parsing of arguments and starts app
- `app/` - contains code of event loop and overall logic of typing session
- `phrase/` - loading and generation of training texts
- `view/` - anyting related to displaying information on the screen
- `stats/` - keeping track of your progress & helping to generate most useful training session
