package phrase

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Generator generates a phrase
type Generator interface {
	Phrase() string
}

func NewGenerator(filename, sourcetext, kind string, maxLength int) (Generator, error) {
	var items []string
	var err error
	if len(sourcetext) > 0 {
		items = strings.Split(sourcetext, "\n")
	} else if len(filename) > 0 {
		items, err = readFileLines(filename)
		if err != nil {
			return nil, err
		}
	} else {
		items = []string{"the quick brown fox jumps over the lazy dog"}
	}
	if kind == "paragraphs" {
		items = makeParagraphs(items)
		return &sequentialLineGenerator{Lines: items}, nil
	} else if kind == "words" {
		return newRandomGenerator(items, maxLength), nil
	}
	return nil, fmt.Errorf("Unknown text type: %s (allowed: paragraphs, random)", kind)
}

// randomGenerator composes a random phrase with given length from given words.
type randomGenerator struct {
	Words     []string
	MinLength int
	seed      int64
}

func newRandomGenerator(words []string, minLength int) *randomGenerator {
	return &randomGenerator{
		Words:     words,
		MinLength: minLength,
		seed:      time.Now().UnixNano(),
	}
}

func (rg *randomGenerator) Phrase() string {
	rand := rand.New(rand.NewSource(rg.seed))
	var phrase []string
	l := -1
	for l < rg.MinLength {
		w := rg.Words[rand.Int31n(int32(len(rg.Words)))]
		phrase = append(phrase, w)
		l += 1 + len(w)
	}
	rg.seed = rand.Int63()
	return strings.Join(phrase, " ")
}

type sequentialLineGenerator struct {
	Lines       []string
	CurrentLine int
}

func (slg *sequentialLineGenerator) Phrase() string {
	cl := slg.CurrentLine
	slg.CurrentLine = (cl + 1) % len(slg.Lines)
	return slg.Lines[cl]
}

func readFileLines(filename string) (lines []string, err error) {
	var data []byte
	if filename == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		return
	}

	reader := bufio.NewReader(bytes.NewBuffer(data))
	for {
		line, rerr := reader.ReadString('\n')
		if rerr != nil {
			if rerr == io.EOF {
				if len(lines) == 0 {
					err = errors.New("datafile contains no usable data")
				}
				return
			}
			err = rerr
			return
		}
		lines = append(lines, line[:len(line)-1])
	}

}

func makeParagraphs(lines []string) []string {
	res := make([]string, 0)
	buf := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			if len(buf) > 0 {
				res = append(res, strings.TrimSpace(buf))
				buf = ""
			}
		} else {
			buf += "\n" + l
		}
	}
	if len(buf) > 0 {
		res = append(res, strings.TrimSpace(buf))
	}
	return res
}
