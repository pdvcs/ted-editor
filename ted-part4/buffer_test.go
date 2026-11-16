package main

import "testing"

func TestPieceTableEditing(t *testing.T) {
	b := &Buffer{original: []rune("hello world"), pieces: []piece{{original, 0, 11}}}
	b.insert(0, 5, ',')
	b.insert(0, 6, '!')
	if got := b.text(); got != "hello,! world" {
		t.Fatalf("got %q", got)
	}
	b.delete(0, 6)
	if got := b.text(); got != "hello, world" {
		t.Fatalf("got %q", got)
	}
	b.newline(0, 6)
	if got := b.text(); got != "hello,\n world" {
		t.Fatalf("got %q", got)
	}
}

func TestUndoRestoresPieceList(t *testing.T) {
	b := &Buffer{original: []rune("abc"), pieces: []piece{{original, 0, 3}}}
	b.insert(0, 1, 'X')
	if b.text() != "aXbc" {
		t.Fatal(b.text())
	}
	row, col, ok := b.undoEdit(0, 2)
	if !ok || row != 0 || col != 1 || b.text() != "abc" {
		t.Fatalf("undo: %d %d %v %q", row, col, ok, b.text())
	}
	_, _, ok = b.redoEdit(row, col)
	if !ok || b.text() != "aXbc" {
		t.Fatalf("redo: %v %q", ok, b.text())
	}
}

func TestUnicodeUsesRuneOffsets(t *testing.T) {
	b := &Buffer{original: []rune("aλz"), pieces: []piece{{original, 0, 3}}}
	b.insert(0, 2, '!')
	if b.text() != "aλ!z" {
		t.Fatal(b.text())
	}
}
