package stats

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/bunyk/gokeybr/fs"
)

// TODO: maybe use integer values in miliseconds, to save space?
// Even microseconds will save 4 chars per datapoint

const MinSessionLength = 5

const LogStatsFile = "sessions_log.jsonl"
const StatsFile = "stats.json"

func SaveSession(start time.Time, text []rune, timeline []float64, training bool) error {
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
	if err := fs.AppendJSONLine(
		LogStatsFile,
		statLogEntry{
			Start:    start.Format(time.RFC3339),
			Text:     string(text),
			Timeline: timeline,
		},
	); err != nil {
		return err
	}
	return updateStats(text, timeline, training)
}

func GenerateTrainingSession(length int) (string, error) {
	stats, err := loadStats()
	if err != nil {
		return "", err
	}
	if length == 0 {
		length = 100
	}
	fmt.Println("Loaded stats, generating training sequence")
	return generateSequence(stats.trigramsToTrain(), length), nil
}

func updateStats(text []rune, timeline []float64, training bool) error {
	stats, err := loadStats()
	if err != nil {
		return err
	}
	stats.addSession(text, timeline, training)
	return fs.SaveJSON(StatsFile, stats)
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

// Score approximates time that will be spent typing this trigram
// It is frequency of trigram (it's count) multiplied by average duration of typing one
func (ts trigramStat) Score(avgDuration float64) float64 {
	duration := ts.Duration.Average(avgDuration)
	score := float64(ts.Count) * duration
	if score == 0 {
		score = 0.00001
	}
	return score
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
		chain[bigram][t[2]] = ts.Score * ts.Score
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
	for _, r := range trigrams[rand.Intn(len(trigrams)/10)].Trigram {
		text = append(text, r)
	}
	for len(text) < length {
		links := chain[string(text[len(text)-2:])]
		if len(links) == 0 {
			text = append(text, text[len(text)%3])
		}
		choice := rand.Float64()
		totalScore := 0.0
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

func (s *stats) addSession(text []rune, timeline []float64, training bool) {
	s.SessionsCount++
	s.TotalCharsTyped += len(text)
	s.TotalSessionsDuration += timeline[len(timeline)-1]
	for i := 0; i < len(text)-3; i++ {
		k := string(text[i : i+3])
		tr := s.Trigrams[k]
		if !training { // we do not count trigram frequencies in training sessions
			tr.Count++ // because that will make them stuck in training longer
		}
		tr.Duration.Append(timeline[i+3] - timeline[i])
		s.Trigrams[k] = tr
	}
}

func loadStats() (*stats, error) {
	var s stats
	err := fs.LoadJSON(StatsFile, &s)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: File %s does not exist! It will be created.\n", StatsFile)
			return &stats{Trigrams: make(map[string]trigramStat)}, nil
		}
		return nil, err
	}
	return &s, nil
}

type statLogEntry struct {
	Start    string    `json:"start"`
	Text     string    `json:"text"`
	Timeline []float64 `json:"timeline"`
}
