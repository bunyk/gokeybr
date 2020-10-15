package stats

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

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

func RandomTraining(length int) (string, error) {
	trigrams, err := getTrigrams()
	if err != nil {
		return "", err
	}
	if length == 0 {
		length = 100
	}
	return markovSequence(trigrams, length), nil
}

func getTrigrams() ([]TrigramScore, error) {
	stats, err := loadStats()
	if err != nil {
		return nil, err
	}
	fmt.Println("Loaded stats, generating training sequence")

	trigrams := stats.trigramsToTrain()
	if len(trigrams) < NWeakest {
		return nil, fmt.Errorf("Not enought stats yet to generate good exercise")
	}
	return trigrams, err
}

func WeakestTraining(length int) (string, error) {
	if length == 0 {
		length = 100
	}
	trigrams, err := getTrigrams()
	if err != nil {
		return "", err
	}
	if length == 0 {
		length = 100
	}
	return weakestSequence(trigrams, length), nil
}

// Typing speed we think is unreachable
const speedOfLight = 200.0 // wpm

// wpm * 5 chars per word / 60 seconds in minute / 3 chars in trigram =
// wpm / 36
const trigramsPerSecSq = speedOfLight * speedOfLight / 1296.0

func effortResult(trigramTime float64) float64 {
	speed := 3.0 / trigramTime
	q := speed * speed / trigramsPerSecSq
	if q > 1 {
		return 0
	}
	return math.Sqrt(1 - q)
}

func weakestSequence(trigrams []TrigramScore, length int) string {
	// First, we start from the weakest trigram, say abc
	// Easiest - we would just repeat it, like abcabcabc..., but
	// maybe bca is already trained good enough. So we threat each
	// trigram abc as graph edge ab -> bc, with the weight = 1 / score of trigram
	// And then we try to find shortest path from bc to ab.
	// After that just repeat that path until we get sequence of required length
	//start := trigrams[0].Trigram
	finish, start := headTail(trigrams[0].Trigram)

	// Build graph
	edges := make([]edge, 0, len(trigrams))
	vertices := make(map[string]bool)
	for _, trigram := range trigrams {
		if trigram.Score > 0 {
			h, t := headTail(trigram.Trigram)
			edges = append(edges, edge{
				v1: h,
				v2: t,
				w:  1.0 / trigram.Score,
			})
			vertices[h] = true
			vertices[t] = true
		}
	}

	// compute shortest way to each vertice
	ways := bellmanFord(start, vertices, edges)

	// Trace path
	path := make([]string, 0)
	step := finish
	for {
		path = append(path, step)
		if step == start {
			break // back at start, now reverse path
		}
		if ways[step] == "" {
			path = nil
			break
		}
		step = ways[step]
	}
	var loop []rune
	if len(path) == 0 {
		loop = []rune(trigrams[0].Trigram)
	} else {
		for i := len(path) - 1; i >= 0; i-- {
			r, _ := utf8.DecodeRuneInString(path[i])
			loop = append(loop, r)
		}
	}

	return wrap(loop, length)
}

// wrap repeats loop (slice of runes) enough times to get string of length n
func wrap(loop []rune, l int) string {
	buffer := make([]rune, l)
	for i := range buffer {
		buffer[i] = loop[i%len(loop)]
	}
	return string(buffer)
}

// split abc to ab & bc (with unicode support)
func headTail(trigram string) (string, string) {
	r := []rune(trigram)
	return string(r[:2]), string(r[1:])
}

type edge struct {
	v1, v2 string
	w      float64
}

// bellmanFord algorithm receives graph as list of vertices and edges
// it returns map that says from which vertice goes shortest path to current
func bellmanFord(start string, vertices map[string]bool, edges []edge) map[string]string {
	distance := make(map[string]float64)
	predecessor := make(map[string]string)
	for v := range vertices {
		distance[v] = math.MaxFloat64 // all vertices are unreachable by default
		predecessor[v] = ""
	}
	distance[start] = 0                  // distance from start to itself is zero
	for i := 1; i < len(vertices); i++ { // need len(vertices) - 1 repetitions
		for _, e := range edges {
			if distance[e.v1]+e.w < distance[e.v2] {
				distance[e.v2] = distance[e.v1] + e.w
				predecessor[e.v2] = e.v1
			}
		}
	}
	return predecessor
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
	return float64(ts.Count) * effortResult(duration)
}

type TrigramScore struct {
	Trigram string
	Score   float64
}

// return list of trigrams with their relative importance to train
// the more frequent is trigram and the more long it takes to type it
// the more important will it be to train it
func (s stats) trigramsToTrain() []TrigramScore {
	res := make([]TrigramScore, 0, len(s.Trigrams))
	for t, ts := range s.Trigrams {
		sc := ts.Score(s.AverageCharDuration() * 3)
		res = append(res, TrigramScore{
			Trigram: t,
			Score:   sc,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})
	return res
}

type markovChain map[string]map[rune]float64

const NWeakest = 10

func markovSequence(trigrams []TrigramScore, length int) string {
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
	for _, r := range trigrams[rand.Intn(NWeakest)].Trigram {
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

func GetReport() (string, error) {
	stats, err := loadStats()
	if err != nil {
		return "", err
	}
	res := make([]string, 0)
	print := func(f string, args ...interface{}) {
		res = append(res, fmt.Sprintf(f, args...))
	}
	print("Total characters typed: %d\n", stats.TotalCharsTyped)
	print("Total time in training: %s\n", time.Second*time.Duration(stats.TotalSessionsDuration))
	print("Average typing speed: %.1f wpm\n", float64(stats.TotalCharsTyped)/stats.TotalSessionsDuration*60.0/5.0)
	print("Training sessions: %d\n", stats.SessionsCount)

	trigrams := stats.trigramsToTrain()
	if len(trigrams) > 0 {
		print("\nTrigrams that need to be trained most:\n")
		print("Trigram | Score | Frequency | Typing time\n")
		avDur := stats.AverageCharDuration() * 3.0
		for _, t := range trigrams[:10] {
			d := stats.Trigrams[t.Trigram]
			tr := fmt.Sprintf("%#v", t.Trigram)
			print(
				"%7s | %5.2f | %9d | %4.2fs\n",
				tr, t.Score, d.Count, d.Duration.Average(avDur),
			)
		}
	}
	return strings.Join(res, ""), nil
}
