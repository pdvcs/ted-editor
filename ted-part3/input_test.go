package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestEscapeSequencesLeaveFollowingInput(t *testing.T) {
	for _, tc := range []struct {
		input, want string
	}{
		{"\x1b[1;5Cz", "right"},
		{"\x1b[3~z", "delete"},
		{"\x1b[7~z", "home"},
		{"\x1b[Az", "up"},
	} {
		r := strings.NewReader(tc.input)
		got, err := readKey(r)
		if err != nil || got != tc.want {
			t.Fatalf("readKey(%q) = %q, %v; want %q", tc.input, got, err, tc.want)
		}
		if got, err = readKey(r); err != nil || got != string(tc.input[len(tc.input)-1]) {
			t.Fatalf("following key = %q, %v", got, err)
		}
	}
}

func TestEscapeOnOpenFileReturnsPromptly(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if _, err := w.Write([]byte{27}); err != nil {
		t.Fatal(err)
	}
	done := make(chan string, 1)
	go func() {
		key, _ := readKey(r)
		done <- key
	}()
	select {
	case key := <-done:
		if key != "escape" {
			t.Fatalf("got %q, want escape", key)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("standalone Escape blocked")
	}
}

func TestStandaloneEscapeAndTab(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"\x1b", "escape"}, {"\x1ba", "a"}, {"\x1b\xc3\xa9", "é"},
		{"\x1b\x7f", "backspace"}, {"\t", "tab"},
	} {
		got, err := readKey(strings.NewReader(tc.input))
		if err != nil || got != tc.want {
			t.Fatalf("readKey(%q) = %q, %v; want %q", tc.input, got, err, tc.want)
		}
	}
}
