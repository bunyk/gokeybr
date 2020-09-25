package phrase

import (
	"math/rand"
	"strings"
	"time"
)

// Generator generates a phrase
type Generator interface {
	Phrase() string
}

type StaticGenerator struct {
	Text string
}

func (sg StaticGenerator) Phrase() string {
	return sg.Text
}

var DefaultGenerator = StaticGenerator{"the quick brown fox jumps over the lazy dog"}

// RandomGenerator composes a random phrase with given length from given words.
type RandomGenerator struct {
	Words     []string
	MinLength int
	seed      int64
}

func NewRandomGenerator(words []string, minLength int) *RandomGenerator {
	return &RandomGenerator{
		Words:     words,
		MinLength: minLength,
		seed:      time.Now().UnixNano(),
	}
}

func (rg *RandomGenerator) Phrase() string {
	rand := rand.New(rand.NewSource(rg.seed))
	var phrase []string
	l := -1
	for l < rg.MinLength {
		w := rg.Words[rand.Int31n(int32(len(rg.Words)))]
		phrase = append(phrase, w)
		l += 1 + len(w)
	}
	rg.seed = rand.Int63()
	return strings.Join(phrase, " ")
}

type SequentialLineGenerator struct {
	Lines       []string
	CurrentLine int
}

func (slg *SequentialLineGenerator) Phrase() string {
	cl := slg.CurrentLine
	slg.CurrentLine = (cl + 1) % len(slg.Lines)
	return slg.Lines[cl]
}
