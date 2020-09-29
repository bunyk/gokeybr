package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const MinSessionLength = 5

func SaveSession(start time.Time, text []rune, timeline []float64) error {
	if len(text) != len(timeline) {
		return fmt.Errorf(
			"Length of text (%d) does not match leght of timeline (%d)! Stats not saved.",
			len(text), len(timeline),
		)
	}
	if len(text) < MinSessionLength {
		fmt.Printf("Not updating stats for session only %d characters long\n", len(text))
		return nil
	}
	if err := saveStatLogEntry(statLogEntry{
		Start:    start.Format(time.RFC3339),
		Text:     string(text),
		Timeline: timeline,
	}); err != nil {
		return err
	}
	return nil
}

type stats struct {
	charCounts       map[rune]int // Count characters
	trigramCounts    map[string]int
	trigramDurations map[string]Window
}

func newStats() *stats {
	return &stats{
		charCounts:       make(map[rune]int),
		trigramCounts:    make(map[string]int),
		trigramDurations: make(map[string]Window),
	}
}

func (s *stats) addSession(text []rune, timeline []float64) {
	for _, r := range text {
		s.charCounts[r]++
	}
}

type statLogEntry struct {
	Start    string    `json:"start"`
	Text     string    `json:"text"`
	Timeline []float64 `json:"timeline"`
}

func statFilePath(name string) string {
	return filepath.Join(
		os.Getenv("HOME"),
		name,
	)
}

func saveStatLogEntry(e statLogEntry) error {
	f, err := os.OpenFile(
		statFilePath(".gokeybr_stats_log.jsonl"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644,
	)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f, string(data))
	return err
}
