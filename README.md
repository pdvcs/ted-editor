# Building `ted`, a tiny text editor in Go

This repository is a four-part post learning about terminal text editors.
Each directory is a complete, runnable snapshot of the editor at the end
of each part.

| Post | Part | What changes |
| --- | --- | --- |
| [Part 1 - The Loop](https://pdutta.net/posts/ted-part-1-the-loop/) | [`ted-part1`](./ted-part1) | Raw terminal input, an array-of-lines buffer, opening, editing, and saving |
| [Part 2 - The Simplest Undo](https://pdutta.net/posts/ted-part-2-simplest-undo/) | [`ted-part2`](./ted-part2) | Snapshot-based undo and redo |
|  | [`ted-part3`](./ted-part3) | A piece table and cheap structural history |
|  | [`ted-part4`](./ted-part4) | Chroma syntax highlighting and light/dark themes |

Compare adjacent `main.go` files with `diff -u` to see what each chapter adds
to the editor loop and drawing code.

Every part is its own Go module. Parts 1–3 need Go 1.24 or newer; Part 4
needs Go 1.25 or newer because of its Chroma dependency. From any part directory:

```sh
go run . ../README.md
```

The editor deliberately stays small. It supports arrows, Home/End, Page Up/Page
Down, Tab, Enter, Backspace, Delete, `Ctrl-S` to save, and `Ctrl-Q` to quit. Parts
2–4 add `Ctrl-Z` for undo and `Ctrl-Y` for redo. Part 4 accepts
`--theme=dark`, `--theme=light`, or `--theme=auto`.

These examples target Unix-like terminals. They use `golang.org/x/term` for raw
mode and terminal sizing; part 4 additionally uses Chroma.
