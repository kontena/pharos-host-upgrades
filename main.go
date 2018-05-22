package main

import (
	"flag"
	"log"
)

type Options struct {
	Schedule string
}

func run(options Options) error {
	host, err := probeHost(options)
	if err != nil {
		return err
	}

	return runSchedule(options, func() error {
		log.Printf("Running host upgrades...")

		return host.Upgrade()
	})
}

func main() {
	var options Options

	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
