package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const DefaultRebootTimeout = 5 * time.Minute

type Options struct {
	ConfigPath    string
	HostMount     string
	Schedule      string
	Reboot        bool
	RebootTimeout time.Duration
	Kube          KubeOptions
}

func run(options Options) error {
	config, err := loadConfig(options)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	host, err := probeHost(options)
	if err != nil {
		return fmt.Errorf("Failed to probe host: %v", err)
	}

	if err := host.Config(config); err != nil {
		return fmt.Errorf("Failed to configure host: %v", err)
	}

	kube, err := makeKube(options, host)
	if err != nil {
		return fmt.Errorf("Failed to connect to kube: %v", err)
	}

	scheduler, err := makeScheduler(options)
	if err != nil {
		return err
	}

	if options.Reboot {
		log.Printf("Using --reboot, will reboot host after upgrades if required")
	}

	return scheduler.Run(func() error {
		if err := kube.AcquireLock(); err != nil {
			return fmt.Errorf("Failed to acquire kube lock: %v", err)
		}

		rebooting, upgradeErr := func() (bool, error) {
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
				log.Printf("Rebooting host...")

				if err := host.Reboot(); err != nil {
					return false, fmt.Errorf("Failed to reboot host: %v", err)
				}

				log.Printf("Host is shutting down...")

				return true, nil // do not release kube lock

			} else if status.RebootRequired {
				log.Printf("Skipping host reboot...")
			}

			return false, nil
		}()

		if rebooting {
			log.Printf("Leaving kube lock held for reboot...")
		} else if err := kube.ReleaseLock(); err != nil {
			if upgradeErr == nil {
				return fmt.Errorf("Failed to release kube lock: %v", err)
			} else {
				log.Printf("Failed to release kube lock: %v", err)
			}
		} else {
			log.Printf("Released kube lock")
		}

		return upgradeErr
	})
}

func main() {
	var options Options

	flag.StringVar(&options.ConfigPath, "config-path", "/etc/host-upgrades", "Path to configmap dir")
	flag.StringVar(&options.HostMount, "host-mount", "/run/host-upgrades", "Path to shared mount with host. Must be under /run to reset when rebooting!")
	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.BoolVar(&options.Reboot, "reboot", false, "Reboot if required")
	flag.DurationVar(&options.RebootTimeout, "reboot-timeout", DefaultRebootTimeout, "Wait for system to shutdown when rebooting")
	flag.StringVar(&options.Kube.Namespace, "kube-namespace", os.Getenv("KUBE_NAMESPACE"), "Name of kube Namespace (KUBE_NAMESPACE)")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", os.Getenv("KUBE_DAEMONSET"), "Name of kube DaemonSet (KUBE_DAEMONSET)")
	flag.StringVar(&options.Kube.Node, "kube-node", os.Getenv("KUBE_NODE"), "Name of kube Node (KUBE_NODE)")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
