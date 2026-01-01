package main

import (
	"io"
	"log"
	"os"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stdout.WriteString(Convert(string(input)))
	if err != nil {
		log.Fatal(err)
	}
}
