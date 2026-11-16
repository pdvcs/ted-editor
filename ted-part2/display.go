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

func renderLine(line []rune, width int) string {
	var out strings.Builder
	x := 0
	for _, r := range line {
		if x >= width {
			break
		}
		if r == '\t' {
			spaces := min(tabWidth-x%tabWidth, width-x)
			out.WriteString(strings.Repeat(" ", spaces))
			x += spaces
		} else {
			out.WriteRune(r)
			x++
		}
	}
	return out.String()
}
