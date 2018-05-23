package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
)

func loadConfig(options Options) (hosts.Config, error) {
	var config hosts.Config

	if path := options.ConfigPath; path == "" {

	} else if exists, err := config.UsePath(path); err != nil {
		return config, fmt.Errorf("Invalid --config-path=%v: %v", path, err)
	} else if !exists {
		return config, fmt.Errorf("Skipping non-existing --config-path=%v", path)
	} else {
		log.Printf("Load config from --config-path=%v", path)
	}

	if path := options.HostMount; path == "" {

	} else if exists, err := config.UseMount(path); err != nil {
		return config, fmt.Errorf("Invalid --host-mount=%v: %v", path, err)
	} else if !exists {
		return config, fmt.Errorf("Skipping non-existing --host-mount=%v", path)
	} else {
		log.Printf("Copying configs to --host-mount=%v", path)
	}

	return config, nil
}
