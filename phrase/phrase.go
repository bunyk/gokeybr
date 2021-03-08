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
	"unicode/utf8"

	"github.com/bunyk/gokeybr/fs"
)

func FromFile(filename string, offset, minLength int) (string, int, error) {
	items, skipped, err := readFileLines(filename, offset)
	if err != nil {
		return "", skipped, err
	}
	items = slice(items, minLength)
	return strings.Join(items, "\n"), skipped, nil
}

func Words(filename string, n int) (string, error) {
	words, _, err := readFileLines(filename, 0)
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

func readFileLines(filename string, offset int) (lines []string, skipped int, err error) {
	var data []byte
	if filename == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(filename)
		if offset < 0 {
			offset = lastFileOffset(filename)
			fmt.Printf("Offset was not given, loaded last saved progress on line %d\n", offset)
		}
	}
	if err != nil {
		return
	}

	skip := offset
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
		if skip > 0 {
			skip--
			skipped += utf8.RuneCountInString(line) + 1
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

func UpdateFileProgress(filename string, linesTyped, offset int) error {
	if filename == "-" { // Not saving for stdin
		return nil
	}
	if linesTyped < 1 {
		return nil // need to type at least line to update progress
	}
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
	if offset < 0 {
		progressTable[filename] += linesTyped
	} else {
		progressTable[filename] = offset + linesTyped
	}
	fmt.Printf("Saving progress for %s to be line #%d\n", filename, progressTable[filename])
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
