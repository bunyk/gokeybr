package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
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

func GenerateTrainingSession(length int) (string, error) {
	filename := statFilePath(StatsFile)
	stats, err := loadStats(filename)
	if err != nil {
		return "", err
	}
	return generateSequence(stats.trigramsToTrain(), length), nil
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

func (s stats) AverageCharDuration() float64 {
	return s.TotalSessionsDuration / float64(s.TotalCharsTyped)
}

type trigramStat struct {
	Count    int    `json:"c"`
	Duration Window `json:"d"`
}

func (ts trigramStat) Score(averateDuration float64) float64 {
	return float64(ts.Count) * ts.Duration.Average(averateDuration)
}

func newStats() *stats {
	return &stats{
		Trigrams: make(map[string]trigramStat),
	}
}

type TrigramScore struct {
	Trigram string
	Score   float64
}

// return list of trigrams with their relative importance to train
// the more frequent is trigram and the more long it takes to type it
// the more important will it be to train it
func (s stats) trigramsToTrain() []TrigramScore {
	totalScore := 0.0
	res := make([]TrigramScore, 0, len(s.Trigrams))
	for t, ts := range s.Trigrams {
		sc := ts.Score(s.AverageCharDuration() * 3)
		res = append(res, TrigramScore{
			Trigram: t,
			Score:   sc,
		})
		totalScore += sc
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})
	return res
}

type markovChain map[string]map[rune]float64

func generateSequence(trigrams []TrigramScore, length int) string {
	chain := make(markovChain)
	// build Markov chain
	for _, ts := range trigrams {
		t := []rune(ts.Trigram)
		bigram := string(t[:2])
		if chain[bigram] == nil {
			chain[bigram] = make(map[rune]float64)
		}
		chain[bigram][t[2]] = ts.Score
	}
	// normalize Markov chain
	for _, links := range chain {
		totalScore := 0.0
		for _, ls := range links {
			totalScore += ls
		}
		for k, ls := range links {
			links[k] = ls / totalScore
		}
	}
	text := make([]rune, 0, length)
	for _, r := range trigrams[0].Trigram {
		text = append(text, r)
	}
	for len(text) < length {
		fmt.Println(string(text))
		links := chain[string(text[len(text)-2:len(text)])]
		choice := rand.Float64()
		totalScore := 0.00001
		for r, sc := range links {
			totalScore += sc
			if choice <= totalScore {
				text = append(text, r)
				break
			}
		}
	}
	return string(text)
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
