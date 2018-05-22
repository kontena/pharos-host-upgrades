package main

import (
	"flag"
	"fmt"
	"log"
)

type Options struct {
	Schedule string
	Kube     KubeOptions
}

func run(options Options) error {
	host, err := probeHost(options)
	if err != nil {
		return fmt.Errorf("Failed to probe host: %v", err)
	}

	kube, err := newKube(options)
	if err != nil {
		return fmt.Errorf("Failed to connect to kube: %v", err)
	}

	return runSchedule(options, func() error {
		log.Printf("Running with kube lock...")

		return withKubeLock(kube, func() error {
			log.Printf("Running host upgrades...")

			return host.Upgrade()
		})
	})
}

func main() {
	var options Options

	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.StringVar(&options.Kube.Namespace, "kube-namespace", "kube-system", "Name of kube Namespace")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", "pharos-host-upgrades", "Name of kube DaemonSet")
	flag.StringVar(&options.Kube.DaemonSet, "kube-node", "", "Name of kube Node")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
