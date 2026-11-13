package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(input io.Reader, output io.Writer) error {
	dependencies := newContainer(input, output)
	console, err := dependencies.Console()
	if err != nil {
		return err
	}
	return console.Run()
}
