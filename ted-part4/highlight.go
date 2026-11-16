package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func resolvedTheme(requested string) string {
	if requested != "auto" {
		return requested
	}
	// COLORFGBG commonly ends in the terminal background palette index.
	parts := strings.Split(os.Getenv("COLORFGBG"), ";")
	if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil && n >= 7 {
		return "light"
	}
	return "dark"
}

func highlightLines(text, filename, requestedTheme string) []string {
	text = expandTabs(text)
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".go" && ext != ".js" && ext != ".mjs" && ext != ".cjs" && ext != ".md" && ext != ".markdown" {
		return strings.Split(text, "\n")
	}
	lexer := lexers.Match(filename)
	if lexer == nil {
		return strings.Split(text, "\n")
	}
	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return strings.Split(text, "\n")
	}
	styleName := "monokai"
	if resolvedTheme(requestedTheme) == "light" {
		styleName = "github"
	}
	style := styles.Get(styleName)
	var out bytes.Buffer
	if err := formatters.TTY16m.Format(&out, style, iterator); err != nil {
		return strings.Split(text, "\n")
	}
	formatted := out.String()
	if !strings.HasSuffix(text, "\n") {
		formatted = strings.TrimSuffix(formatted, "\n")
	}
	return strings.Split(formatted, "\n")
}
