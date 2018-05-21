package main

import (
	"flag"
	"log"
)

type Options struct {
	Schedule string
	Cmd      []string
}

func run(options Options) error {
	log.Printf("cmd: %v", options.Cmd)

	runUpgrade := func() {
		log.Printf("Running host upgrades...")

		if err := SystemdExec(options.Cmd); err != nil {
			log.Fatalf("exec %v: %v", options.Cmd, err)
		} else {
			log.Printf("exec %v", options.Cmd)
		}
	}

	if runner, err := scheduledRunner(options, runUpgrade); err != nil {
		return err
	} else {
		runner()
	}

	return nil
}

func main() {
	var options Options

	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.Parse()

	options.Cmd = flag.Args()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
