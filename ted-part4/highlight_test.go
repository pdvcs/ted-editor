package main

import (
	"strings"
	"testing"
)

func TestSupportedLanguageIsHighlighted(t *testing.T) {
	lines := highlightLines("package main\n", "main.go", "dark")
	if len(lines) != 2 || lines[1] != "" || !strings.Contains(lines[0], "\x1b[") {
		t.Fatalf("expected ANSI styling, got %q", lines)
	}
}

func TestHighlightedTextWithoutTrailingNewlineHasOneLine(t *testing.T) {
	lines := highlightLines("package main", "main.go", "dark")
	if len(lines) != 1 {
		t.Fatalf("expected one line, got %q", lines)
	}
}

func TestHighlightedTabsAreExpanded(t *testing.T) {
	lines := highlightLines("\tpackage main\n", "main.go", "dark")
	if len(lines) != 2 || strings.Contains(strings.Join(lines, "\n"), "\t") {
		t.Fatalf("expected two lines without literal tabs, got %q", lines)
	}
}

func TestUnknownLanguageStaysPlain(t *testing.T) {
	lines := highlightLines("hello\n\tworld", "notes.txt", "dark")
	if got := strings.Join(lines, "\n"); got != "hello\n        world" {
		t.Fatalf("got %q", got)
	}
}
