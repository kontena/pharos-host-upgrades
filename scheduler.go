package main

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron"
)

func runSchedule(options Options, f func() error) error {
	if options.Schedule == "" {
		log.Printf("No --schedule given, will run once")

		return f()
	} else if schedule, err := cron.Parse(options.Schedule); err != nil {
		return fmt.Errorf("Invalid --schedule=%v: %v", options.Schedule, err)
	} else {
		t0 := time.Now()
		t1 := schedule.Next(t0)

		scheduler := cron.New()
		scheduler.Schedule(schedule, cron.FuncJob(func() {
			t0 := time.Now()

			if err := f(); err != nil {
				// TODO: break out of cron scheduler instead?
				log.Fatalf("%v", err)
			}

			t1 := time.Now()
			t2 := schedule.Next(t1)

			log.Printf("Schedule run completed in %v, next upgrade at: %v (in %v)", t1.Sub(t0), t2, t2.Sub(t1))
		}))

		log.Printf("Using --schedule=%#v, first upgrade at: %v (in %v)", options.Schedule, t1, (t1.Sub(t0)))

		scheduler.Run()

		return nil
	}
}
