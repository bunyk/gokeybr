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

	"github.com/bunyk/gokeybr/stats"
)

// TODO: reorganize this to better fit single session per single run
// Maybe limit/offset for typing and report about offset in the end?
// Or save offset in stats...

// Will return tex to train on,
// and boolean that will be true if that text is randomly generated and not a real text
func FetchPhrase(filename, sourcetext, kind string, maxLength int) (string, bool, error) {
	var items []string
	var err error
	if kind == "stats" {
		sourcetext, err = stats.GenerateTrainingSession(maxLength)
		if err != nil {
			return "", false, err
		}
		return sourcetext, true, nil
	}
	if len(sourcetext) > 0 {
		items = strings.Split(sourcetext, "\n")
	} else if len(filename) > 0 {
		items, err = readFileLines(filename)
		if err != nil {
			return "", false, err
		}
	} else {
		items = []string{"the quick brown fox jumps over the lazy dog"}
	}
	if kind == "paragraphs" {
		items = slice(items, maxLength)
		return strings.Join(items, "\n"), false, nil
	} else if kind == "words" {
		return randomWords(items, maxLength), false, nil
	}
	return "", false, fmt.Errorf("Unknown text type: %s (allowed: paragraphs, words, stats)", kind)
}

func randomWords(words []string, minLength int) string {
	var phrase []string
	l := -1
	for l < minLength {
		w := words[rand.Intn(len(words))]
		phrase = append(phrase, w)
		l += 1 + len([]rune(w))
	}
	return strings.Join(phrase, " ")
}

type sequentialLineGenerator struct {
	Lines       []string
	CurrentLine int
	isTraining  bool
}

func (slg *sequentialLineGenerator) Phrase() string {
	cl := slg.CurrentLine
	slg.CurrentLine = (cl + 1) % len(slg.Lines)
	return slg.Lines[cl]
}

func (slg sequentialLineGenerator) IsTraining() bool {
	return slg.isTraining
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

func slice(lines []string, maxLength int) []string {
	res := make([]string, 0)
	totalLen := 0
	for _, l := range lines {
		l = strings.TrimSpace(l)
		res = append(res, l)
		chars := len([]rune(l))
		totalLen += chars + 1
		if totalLen >= maxLength {
			break
		}
	}
	return res
}
