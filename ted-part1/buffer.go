package main

import (
	"os"
	"strings"
)

// Buffer uses the least surprising representation for a first editor: one
// rune slice per line. A rune is a Unicode code point, so columns do not split
// a multi-byte UTF-8 character.
type Buffer struct {
	lines    [][]rune
	filename string
	dirty    bool
}

func openBuffer(filename string) (*Buffer, error) {
	b := &Buffer{filename: filename, lines: [][]rune{{}}}
	if filename == "" {
		return b, nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return b, nil
		}
		return nil, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	parts := strings.Split(text, "\n")
	b.lines = make([][]rune, len(parts))
	for i, line := range parts {
		b.lines[i] = []rune(line)
	}
	return b, nil
}

func (b *Buffer) line(row int) []rune { return b.lines[row] }

func (b *Buffer) lineCount() int { return len(b.lines) }

func (b *Buffer) insert(row, col int, r rune) {
	line := b.lines[row]
	// Grow the slice, then shift its tail right to make room at col.
	line = append(line, 0)
	copy(line[col+1:], line[col:])
	line[col] = r
	b.lines[row] = line
	b.dirty = true
}

func (b *Buffer) newline(row, col int) {
	right := append([]rune(nil), b.lines[row][col:]...)
	b.lines[row] = append([]rune(nil), b.lines[row][:col]...)
	b.lines = append(b.lines, nil)
	copy(b.lines[row+2:], b.lines[row+1:])
	b.lines[row+1] = right
	b.dirty = true
}

func (b *Buffer) backspace(row, col int) (int, int) {
	if col > 0 {
		line := b.lines[row]
		b.lines[row] = append(line[:col-1], line[col:]...)
		b.dirty = true
		return row, col - 1
	}
	if row == 0 {
		return row, col
	}
	newCol := len(b.lines[row-1])
	b.lines[row-1] = append(b.lines[row-1], b.lines[row]...)
	b.lines = append(b.lines[:row], b.lines[row+1:]...)
	b.dirty = true
	return row - 1, newCol
}

func (b *Buffer) delete(row, col int) {
	if col < len(b.lines[row]) {
		line := b.lines[row]
		b.lines[row] = append(line[:col], line[col+1:]...)
		b.dirty = true
	} else if row+1 < len(b.lines) {
		b.lines[row] = append(b.lines[row], b.lines[row+1]...)
		b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
		b.dirty = true
	}
}

func (b *Buffer) text() string {
	lines := make([]string, len(b.lines))
	for i, line := range b.lines {
		lines[i] = string(line)
	}
	return strings.Join(lines, "\n")
}

func (b *Buffer) save() error {
	if b.filename == "" {
		return os.ErrInvalid
	}
	if err := os.WriteFile(b.filename, []byte(b.text()+"\n"), 0o644); err != nil {
		return err
	}
	b.dirty = false
	return nil
}
