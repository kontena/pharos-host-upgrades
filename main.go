package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

	kube, err := makeKube(options)
	if err != nil {
		return fmt.Errorf("Failed to connect to kube: %v", err)
	}

	return runSchedule(options, func() error {
		return kube.withLock(func() error {
			log.Printf("Running host upgrades...")

			return host.Upgrade()
		})
	})
}

func main() {
	var options Options

	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.StringVar(&options.Kube.Namespace, "kube-namespace", os.Getenv("KUBE_NAMESPACE"), "Name of kube Namespace")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", os.Getenv("KUBE_DAEMONSET"), "Name of kube DaemonSet")
	flag.StringVar(&options.Kube.Node, "kube-node", os.Getenv("KUBE_NODE"), "Name of kube Node")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
