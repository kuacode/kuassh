package main

import (
	"fmt"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-tty"
)

func main() {
	t, err := tty.Open()
	if err != nil {
		fmt.Print(err)
	}
	defer t.Close()
	out := colorable.NewColorable(t.Output())
	fmt.Fprintln(out, "\x1b[2J")
	for {
		r, _ := t.ReadRune()
		if r == 0 {
			continue
		}
		fmt.Fprintf(out, "0x%X: %c\n", r, r)
	}
}
