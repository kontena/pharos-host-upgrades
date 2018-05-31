package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const DefaultRebootTimeout = 5 * time.Minute
const DefaultScheduleWindow = 1 * time.Hour

type Options struct {
	ConfigPath     string
	HostMount      string
	Schedule       string
	ScheduleWindow time.Duration
	Reboot         bool
	RebootTimeout  time.Duration
	Drain          bool
	Kube           KubeOptions
}

func run(options Options) error {
	config, err := loadConfig(options)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	host, hostInfo, err := probeHost(options)
	if err != nil {
		return fmt.Errorf("Failed to probe host: %v", err)
	}

	if err := host.Config(config); err != nil {
		return fmt.Errorf("Failed to configure host: %v", err)
	}

	scheduler, err := makeScheduler(options)
	if err != nil {
		return err
	}

	// kube is optional, this will be nil if not configured; the methods are no-op when called on nil
	kube, err := makeKube(options, hostInfo)
	if err != nil {
		return fmt.Errorf("Failed to initialize kube: %v", err)
	}

	if options.Reboot && options.Drain {
		log.Printf("Using --reboot --drain, will drain kube node and reboot host after upgrades if required")
	} else if options.Reboot {
		log.Printf("Using --reboot, will reboot host after upgrades if required")
	} else {
		log.Printf("Skipping host reboot after upgrades")
	}

	return scheduler.Run(func(ctx context.Context) error {
		if err := kube.AcquireLock(ctx); err != nil {
			return fmt.Errorf("Failed to acquire kube lock: %v", err)
		}

		// runs with the kube lock held
		rebooting, err := func() (bool, error) {
			log.Printf("Running host upgrades...")

			status, err := host.Upgrade()

			if err != nil {
				kube.UpdateHostStatus(status, err)

				return false, err
			}

			if err := kube.UpdateHostStatus(status, err); err != nil {
				return false, fmt.Errorf("Kube node status update failed: %v", err)
			}

			if options.Reboot && status.RebootRequired {
				if !options.Drain {
					log.Printf("Reboot required, rebooting without draining kube node...")
				} else {
					log.Printf("Reboot required, draining kube node...")

					if err := kube.DrainNode(); err != nil {
						// XXX: bad idea to release the lock with the node drained?
						return false, fmt.Errorf("Failed to drain kube node for host reboot: %v", err)
					} else if err := kube.MarkReboot(time.Now()); err != nil {
						// XXX: bad idea to release the lock with the node drained?
						return false, fmt.Errorf("Failed to mark kube node for host reboot: %v", err)
					}

					log.Printf("Rebooting...")
				}

				if err := host.Reboot(); err != nil {
					// XXX: will enter a fail-loop while drained after restarting due to kube node reboot state if unable to reboot
					// XXX: bad idea to release the lock with the node drained?
					return false, fmt.Errorf("Failed to reboot host: %v", err)
				}

				log.Printf("Host is shutting down...")

				return true, nil // do not release kube lock

			} else if status.RebootRequired {
				log.Printf("Reboot required, but skipping without --reboot")

			} else {
				log.Printf("No reboot required")
			}

			return false, nil
		}()

		// either release lock, or wait for reboot to happen
		if err != nil {
			log.Printf("Upgrade failed, releasing kube lock... (%v)", err)

			if lockErr := kube.ReleaseLock(); lockErr != nil {
				log.Printf("Failed to release kube lock: %v", lockErr)
			}

			return err

		} else if rebooting {
			log.Printf("Leaving kube lock held for reboot, waiting for termination...")

			// wait for systemd shutdown => docker terminate to kill us
			time.Sleep(options.RebootTimeout)

			return fmt.Errorf("Timeout waiting for host to shutdown")

		} else if err := kube.ReleaseLock(); err != nil {
			return fmt.Errorf("Failed to release kube lock: %v", err)

		} else {
			log.Printf("Done")
		}

		return nil
	})
}

func main() {
	var options Options

	flag.StringVar(&options.ConfigPath, "config-path", "/etc/host-upgrades", "Path to configmap dir")
	flag.StringVar(&options.HostMount, "host-mount", "/run/host-upgrades", "Path to shared mount with host. Must be under /run to reset when rebooting!")
	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.DurationVar(&options.ScheduleWindow, "schedule-window", DefaultScheduleWindow, "Set a deadline for the scheduled upgrade to start (duration syntax)")
	flag.BoolVar(&options.Reboot, "reboot", false, "Reboot if required")
	flag.DurationVar(&options.RebootTimeout, "reboot-timeout", DefaultRebootTimeout, "Wait for system to shutdown when rebooting")
	flag.BoolVar(&options.Drain, "drain", false, "Drain kube node before reboot, uncordon after reboot")

	flag.StringVar(&options.Kube.Namespace, "kube-namespace", os.Getenv("KUBE_NAMESPACE"), "Name of kube Namespace (KUBE_NAMESPACE)")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", os.Getenv("KUBE_DAEMONSET"), "Name of kube DaemonSet (KUBE_DAEMONSET)")
	flag.StringVar(&options.Kube.Node, "kube-node", os.Getenv("KUBE_NODE"), "Name of kube Node (KUBE_NODE)")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
