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
	os.Stdout.WriteString(Convert(string(input)))
}
