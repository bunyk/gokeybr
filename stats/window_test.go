package stats

import "testing"

func TestWindowCounting(t *testing.T) {
	stats := make(map[string]*Window)

	stats["test"] = WindowAppend(stats["test"], 1.0)
	got := stats["test"].Average()
	if got != 1.0 {
		t.Errorf("Average of 1 value 1.0 should be 1.0, got %f", got)
	}
}
