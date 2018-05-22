package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/hosts/ubuntu"
)

func probeHost(options Options) (hosts.Host, error) {
	var hosts = []hosts.Host{
		ubuntu.Host{},
	}

	for _, host := range hosts {
		if hostInfo, ok := host.Probe(); !ok {
			continue
		} else {
			log.Printf("Probed host: %#v", hostInfo)

			return host, nil
		}
	}

	return nil, fmt.Errorf("No hosts matched")
}
