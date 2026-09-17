package main

import "testing"

func TestAbs(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}

	for _, tt := range tests {
		if got := abs(tt.input); got != tt.want {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseFlags(t *testing.T) {
	up := true
	down := false
	reset := false
	version := true
	steps := 3
	force := -1

	opts := parseFlags(&up, &down, &reset, &version, &steps, &force)

	if !opts.up {
		t.Error("expected up to be true")
	}
	if opts.down {
		t.Error("expected down to be false")
	}
	if opts.reset {
		t.Error("expected reset to be false")
	}
	if !opts.version {
		t.Error("expected version to be true")
	}
	if opts.steps != 3 {
		t.Errorf("expected steps 3, got %d", opts.steps)
	}
	if opts.force != -1 {
		t.Errorf("expected force -1, got %d", opts.force)
	}
}
