package main

import "testing"

func TestUnhandledKeyNamesDoNotInsertText(t *testing.T) {
	b, err := openBuffer("")
	if err != nil {
		t.Fatal(err)
	}
	e := &editor{buf: b}
	for _, key := range []string{"escape", "ctrl-x", "ctrl-z", "ctrl-y", string([]byte{127})} {
		e.handle(key)
		if b.text() != "" || b.dirty {
			t.Fatalf("key %q changed buffer to %q", key, b.text())
		}
	}
	e.handle("é")
	if b.text() != "é" {
		t.Fatalf("printable rune not inserted: %q", b.text())
	}
}
