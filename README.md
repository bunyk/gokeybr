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

`gokeybr --help` will give you the latest & most true information.

## Code organization
Code is split in following packages:

- `cmd/` - is entry point of the program, handles parsing of arguments and starts app
- `app/` - contains code of event loop and overall logic of app
- `phrase/` - loading and generation of training texts
- `view/` - anyting related to displaying information on the screen
