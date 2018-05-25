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
		return kube.WithLock(func() error {
			log.Printf("Running host upgrades...")

			status, err := host.Upgrade()

			if err != nil {
				kube.UpdateHostStatus(status, err)

				return err
			}

			if err := kube.UpdateHostStatus(status, err); err != nil {
				return fmt.Errorf("Kube node status update failed: %v", err)
			}

			if options.Reboot && status.RebootRequired {
				log.Printf("Rebooting host...")

				if err := host.Reboot(); err != nil {
					return fmt.Errorf("Failed to reboot host: %v", err)
				}

				log.Printf("Waiting for host shutdown...")

				// wait for reboot to happen... systemd will kill us, leaving the lock acquired
				// XXX: this is up to implementation details: defer in goroutine does not execute when process exits
				time.Sleep(options.RebootTimeout)

				return fmt.Errorf("Timeout waiting for host to shutdown")

			} else if status.RebootRequired {
				log.Printf("Skipping host reboot...")
			}

			return nil
		})
	})
}

func main() {
	var options = Options{
		RebootTimeout: DefaultRebootTimeout,
	}

	flag.StringVar(&options.ConfigPath, "config-path", "/etc/host-upgrades", "Path to configmap dir")
	flag.StringVar(&options.HostMount, "host-mount", "/run/host-upgrades", "Path to shared mount with host. Must be under /run to reset when rebooting!")
	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.BoolVar(&options.Reboot, "reboot", false, "Reboot if required")
	flag.StringVar(&options.Kube.Namespace, "kube-namespace", os.Getenv("KUBE_NAMESPACE"), "Name of kube Namespace (KUBE_NAMESPACE)")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", os.Getenv("KUBE_DAEMONSET"), "Name of kube DaemonSet (KUBE_DAEMONSET)")
	flag.StringVar(&options.Kube.Node, "kube-node", os.Getenv("KUBE_NODE"), "Name of kube Node (KUBE_NODE)")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
