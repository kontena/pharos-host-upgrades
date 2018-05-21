package main

import (
	"flag"
	"log"
)

type Options struct {
	cmd []string
}

func run(options Options) error {
	log.Printf("exec: %v", options.cmd)

	return SystemdExec(options.cmd)
}

func main() {
	var options Options

	flag.Parse()

	options.cmd = flag.Args()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
