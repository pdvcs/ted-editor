package main

import "testing"

func TestTabsUseEightColumnStops(t *testing.T) {
	line := []rune("a\t\tb")
	if got := visualColumn(line, 3); got != 16 {
		t.Fatalf("visual column = %d, want 16", got)
	}
	if got := expandTabs("a\t\tb\n\tc"); got != "a               b\n        c" {
		t.Fatalf("expanded text = %q", got)
	}
}
