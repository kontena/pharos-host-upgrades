package main

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron"
)

type Scheduler struct {
	option   string
	schedule cron.Schedule
	cron     *cron.Cron
}

func makeScheduler(options Options) (Scheduler, error) {
	var scheduler = Scheduler{
		option: options.Schedule,
	}

	if options.Schedule == "" {
		log.Printf("No --schedule given, will run once")

		return scheduler, nil
	} else if schedule, err := cron.Parse(options.Schedule); err != nil {
		return scheduler, fmt.Errorf("Invalid --schedule=%v: %v", options.Schedule, err)
	} else {
		scheduler.schedule = schedule
	}

	return scheduler, nil
}

func (scheduler Scheduler) run(f func() error) error {
	if scheduler.schedule == nil {
		return f()
	}

	t0 := time.Now()
	t1 := scheduler.schedule.Next(t0)

	scheduler.cron = cron.New()
	scheduler.cron.Schedule(scheduler.schedule, cron.FuncJob(func() {
		t0 := time.Now()

		if err := f(); err != nil {
			// TODO: break out of cron scheduler instead?
			log.Fatalf("%v", err)
		}

		t1 := time.Now()
		t2 := scheduler.schedule.Next(t1)

		log.Printf("Schedule run completed in %v, next upgrade at: %v (in %v)", t1.Sub(t0), t2, t2.Sub(t1))
	}))

	log.Printf("Using --schedule=%#v, first upgrade at: %v (in %v)", scheduler.option, t1, (t1.Sub(t0)))

	scheduler.cron.Run()

	return nil
}
