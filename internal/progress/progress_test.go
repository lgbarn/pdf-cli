package progress

import "testing"

func TestNewProgressBar(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		threshold int
		wantNil   bool
	}{
		{"below threshold", 3, 5, true},
		{"equals threshold", 5, 5, true},
		{"above threshold", 10, 5, false},
		{"zero threshold", 5, 0, false},
		{"zero total", 0, 5, true},
		{"negative threshold", 5, -1, false},
		{"boundary +1", 2, 1, false},
		{"boundary equal large", 10, 10, true},
		{"boundary +1 large", 11, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar("Test", tt.total, tt.threshold)
			if (bar == nil) != tt.wantNil {
				t.Errorf("NewProgressBar(_, %d, %d) nil=%v, want nil=%v", tt.total, tt.threshold, bar == nil, tt.wantNil)
			}
		})
	}
}

func TestNewProgressBarUsage(t *testing.T) {
	bar := NewProgressBar("Test", 10, 5)
	if bar == nil {
		t.Fatal("returned nil when total > threshold")
	}
	if err := bar.Add(1); err != nil {
		t.Errorf("bar.Add() error = %v", err)
	}
}

func TestNewProgressBarDescriptions(t *testing.T) {
	for _, desc := range []string{"Processing files", "", "Very long description"} {
		bar := NewProgressBar(desc, 10, 0)
		if bar == nil {
			t.Errorf("NewProgressBar(%q, ...) returned nil", desc)
		}
	}
}

func TestNewBytesProgressBar(t *testing.T) {
	for _, total := range []int64{1024, 1024 * 1024, 1024 * 1024 * 1024, 0} {
		bar := NewBytesProgressBar("Downloading", total)
		if bar == nil {
			t.Errorf("NewBytesProgressBar(_, %d) returned nil", total)
		}
	}

	bar := NewBytesProgressBar("Download", 1000)
	if err := bar.Add64(100); err != nil {
		t.Errorf("Add64() error = %v", err)
	}
	if err := bar.Add64(400); err != nil {
		t.Errorf("Add64() error = %v", err)
	}
}

func TestFinishProgressBar(t *testing.T) {
	FinishProgressBar(nil)
	FinishProgressBar(NewProgressBar("Test", 10, 0))

	bar := NewProgressBar("Test", 5, 0)
	for i := 0; i < 5; i++ {
		_ = bar.Add(1)
	}
	FinishProgressBar(bar)
}

func TestProgressBarTheme(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{ProgressBarTheme.Saucer, "="},
		{ProgressBarTheme.SaucerHead, ">"},
		{ProgressBarTheme.SaucerPadding, " "},
		{ProgressBarTheme.BarStart, "["},
		{ProgressBarTheme.BarEnd, "]"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("theme value = %q, want %q", tt.got, tt.want)
		}
	}
}

func TestMultipleProgressBars(t *testing.T) {
	bar1 := NewProgressBar("Task 1", 10, 0)
	bar2 := NewProgressBar("Task 2", 20, 0)
	bar3 := NewBytesProgressBar("Download", 1000)

	if bar1 == nil || bar2 == nil || bar3 == nil {
		t.Fatal("Failed to create multiple progress bars")
	}

	_ = bar1.Add(5)
	_ = bar2.Add(10)
	_ = bar3.Add64(500)

	FinishProgressBar(bar1)
	FinishProgressBar(bar2)
	FinishProgressBar(bar3)
}
