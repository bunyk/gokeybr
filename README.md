# Gokeybr
Minimalistic touch-typing training program, that tries to improve on [keybr.com](https://www.keybr.com/) by generating training lessons not only from words, but with symbols from code, or any other type of text you would want to train on.

You could, for example, train to type ten `if err != nil { return err }` per minute and become fastest Go coder.

![Screenshot of a Gokeybr session](screenshot.png)

On the screenshot you see `gokeybr` running in `stats` mode, where it generates training session based on your typing stats. In this case it mixes code with "words", based on frequency and typing speed of character sequences in texts that were used for other training sessions.


## Installation

```bash
go get github.com/bunyk/gokeybr
```

## Usage

`gokeybr --help` will give you the latest & most true information with which parameters this could be started.

## Code
A fork of [gotypist](https://github.com/pb-/gotypist), rewritten to use [tcell](https://github.com/gdamore/tcell/) instead of [termbox-go](https://github.com/nsf/termbox-go). Also added support for multiline typing sessions and statistically generated exercises. Removed modes, so in each session you could strive for any result you wish.

Acrhitecture is changed from Elm-like to more classical. Code is split in following packages:

- `cmd/` - is entry point of the program, handles parsing of arguments and starts app
- `app/` - contains code of event loop and overall logic of typing session
- `phrase/` - loading and generation of training texts
- `view/` - anyting related to displaying information on the screen
- `stats/` - keeping track of your progress & helping to generate most useful training session
- `fs/` - utilities to work with filesystem storage
