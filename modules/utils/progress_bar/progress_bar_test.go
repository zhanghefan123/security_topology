package progress_bar

import "testing"

func TestProgressBar(t *testing.T) {
	bar := NewProgressBar(5, "create")
	bar.Add(1)
}
