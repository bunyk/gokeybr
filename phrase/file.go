package phrase

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func ReadFileLines(filename string) (lines []string, err error) {
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
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return lines, nil
			}
			return lines, err
		}
		lines = append(lines, line[:len(line)-1])
	}
}

func FilterWords(words []string, pattern string, maxLength int) []string {
	filtered := make([]string, 0)
	compiled := regexp.MustCompile(pattern)

	for _, word := range words {
		trimmed := strings.TrimSpace(word)
		if compiled.MatchString(trimmed) && len(trimmed) <= maxLength {
			filtered = append(filtered, trimmed)
		}
	}

	return filtered
}
