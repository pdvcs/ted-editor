package main

import (
	"os"
	"strings"
)

type snapshot struct {
	lines    [][]rune
	row, col int
}

type Buffer struct {
	lines      [][]rune
	filename   string
	dirty      bool
	undo, redo []snapshot
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
	text := strings.TrimSuffix(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	parts := strings.Split(text, "\n")
	b.lines = make([][]rune, len(parts))
	for i := range parts {
		b.lines[i] = []rune(parts[i])
	}
	return b, nil
}

func cloneLines(lines [][]rune) [][]rune {
	copyLines := make([][]rune, len(lines))
	for i := range lines {
		copyLines[i] = append([]rune(nil), lines[i]...)
	}
	return copyLines
}

func (b *Buffer) checkpoint(row, col int) {
	b.undo = append(b.undo, snapshot{cloneLines(b.lines), row, col})
	if len(b.undo) > 200 {
		b.undo = b.undo[1:]
	}
	b.redo = nil
}

func (b *Buffer) undoEdit(row, col int) (int, int, bool) {
	if len(b.undo) == 0 {
		return row, col, false
	}
	b.redo = append(b.redo, snapshot{cloneLines(b.lines), row, col})
	last := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]
	b.lines, b.dirty = cloneLines(last.lines), true
	return last.row, last.col, true
}

func (b *Buffer) redoEdit(row, col int) (int, int, bool) {
	if len(b.redo) == 0 {
		return row, col, false
	}
	b.undo = append(b.undo, snapshot{cloneLines(b.lines), row, col})
	last := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]
	b.lines, b.dirty = cloneLines(last.lines), true
	return last.row, last.col, true
}

func (b *Buffer) line(row int) []rune { return b.lines[row] }

func (b *Buffer) lineCount() int { return len(b.lines) }

func (b *Buffer) insert(row, col int, r rune) {
	b.checkpoint(row, col)
	line := append(b.lines[row], 0)
	copy(line[col+1:], line[col:])
	line[col] = r
	b.lines[row] = line
	b.dirty = true
}

func (b *Buffer) newline(row, col int) {
	b.checkpoint(row, col)
	right := append([]rune(nil), b.lines[row][col:]...)
	b.lines[row] = append([]rune(nil), b.lines[row][:col]...)
	b.lines = append(b.lines, nil)
	copy(b.lines[row+2:], b.lines[row+1:])
	b.lines[row+1] = right
	b.dirty = true
}

func (b *Buffer) backspace(row, col int) (int, int) {
	if col == 0 && row == 0 {
		return row, col
	}
	b.checkpoint(row, col)
	if col > 0 {
		line := b.lines[row]
		b.lines[row] = append(line[:col-1], line[col:]...)
		b.dirty = true
		return row, col - 1
	}
	n := len(b.lines[row-1])
	b.lines[row-1] = append(b.lines[row-1], b.lines[row]...)
	b.lines = append(b.lines[:row], b.lines[row+1:]...)
	b.dirty = true
	return row - 1, n
}

func (b *Buffer) delete(row, col int) {
	if col == len(b.lines[row]) && row+1 == len(b.lines) {
		return
	}
	b.checkpoint(row, col)
	if col < len(b.lines[row]) {
		line := b.lines[row]
		b.lines[row] = append(line[:col], line[col+1:]...)
	} else {
		b.lines[row] = append(b.lines[row], b.lines[row+1]...)
		b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
	}
	b.dirty = true
}

func (b *Buffer) text() string {
	lines := make([]string, len(b.lines))
	for i := range b.lines {
		lines[i] = string(b.lines[i])
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
