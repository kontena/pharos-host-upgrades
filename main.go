package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Options struct {
	ConfigPath string
	HostMount  string
	Schedule   string
	Kube       KubeOptions
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

	kube, err := makeKube(options)
	if err != nil {
		return fmt.Errorf("Failed to connect to kube: %v", err)
	}

	scheduler, err := makeScheduler(options)
	if err != nil {
		return err
	}

	return scheduler.Run(func() error {
		return kube.withLock(func() error {
			log.Printf("Running host upgrades...")

			return host.Upgrade()
		})
	})
}

func main() {
	var options Options

	flag.StringVar(&options.ConfigPath, "config-path", "/etc/host-upgrades", "Path to configmap dir")
	flag.StringVar(&options.HostMount, "host-mount", "/run/host-upgrades", "Path to host mount")
	flag.StringVar(&options.Schedule, "schedule", "", "Scheduled upgrade (cron syntax)")
	flag.StringVar(&options.Kube.Namespace, "kube-namespace", os.Getenv("KUBE_NAMESPACE"), "Name of kube Namespace (KUBE_NAMESPACE)")
	flag.StringVar(&options.Kube.DaemonSet, "kube-daemonset", os.Getenv("KUBE_DAEMONSET"), "Name of kube DaemonSet (KUBE_DAEMONSET)")
	flag.StringVar(&options.Kube.Node, "kube-node", os.Getenv("KUBE_NODE"), "Name of kube Node (KUBE_NODE)")
	flag.Parse()

	if err := run(options); err != nil {
		log.Fatalf("%v", err)
	}
}
