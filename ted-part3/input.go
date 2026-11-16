package main

import (
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// Escape sequences arrive as several bytes. A short wait distinguishes a
// standalone Escape key from the start of a terminal control sequence.
func escapeByte(r io.Reader) (byte, error) {
	if f, ok := r.(*os.File); ok {
		fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}
		n, err := unix.Poll(fds, 50)
		if err != nil || n == 0 {
			return 0, io.EOF
		}
	}
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	return b[0], err
}

func readKey(r io.Reader) (string, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return "", err
	}
	return keyFromByte(r, b[0])
}

func keyFromByte(r io.Reader, first byte) (string, error) {
	switch first {
	case 17:
		return "ctrl-q", nil
	case 19:
		return "ctrl-s", nil
	case 26:
		return "ctrl-z", nil
	case 25:
		return "ctrl-y", nil
	case 9:
		return "tab", nil
	case 13:
		return "enter", nil
	case 8, 127:
		return "backspace", nil
	case 27:
		prefix, err := escapeByte(r)
		if err != nil {
			return "escape", nil
		}
		if prefix != '[' && prefix != 'O' {
			// A quick key after Escape is still a key, not a CSI sequence.
			return keyFromByte(r, prefix)
		}
		var seq []byte
		for len(seq) < 32 {
			c, err := escapeByte(r)
			if err != nil {
				return "escape", nil
			}
			seq = append(seq, c)
			if c >= 0x40 && c <= 0x7e {
				break
			}
		}
		if len(seq) == 0 {
			return "escape", nil
		}
		final := seq[len(seq)-1]
		switch final {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		case 'C':
			return "right", nil
		case 'D':
			return "left", nil
		case 'H':
			return "home", nil
		case 'F':
			return "end", nil
		case '~':
			if prefix == '[' && len(seq) == 2 {
				switch seq[0] {
				case '3':
					return "delete", nil
				case '5':
					return "pageup", nil
				case '6':
					return "pagedown", nil
				case '1', '7':
					return "home", nil
				case '4', '8':
					return "end", nil
				}
			}
		}
		return "escape", nil
	default:
		if first < utf8.RuneSelf {
			return string([]byte{first}), nil
		}
		n := utf8SequenceLen(first)
		rest := make([]byte, n-1)
		if _, err := io.ReadFull(r, rest); err != nil {
			return "", err
		}
		return string(append([]byte{first}, rest...)), nil
	}
}

func utf8SequenceLen(first byte) int {
	switch {
	case first&0xf8 == 0xf0:
		return 4
	case first&0xf0 == 0xe0:
		return 3
	case first&0xe0 == 0xc0:
		return 2
	default:
		return 1
	}
}
