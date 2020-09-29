package stats

// Window will hold last Capacity values in circular buffer to compute running averages
type Window struct {
	Length   int
	Capacity int
	Index    int
	Values   []float64
}
