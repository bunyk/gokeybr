package stats

const Capacity = 10

// Window will hold last Capacity values in circular buffer to compute running averages
type Window struct {
	Length int
	Index  int
	Values [Capacity]float64
}

func (w *Window) Append(val float64) {
	w.Values[w.Index] = val
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
		sum += w.Values[i]
	}
	return sum / float64(w.Length)
}
