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
	host, err := probeHost(options)
	if err != nil {
		return err
	}

	runUpgrade := func() {
		log.Printf("Running host upgrades...")

		host.Upgrade()
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
