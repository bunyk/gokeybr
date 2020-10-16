package stats

import "math"

const Capacity = 10

// Window will hold last Capacity values in circular buffer to compute running averages
type Window struct {
	Length int           `json:"l"`
	Index  int           `json:"i"`
	Values [Capacity]int `json:"v"`
}

const MillisecondsInSecond = 1000.0

func (w *Window) Append(val float64) {
	w.Values[w.Index] = int(math.Round(val * 1000.0))
	w.Index = (w.Index + 1) % Capacity
	if w.Length < Capacity {
		w.Length++
	}
}

func (w Window) Average(def float64) float64 {
	sum := 0.0
	if w.Length == 0 {
		return def
	}
	for i := 0; i < w.Length; i++ {
		sum += float64(w.Values[i]) / 1000.0
	}
	return sum / float64(w.Length)
}
