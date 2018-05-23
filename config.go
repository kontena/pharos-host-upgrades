package main

import (
	"fmt"
	"log"

	"github.com/kontena/pharos-host-upgrades/hosts"
)

func loadConfig(options Options) (hosts.Config, error) {
	var config hosts.Config

	if path := options.ConfigPath; path == "" {
		return config, nil
	} else if err := config.Load(path); err != nil {
		return config, fmt.Errorf("Failed to load --config-path=%v: %v", path, err)
	} else {
		log.Printf("Load config from --config-path=%v", path)
	}

	return config, nil
}
