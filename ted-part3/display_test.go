package main

import "testing"

func TestTabsUseEightColumnStops(t *testing.T) {
	line := []rune("a\t\tb")
	if got := visualColumn(line, 3); got != 16 {
		t.Fatalf("visual column = %d, want 16", got)
	}
	if got := renderLine(line, 17); got != "a               b" {
		t.Fatalf("rendered line = %q", got)
	}
	if got := renderLine(line, 10); got != "a         " {
		t.Fatalf("clipped line = %q", got)
	}
}
