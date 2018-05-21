package main

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron"
)

type Runnable interface {
	Run()
}

func scheduledRunner(options Options, f func()) (func(), error) {
	if options.Schedule == "" {
		log.Printf("No --schedule given, will run once")

		return f, nil
	} else if schedule, err := cron.Parse(options.Schedule); err != nil {
		return nil, fmt.Errorf("Invalid --schedule=%v: %v", options.Schedule, err)
	} else {
		t0 := time.Now()
		t1 := schedule.Next(t0)

		log.Printf("Using --schedule=%#v, first upgrade at: %v (in %v)", options.Schedule, t1, (t1.Sub(t0)))

		scheduler := cron.New()
		scheduler.Schedule(schedule, cron.FuncJob(f))

		return func() { scheduler.Run() }, nil
	}
}
