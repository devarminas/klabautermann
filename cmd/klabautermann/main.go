package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	pipe := flag.Bool("pipe", false, "run as daemon with pipe protocol")
	flag.Parse()

	if *pipe {
		fmt.Fprintln(os.Stderr, "pipe daemon: not implemented")
		os.Exit(1)
	}

	fmt.Println("klabautermann: not implemented")
}
