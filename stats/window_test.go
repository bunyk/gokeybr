package stats

import "testing"

func TestWindowCounting(t *testing.T) {
	var w Window

	if w.Average(45.0) != 45.0 {
		t.Errorf("Average of 1 value 1.0 should equal to default")
	}
	w.Append(1.0)

	got := w.Average(13.0)
	if got != 1.0 {
		t.Errorf("Average of 1 value 1.0 should be 1.0, got %f", got)
	}
}
