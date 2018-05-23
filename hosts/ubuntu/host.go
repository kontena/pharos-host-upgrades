package ubuntu

import (
	"fmt"
	"log"
	"regexp"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/systemd"
)

const OperatingSystem = "Ubuntu"

var osPrettyNameRegexp = regexp.MustCompile(`Ubuntu (\S+)( LTS)?`)

type aptConfVars struct {
	ConfigPath string
}

// because of how lists in apt configs are merged, the unattended-upgrades.conf must be loaded last to allow overriding the Unattended-Upgrade::Allowed-Origins list
// use a generated APT_CONFIG=... file to load the given unattended-upgrades.conf using Dir::Etc::main after the Dir::Etc::Parts
// this assumes that /etc/apt.conf does not exist, or does not contain anything important...
const aptConfTemplate = `
Dir::Etc::main "{{.ConfigPath}}";
`

type Host struct {
	configPath    string
	aptConfigPath string
}

func (host *Host) Probe() (hosts.HostInfo, bool) {
	if hi, err := systemd.GetHostInfo(); err != nil {
		log.Printf("hosts/ubuntu probe failed: %v", err)

		return hosts.HostInfo{}, false
	} else if match := osPrettyNameRegexp.FindStringSubmatch(hi.OperatingSystemPrettyName); match == nil {
		log.Printf("hosts/ubuntu probe mismatch: %v", hi.OperatingSystemPrettyName)

		return hosts.HostInfo{}, false
	} else {
		log.Printf("hosts/ubuntu probe success: %#v", hi)

		var hostInfo = hosts.HostInfo{
			OperatingSystem:        OperatingSystem,
			OperatingSystemRelease: match[1],
			Kernel:                 hi.KernelName,
			KernelRelease:          hi.KernelRelease,
		}

		return hostInfo, true
	}
}

func (host *Host) Config(config hosts.Config) error {
	if exists, err := config.FileExists("unattended-upgrades.conf"); err != nil {
		return err
	} else if !exists {
		log.Printf("hosts/ubuntu: no unattended-upgrades.conf configured")
	} else if configPath, err := config.CopyHostFile("unattended-upgrades.conf"); err != nil {
		return fmt.Errorf("hosts/ubuntu failed to CopyHostFile unattended-upgrades.conf: %v", err)
	} else {
		log.Printf("hosts/ubuntu: using copied unattended-upgrades.conf at %v", configPath)

		host.configPath = configPath
	}

	if host.configPath == "" {

	} else if path, err := config.GenerateFile("apt.conf", aptConfTemplate, aptConfVars{ConfigPath: host.configPath}); err != nil {
		return fmt.Errorf("hosts/ubuntu failed to GenerateFile apt.conf: %v", err)
	} else {
		host.aptConfigPath = path
	}

	return nil
}

func (host *Host) exec(name string, cmd ...string) error {
	if err := systemd.Exec(name, systemd.ExecOptions{Cmd: cmd}); err != nil {
		return fmt.Errorf("exec %v(%v): %v", name, cmd, err)
	}

	return nil
}

func (host *Host) execEnv(name string, env []string, cmd ...string) error {
	if err := systemd.Exec(name, systemd.ExecOptions{Env: env, Cmd: cmd}); err != nil {
		return fmt.Errorf("exec %v(%v): %v", name, cmd, err)
	}

	return nil
}

func (host *Host) Upgrade() error {
	log.Printf("hosts/ubuntu upgrade...")

	if err := host.exec("host-upgrades-update", "/usr/bin/apt-get", "update"); err != nil {
		return err
	}

	if host.configPath == "" {
		return host.exec("host-upgrades", "/usr/bin/unattended-upgrade", "-v")
	} else {
		return host.execEnv("host-upgrades", []string{"APT_CONFIG=" + host.aptConfigPath}, "/usr/bin/unattended-upgrade", "-v")
	}

	return nil
}
