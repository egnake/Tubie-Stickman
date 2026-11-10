package main

import (
	"os"

	"tubie-stickman/internal/tubie"
)

func main() {
	os.Exit(tubie.Main(os.Args[1:]))
}
