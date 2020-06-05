package main

import (
	"fmt"
	"log"

	"github.com/mattn/go-tty"
)

func main() {
	t, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()

	go func() {
		for ws := range t.SIGWINCH() {
			fmt.Println("Resized", ws.W, ws.H)
		}
	}()

	clean, err := t.Raw()
	if err != nil {
		log.Fatal(err)
	}
	defer clean()

	fmt.Println("Hit any key")
	// buf := make([]byte,128)
	for {
		r, err := t.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if r != 0 {
			fmt.Println(r)
			if r == 'q' {
				break
			}
		}

		//if err != nil {
		//	log.Fatal(err)
		//}
		//if r == 0 {
		//	continue
		//}
		fmt.Printf("0x%X: %c\n", r, r)
		//if !t.Buffered() {
		//	break
		//}
	}
}
