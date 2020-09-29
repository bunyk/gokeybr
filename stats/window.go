package stats

const Capacity = 10

// Window will hold last Capacity values in circular buffer to compute running averages
type Window struct {
	Length int
	Index  int
	Values [Capacity]float64
}

func WindowAppend(w *Window, val float64) *Window {
	if w == nil {
		w = &Window{}
	}
	w.Values[w.Index] = val
	w.Index = (w.Index + 1) % Capacity
	if w.Length < Capacity {
		w.Length++
	}
	return w
}

func (w Window) Average() float64 {
	sum := 0.0
	for i := 0; i < w.Length; i++ {
		sum += w.Values[i]
	}
	return sum / float64(w.Length)
}
