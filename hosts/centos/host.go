package centos

import (
	"fmt"
	"log"
	"regexp"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/systemd"
)

const OperatingSystem = "CentOS"

var osPrettyNameRegexp = regexp.MustCompile(`CentOS Linux (.+?)( \(.+?\))?`)

type Host struct {
	configPath string
}

func (host *Host) Probe() (hosts.HostInfo, bool) {
	if hi, err := systemd.GetHostInfo(); err != nil {
		log.Printf("hosts/centos probe failed: %v", err)

		return hosts.HostInfo{}, false
	} else if match := osPrettyNameRegexp.FindStringSubmatch(hi.OperatingSystemPrettyName); match == nil {
		log.Printf("hosts/centos probe mismatch: %v", hi.OperatingSystemPrettyName)

		return hosts.HostInfo{}, false
	} else {
		log.Printf("hosts/centos probe success: %#v", hi)

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
	if exists, err := config.FileExists("yum-cron.conf"); err != nil {
		return err
	} else if !exists {
		log.Printf("hosts/centos: no yum-cron.conf configured")
	} else if configPath, err := config.CopyHostFile("yum-cron.conf"); err != nil {
		return fmt.Errorf("hosts/centos failed to CopyHostFile yum-cron.conf: %v", err)
	} else {
		log.Printf("hosts/centos: using copied yum-cron.conf at %v", configPath)

		host.configPath = configPath
	}

	return nil
}

func (host *Host) exec(cmd ...string) error {
	if err := systemd.Exec("host-upgrades", systemd.ExecOptions{Cmd: cmd}); err != nil {
		return fmt.Errorf("exec %v: %v", cmd, err)
	}

	return nil
}

func (host *Host) Upgrade() error {
	log.Printf("hosts/centos upgrade...")

	if host.configPath == "" {
		log.Printf("hosts/centos upgrade... (default config)")

		return host.exec("/usr/sbin/yum-cron")
	} else {
		log.Printf("hosts/centos upgrade... (with config)")

		return host.exec("/usr/sbin/yum-cron", host.configPath)
	}
}
