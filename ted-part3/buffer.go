package main

import (
	"os"
	"strings"
)

type source uint8

const (
	original source = iota
	added
)

type piece struct {
	source        source
	start, length int
}

type tableSnapshot struct {
	pieces   []piece
	row, col int
}

type Buffer struct {
	original, add []rune
	pieces        []piece
	filename      string
	dirty         bool
	undo, redo    []tableSnapshot
}

func openBuffer(filename string) (*Buffer, error) {
	b := &Buffer{filename: filename}
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
	b.original = []rune(text)
	if len(b.original) > 0 {
		b.pieces = []piece{{original, 0, len(b.original)}}
	}
	return b, nil
}

func (b *Buffer) slice(p piece) []rune {
	if p.source == original {
		return b.original[p.start : p.start+p.length]
	}
	return b.add[p.start : p.start+p.length]
}

func (b *Buffer) runes() []rune {
	n := 0
	for _, p := range b.pieces {
		n += p.length
	}
	out := make([]rune, 0, n)
	for _, p := range b.pieces {
		out = append(out, b.slice(p)...)
	}
	return out
}

func (b *Buffer) text() string { return string(b.runes()) }

func (b *Buffer) lines() [][]rune {
	rs := b.runes()
	parts := strings.Split(string(rs), "\n")
	out := make([][]rune, len(parts))
	for i := range parts {
		out[i] = []rune(parts[i])
	}
	return out
}

func (b *Buffer) line(row int) []rune { return b.lines()[row] }

func (b *Buffer) lineCount() int { return len(b.lines()) }

func (b *Buffer) offset(row, col int) int {
	lines := b.lines()
	off := 0
	for i := 0; i < row; i++ {
		off += len(lines[i]) + 1
	}
	return off + col
}

func (b *Buffer) position(off int) (int, int) {
	row, col := 0, 0
	for _, r := range b.runes() {
		if off == 0 {
			return row, col
		}
		off--
		if r == '\n' {
			row++
			col = 0
		} else {
			col++
		}
	}
	return row, col
}

func clonePieces(ps []piece) []piece { return append([]piece(nil), ps...) }

func (b *Buffer) checkpoint(row, col int) {
	b.undo = append(b.undo, tableSnapshot{clonePieces(b.pieces), row, col})
	if len(b.undo) > 200 {
		b.undo = b.undo[1:]
	}
	b.redo = nil
}

func (b *Buffer) undoEdit(row, col int) (int, int, bool) {
	if len(b.undo) == 0 {
		return row, col, false
	}
	b.redo = append(b.redo, tableSnapshot{clonePieces(b.pieces), row, col})
	s := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]
	b.pieces = clonePieces(s.pieces)
	b.dirty = true
	return s.row, s.col, true
}

func (b *Buffer) redoEdit(row, col int) (int, int, bool) {
	if len(b.redo) == 0 {
		return row, col, false
	}
	b.undo = append(b.undo, tableSnapshot{clonePieces(b.pieces), row, col})
	s := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]
	b.pieces = clonePieces(s.pieces)
	b.dirty = true
	return s.row, s.col, true
}

func (b *Buffer) insertAt(off int, text []rune) {
	if len(text) == 0 {
		return
	}
	start := len(b.add)
	b.add = append(b.add, text...)
	fresh := piece{added, start, len(text)}
	if len(b.pieces) == 0 {
		b.pieces = []piece{fresh}
		return
	}
	pos := 0
	for i, p := range b.pieces {
		next := pos + p.length
		if off <= next {
			inside := off - pos
			replacement := make([]piece, 0, 3)
			if inside > 0 {
				replacement = append(replacement, piece{p.source, p.start, inside})
			}
			replacement = append(replacement, fresh)
			if inside < p.length {
				replacement = append(replacement, piece{p.source, p.start + inside, p.length - inside})
			}
			b.pieces = append(b.pieces[:i], append(replacement, b.pieces[i+1:]...)...)
			b.coalesce()
			return
		}
		pos = next
	}
	b.pieces = append(b.pieces, fresh)
	b.coalesce()
}

func (b *Buffer) deleteRange(start, end int) {
	if start >= end {
		return
	}
	out := make([]piece, 0, len(b.pieces))
	pos := 0
	for _, p := range b.pieces {
		pStart, pEnd := pos, pos+p.length
		if pEnd <= start || pStart >= end {
			out = append(out, p)
		} else {
			if start > pStart {
				out = append(out, piece{p.source, p.start, start - pStart})
			}
			if end < pEnd {
				cut := end - pStart
				out = append(out, piece{p.source, p.start + cut, pEnd - end})
			}
		}
		pos = pEnd
	}
	b.pieces = out
	b.coalesce()
}

func (b *Buffer) coalesce() {
	out := make([]piece, 0, len(b.pieces))
	for _, p := range b.pieces {
		if p.length == 0 {
			continue
		}
		n := len(out)
		if n > 0 && out[n-1].source == p.source && out[n-1].start+out[n-1].length == p.start {
			out[n-1].length += p.length
		} else {
			out = append(out, p)
		}
	}
	b.pieces = out
}

func (b *Buffer) insert(row, col int, r rune) {
	b.checkpoint(row, col)
	b.insertAt(b.offset(row, col), []rune{r})
	b.dirty = true
}

func (b *Buffer) newline(row, col int) {
	b.checkpoint(row, col)
	b.insertAt(b.offset(row, col), []rune{'\n'})
	b.dirty = true
}

func (b *Buffer) backspace(row, col int) (int, int) {
	off := b.offset(row, col)
	if off == 0 {
		return row, col
	}
	b.checkpoint(row, col)
	b.deleteRange(off-1, off)
	b.dirty = true
	return b.position(off - 1)
}

func (b *Buffer) delete(row, col int) {
	off := b.offset(row, col)
	if off >= len(b.runes()) {
		return
	}
	b.checkpoint(row, col)
	b.deleteRange(off, off+1)
	b.dirty = true
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
