package stats

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
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
const speedOfLight = 150.0 // wpm

func effortResult(trigramTime float64) float64 {
	speed := time2wpm(trigramTime)
	q := speed / speedOfLight
	q = q * q
	if q > 1.0 {
		return 0
	}
	return math.Sqrt(1.0 - q)
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
			r := []rune(path[i])[0]
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
// It is total frequency of trigram (it's count)
// multiplied by current average duration of typing one
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
		if ts.Score == 0 {
			ts.Score = 0.00000001
		}
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

var statsCache *stats

func loadStats() (*stats, error) {
	if statsCache != nil {
		return statsCache, nil
	}
	statsCache = &stats{Trigrams: make(map[string]trigramStat)}
	err := fs.LoadJSON(StatsFile, statsCache)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: File %s does not exist! It will be created.\n", StatsFile)
			return statsCache, nil
		}
		return nil, err
	}
	return statsCache, nil
}

type statLogEntry struct {
	Start    string    `json:"start"`
	Text     string    `json:"text"`
	Timeline []float64 `json:"timeline"`
}

const wpmPer1secTrigramTime = 36.0 // 3 / 5 * 60
func time2wpm(t float64) float64 {
	return wpmPer1secTrigramTime / t
}

func AverageWPM() float64 {
	stats, err := loadStats()
	if err != nil { // If stats loaded to fail
		return 50.0 // return world average
	}
	avDur := stats.AverageCharDuration() * 3.0
	return time2wpm(avDur)
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
	print("Average typing speed: %.1f wpm\n", AverageWPM())
	print("Training sessions: %d\n", stats.SessionsCount)
	var fastestTr, slowestTr string
	fastestTime := 10.0
	slowestTime := 0.0
	for t, s := range stats.Trigrams {
		dur := s.Duration.Average(0)
		if dur < fastestTime {
			fastestTime = dur
			fastestTr = t
		}
		if dur > slowestTime {
			slowestTime = dur
			slowestTr = t
		}
	}
	print("\nTrigram stats:\n")
	print("Slowest: %#v %4.2fs (%.1f wpm)\n", slowestTr, slowestTime, time2wpm(slowestTime))
	print("Fastest: %#v %4.2fs (%.1f wpm)\n", fastestTr, fastestTime, time2wpm(fastestTime))

	trigrams := stats.trigramsToTrain()
	if len(trigrams) > 0 {
		print("\nNeed to be trained most:\n")
		print("Trigram |   Score | Frequency | Typing time\n")
		for _, t := range trigrams[:20] {
			d := stats.Trigrams[t.Trigram]
			tr := fmt.Sprintf("%#v", t.Trigram)
			dur := d.Duration.Average(0)
			print(
				"%7s | %7.2f | %9d | %4.2fs (%.1f wpm)\n",
				tr, t.Score/stats.TotalSessionsDuration*1000.0, d.Count, dur, time2wpm(dur),
			)
			// we divide score to total session duration go get score approximated in promille
			// if trigram will be the only one we type - it will have 1000 score,
			// if it's current typing speed equals to total, average, or less if it is typed faster.
			// if it is typed slower - score will be greater than 1000
		}
	}
	if stats.TotalSessionsDuration < 600 { // Less than 10 minutes of training, not much to show
		print("\nTrain more to get some progress!")
		return strings.Join(res, ""), nil
	}
	progressInterval := time.Minute * 10      // Show progress in 10 minute intervals
	if stats.TotalSessionsDuration > 2*3600 { // If trained for more than 2 hours - in 30 minutes intervals
		progressInterval = time.Minute * 30
	}
	if stats.TotalSessionsDuration > 10*3600 { // If trained for more than 10 hours - in hour intervals
		progressInterval = time.Minute * 30
	}
	progress, err := wpmProgress(progressInterval)
	if err != nil {
		return "", err
	}
	print("\nTraining progress:\n")
	print("   Time | WPM\n")
	for i, wpm := range progress {
		print("%7s | %.1f\n", formatDuration(time.Duration(i)*progressInterval), wpm)
	}
	return strings.Join(res, ""), nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh%dm", h, m-h*60)
}

const WPMinCPS = 12.0

func calcWPM(chars int, seconds float64) float64 {
	return float64(chars) / seconds * WPMinCPS
}

func wpmProgress(intervalSize time.Duration) ([]float64, error) {
	logStatsIter, err := fs.NewJSONLinesIterator(LogStatsFile)
	if err != nil {
		return nil, err
	}
	defer logStatsIter.Close()

	var logEntry statLogEntry
	iSec := intervalSize.Seconds()
	var countedSeconds float64
	var countedChars int
	var res []float64
	for {
		cont, err := logStatsIter.UnmarshalNextLine(&logEntry)
		if err != nil {
			return nil, err
		}
		if !cont {
			break
		}
		for i, t := range logEntry.Timeline {
			if t-countedSeconds >= iSec { // Counted approximately for interval
				res = append(res, calcWPM(i-countedChars, t-countedSeconds))
				countedSeconds = t
				countedChars = i
			}
		}
		// compute counting debt
		countedSeconds = countedSeconds - logEntry.Timeline[len(logEntry.Timeline)-1]
		countedChars = countedChars - len(logEntry.Timeline)
	}
	res = append(res, calcWPM(-countedChars, -countedSeconds))
	return res, nil
}
