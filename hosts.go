package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/hosts/centos"
	"github.com/kontena/pharos-host-upgrades/hosts/ubuntu"
)

func probeHost(options Options) (hosts.Host, error) {
	var hosts = []hosts.Host{
		&ubuntu.Host{},
		&centos.Host{},
	}

	for _, host := range hosts {
		if ok := host.Probe(); !ok {
			continue
		} else {
			log.Printf("Probed host: %v", host)

			return host, nil
		}
	}

	return nil, fmt.Errorf("No hosts matched")
}
