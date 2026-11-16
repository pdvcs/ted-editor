package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

type editor struct {
	buf                 *Buffer
	row, col, rowOffset int
	status              string
	quitConfirm         bool
	in, out             *os.File
	theme               string
}

func main() {
	name, theme := "", "auto"
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--theme=") {
			theme = strings.TrimPrefix(arg, "--theme=")
		} else {
			name = arg
		}
	}
	if theme != "auto" && theme != "dark" && theme != "light" {
		fmt.Fprintln(os.Stderr, "theme must be auto, dark, or light")
		os.Exit(2)
	}
	b, err := openBuffer(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "ted needs an interactive terminal")
		os.Exit(1)
	}
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		term.Restore(int(os.Stdin.Fd()), old)
		fmt.Fprint(os.Stdout, "\x1b[?7h\x1b[2J\x1b[H\x1b[?25h")
	}()
	e := &editor{buf: b, theme: theme, in: os.Stdin, out: os.Stdout}
	if err := e.run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (e *editor) run() error {
	for {
		e.draw()
		k, err := readKey(e.in)
		if err != nil {
			return err
		}
		if e.handle(k) {
			return nil
		}
	}
}

func (e *editor) handle(k string) bool {
	e.status = ""
	if k != "ctrl-q" {
		e.quitConfirm = false
	}
	switch k {
	case "ctrl-q":
		if e.buf.dirty && !e.quitConfirm {
			e.status = "unsaved changes; Ctrl-Q again to quit"
			e.quitConfirm = true
			return false
		}
		return true
	case "ctrl-s":
		if e.buf.filename == "" {
			e.status = "start ted with a filename to save"
		} else if err := e.buf.save(); err != nil {
			e.status = "save: " + err.Error()
		} else {
			e.status = "saved"
		}
	case "ctrl-z":
		var ok bool
		e.row, e.col, ok = e.buf.undoEdit(e.row, e.col)
		if !ok {
			e.status = "nothing to undo"
		}
	case "ctrl-y":
		var ok bool
		e.row, e.col, ok = e.buf.redoEdit(e.row, e.col)
		if !ok {
			e.status = "nothing to redo"
		}
	case "left":
		if e.col > 0 {
			e.col--
		} else if e.row > 0 {
			e.row--
			e.col = len(e.buf.line(e.row))
		}
	case "right":
		if e.col < len(e.buf.line(e.row)) {
			e.col++
		} else if e.row+1 < e.buf.lineCount() {
			e.row++
			e.col = 0
		}
	case "up":
		if e.row > 0 {
			e.row--
			e.clamp()
		}
	case "down":
		if e.row+1 < e.buf.lineCount() {
			e.row++
			e.clamp()
		}
	case "home":
		e.col = 0
	case "end":
		e.col = len(e.buf.line(e.row))
	case "pageup":
		_, h, _ := term.GetSize(int(e.out.Fd()))
		e.row = max(0, e.row-max(1, h-2))
		e.clamp()
	case "pagedown":
		_, h, _ := term.GetSize(int(e.out.Fd()))
		e.row = min(e.buf.lineCount()-1, e.row+max(1, h-2))
		e.clamp()
	case "tab":
		e.buf.insert(e.row, e.col, '\t')
		e.col++
	case "enter":
		e.buf.newline(e.row, e.col)
		e.row++
		e.col = 0
	case "backspace":
		e.row, e.col = e.buf.backspace(e.row, e.col)
	case "delete":
		e.buf.delete(e.row, e.col)
	default:
		r, n := utf8.DecodeRuneInString(k)
		if n == len(k) && r >= 32 && r != 127 && r != utf8.RuneError {
			e.buf.insert(e.row, e.col, r)
			e.col++
		}
	}
	return false
}

func (e *editor) clamp() { e.col = min(e.col, len(e.buf.line(e.row))) }

func (e *editor) draw() {
	w, h, err := term.GetSize(int(e.out.Fd()))
	if err != nil {
		w, h = 80, 24
	}
	rows := max(1, h-1)
	if e.row < e.rowOffset {
		e.rowOffset = e.row
	}
	if e.row >= e.rowOffset+rows {
		e.rowOffset = e.row - rows + 1
	}
	out := []byte("\x1b[?7l\x1b[?25l\x1b[H")
	lines := highlightLines(e.buf.text(), e.buf.filename, e.theme)
	for sr := 0; sr < rows; sr++ {
		fr := e.rowOffset + sr
		if fr < len(lines) {
			out = append(out, lines[fr]...)
		} else {
			out = append(out, '~')
		}
		out = append(out, "\x1b[K\r\n"...)
	}
	name := "[new file]"
	if e.buf.filename != "" {
		name = filepath.Base(e.buf.filename)
	}
	mark := ""
	if e.buf.dirty {
		mark = " [+]"
	}
	s := fmt.Sprintf(" %s%s %d:%d  %s theme | ^S save ^Z undo ^Y redo ^Q quit", name, mark, e.row+1, e.col+1, resolvedTheme(e.theme))
	if e.status != "" {
		s = " " + e.status
	}
	r := []rune(s)
	if len(r) > w {
		r = r[:w]
	}
	out = append(out, "\x1b[0m\x1b[7m"...)
	out = append(out, string(r)...)
	for len(r) < w {
		out = append(out, ' ')
		r = append(r, ' ')
	}
	out = append(out, "\x1b[m"...)
	out = append(out, fmt.Sprintf("\x1b[%d;%dH\x1b[?25h", e.row-e.rowOffset+1, min(w, visualColumn(e.buf.line(e.row), e.col)+1))...)
	e.out.Write(out)
}
