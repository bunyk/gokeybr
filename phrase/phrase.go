package phrase

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bunyk/gokeybr/fs"
	"github.com/bunyk/gokeybr/stats"
)

// Will return text to train on,
// and boolean that will be true if that text is randomly generated and not a real text
func FetchPhrase(filename, kind string, minLength int, offset int) (string, bool, error) {
	var items []string
	var err error
	if kind == "stats" {
		sourcetext, err := stats.GenerateTrainingSession(minLength)
		if err != nil {
			return "", false, err
		}
		return sourcetext, true, nil
	}
	if len(filename) > 0 {
		items, err = readFileLines(filename, offset)
		if err != nil {
			return "", false, err
		}
	} else {
		items = []string{"the quick brown fox jumps over the lazy dog"}
	}
	if kind == "lines" {
		items = slice(items, minLength)
		return strings.Join(items, "\n"), false, nil
	}
	return "", false, fmt.Errorf("Unknown text type: %s (allowed: lines, stats)", kind)
}

func Words(filename string, n int) (string, error) {
	words, err := readFileLines(filename, 0)
	if err != nil {
		return "", err
	}
	rand.Seed(time.Now().UTC().UnixNano())
	var phrase []string
	for i := 0; i < n; i++ {
		w := words[rand.Intn(len(words))]
		phrase = append(phrase, w)
	}
	return strings.Join(phrase, " "), nil
}

func readFileLines(filename string, offset int) (lines []string, err error) {
	var data []byte
	if filename == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(filename)
		if offset < 0 {
			offset = lastFileOffset(filename)
		}
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
					err = fmt.Errorf("datafile %s contains no usable data at offset %d", filename, offset)
				}
				return
			}
			err = rerr
			return
		}
		if offset > 0 {
			offset--
		} else {
			lines = append(lines, line[:len(line)-1])
		}
	}

}

func slice(lines []string, minLength int) []string {
	res := make([]string, 0)
	totalLen := 0
	for _, l := range lines {
		l = strings.TrimSpace(l)
		res = append(res, l)
		chars := len([]rune(l))
		totalLen += chars + 1
		if minLength > 0 && totalLen >= minLength {
			break
		}
	}
	return res
}

const ProgressFile = "progress.json"

func UpdateFileProgress(filename string, linesTyped int) error {
	var progressTable map[string]int
	if err := fs.LoadJSON(ProgressFile, &progressTable); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s is not found, will be created\n", ProgressFile)
			progressTable = make(map[string]int)
		} else {
			return err
		}
	}
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	progressTable[filename] += linesTyped
	return fs.SaveJSON(ProgressFile, progressTable)
}

func lastFileOffset(filename string) int {
	var progressTable map[string]int
	if err := fs.LoadJSON(ProgressFile, &progressTable); err != nil {
		fmt.Println(err)
		return 0
	}
	filename, err := filepath.Abs(filename)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return progressTable[filename]
}
