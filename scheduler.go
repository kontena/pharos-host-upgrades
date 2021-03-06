package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron"
)

type Scheduler struct {
	option   string
	schedule cron.Schedule
	window   time.Duration
	ch       chan time.Time
}

func makeScheduler(options Options) (Scheduler, error) {
	var scheduler = Scheduler{
		option: options.Schedule,
		window: options.ScheduleWindow,
		ch:     make(chan time.Time),
	}

	if options.Schedule == "" {
		log.Printf("No --schedule given, will run once")

		return scheduler, nil
	} else if schedule, err := cron.ParseStandard(options.Schedule); err != nil {
		return scheduler, fmt.Errorf("Invalid --schedule=%v: %v", options.Schedule, err)
	} else {
		scheduler.schedule = schedule
	}

	return scheduler, nil
}

func (scheduler *Scheduler) run(f func(ctx context.Context) error) {
	initTime := time.Now()
	nextTime := scheduler.schedule.Next(initTime)

	log.Printf("Using --schedule=%#v, first upgrade at: %v (in %v)", scheduler.option, nextTime, nextTime.Sub(initTime))

	if scheduler.window != 0 {
		log.Printf("Using --schedule-window=%v", scheduler.window)
	}

	for startTime := range scheduler.ch {
		func() {
			ctx := context.Background()

			if scheduler.window != 0 {
				deadline := startTime.Add(scheduler.window)
				deadlineCtx, cancelCtx := context.WithDeadline(ctx, deadline)

				log.Printf("Schedule run started, deadline at %v", deadline)

				ctx = deadlineCtx
				defer cancelCtx()
			} else {
				log.Printf("Schedule run started, no deadline")

			}

			if err := f(ctx); err != nil {
				// TODO: break out of scheduler loop instead?
				log.Fatalf("%v", err)
			}
		}()

		endTime := time.Now()
		nextTime := scheduler.schedule.Next(endTime)

		log.Printf("Schedule run completed in %v, next upgrade at: %v (in %v)", endTime.Sub(startTime), nextTime, nextTime.Sub(endTime))
	}
}

func (scheduler *Scheduler) trigger() {
	select {
	case scheduler.ch <- time.Now():
		return
	default:
		log.Printf("Scheduler is busy, skipping scheduled run")
	}
}

func (scheduler Scheduler) Run(f func(ctx context.Context) error) error {
	if scheduler.schedule == nil {
		return f(context.Background())
	} else {
		defer close(scheduler.ch)
		go scheduler.run(f)
	}

	c := cron.New()
	c.Schedule(scheduler.schedule, cron.FuncJob(func() {
		scheduler.trigger()
	}))
	c.Run()

	return nil
}
