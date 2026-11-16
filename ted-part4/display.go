package main

import "strings"

const tabWidth = 8

// visualColumn counts terminal cells for the part of a line before the cursor.
func visualColumn(line []rune, col int) int {
	x := 0
	for _, r := range line[:min(col, len(line))] {
		if r == '\t' {
			x += tabWidth - x%tabWidth
		} else {
			x++
		}
	}
	return x
}

// expandTabs makes painted text and cursor positions agree even when the
// terminal uses different tab stops. It leaves the buffer's text untouched.
func expandTabs(text string) string {
	var out strings.Builder
	x := 0
	for _, r := range text {
		switch r {
		case '\n':
			out.WriteRune(r)
			x = 0
		case '\t':
			spaces := tabWidth - x%tabWidth
			out.WriteString(strings.Repeat(" ", spaces))
			x += spaces
		default:
			out.WriteRune(r)
			x++
		}
	}
	return out.String()
}
