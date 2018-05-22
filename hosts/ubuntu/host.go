package ubuntu

import (
	"fmt"
	"log"
	"regexp"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/systemd"
)

const OperatingSystem = "Ubuntu"

var upgradeCmd = []string{"/usr/bin/unattended-upgrade", "-v"}
var osPrettyNameRegexp = regexp.MustCompile(`Ubuntu (\S+)( LTS)?`)

type Host struct {
}

func (host Host) Probe() (hosts.HostInfo, bool) {
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

func (host Host) exec(cmd []string) error {
	if err := systemd.Exec("host-upgrades", cmd); err != nil {
		return fmt.Errorf("exec %v: %v", cmd, err)
	}

	return nil
}

func (host Host) Upgrade() error {
	log.Printf("hosts/ubuntu upgrade: %v", upgradeCmd)

	if err := host.exec(upgradeCmd); err != nil {
		return err
	}

	return nil
}
