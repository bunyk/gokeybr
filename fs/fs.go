package fs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const FileAccess = 0644

func homeFilePath(name string) string {
	return filepath.Join(
		os.Getenv("HOME"),
		".gokeybr",
		name,
	)
}
func mkdir() {
	dir := homeFilePath("/")
	if _, err := os.Stat(dir); err != nil {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
}

func SaveJSON(filename string, o interface{}) error {
	data, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return err
	}
	mkdir()
	return ioutil.WriteFile(homeFilePath(filename), data, FileAccess)
}

func LoadJSON(filename string, v interface{}) error {
	mkdir()
	data, err := ioutil.ReadFile(homeFilePath(filename))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func AppendJSONLine(filename string, v interface{}) error {
	mkdir()
	f, err := os.OpenFile(homeFilePath(filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, FileAccess)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f, string(data))
	return err
}

type JSONLinesIterator struct {
	scanner *bufio.Scanner
	file    *os.File
}

func NewJSONLinesIterator(filename string) (*JSONLinesIterator, error) {
	file, err := os.Open(homeFilePath(filename))
	if err != nil {
		return nil, err
	}
	return &JSONLinesIterator{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (i JSONLinesIterator) Close() {
	i.file.Close()
}

func (i JSONLinesIterator) UnmarshalNextLine(v interface{}) (bool, error) {
	if !i.scanner.Scan() {
		return false, i.scanner.Err()
	}

	return true, json.Unmarshal(i.scanner.Bytes(), v)
}
