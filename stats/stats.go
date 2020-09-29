package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// TODO: maybe use integer values in miliseconds, to save space?
// Even microseconds will save 4 chars per datapoint

const MinSessionLength = 5

const LogStatsFile = ".gokeybr_stats_log.jsonl"
const StatsFile = ".gokeybr_stats.json"
const FileAccess = 0644

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
	return updateStats(text, timeline)
}

func updateStats(text []rune, timeline []float64) error {
	filename := statFilePath(StatsFile)
	stats, err := loadStats(filename)
	if err != nil {
		return err
	}
	stats.addSession(text, timeline)
	return saveStats(filename, stats)
}

type stats struct {
	TotalCharsTyped       int
	TotalSessionsDuration float64
	SessionsCount         int
	Trigrams              map[string]trigramStat
}

type trigramStat struct {
	Count    int    `json:"c"`
	Duration Window `json:"d"`
}

func newStats() *stats {
	return &stats{
		Trigrams: make(map[string]trigramStat),
	}
}

func (s *stats) addSession(text []rune, timeline []float64) {
	s.SessionsCount++
	s.TotalCharsTyped += len(text)
	s.TotalSessionsDuration += timeline[len(timeline)-1]
	for i := 0; i < len(text)-3; i++ {
		k := string(text[i : i+3])
		tr := s.Trigrams[k]
		tr.Count++
		tr.Duration.Append(timeline[i+3] - timeline[i])
		s.Trigrams[k] = tr
	}
}

func saveStats(filename string, s *stats) error {
	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, FileAccess)
}

func loadStats(filename string) (*stats, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: File %s not exist! It will be created.\n", filename)
			return newStats(), nil
		}
		return nil, err
	}
	var s stats
	return &s, json.Unmarshal(data, &s)
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
		statFilePath(LogStatsFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, FileAccess,
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
