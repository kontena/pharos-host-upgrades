package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/hosts/centos"
	"github.com/kontena/pharos-host-upgrades/hosts/ubuntu"
)

func probeHost(options Options) (hosts.Host, hosts.Info, error) {
	var probeHosts = []hosts.Host{
		&ubuntu.Host{},
		&centos.Host{},
	}

	for _, host := range probeHosts {
		if info, ok := host.Probe(); !ok {
			continue
		} else {
			log.Printf("Probed host: %v", host)

			return host, info, nil
		}
	}

	return nil, hosts.Info{}, fmt.Errorf("No hosts matched")
}
