package ubuntu

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/proc"
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

const upgradeScript = `
set -ue

apt-get update

unattended-upgrade -v > $HOST_PATH/unattended-upgrade.out

if [ -e /run/reboot-required ]; then
	# preserve timestamp
	cp -a /run/reboot-required $HOST_PATH/reboot-required
fi

# TODO: also restart for service upgrades?
# which needrestart && needrestart -b > $HOST_PATH/needrestart
`

type Host struct {
	info   hosts.Info
	config hosts.Config

	configPath    string
	aptConfigPath string
	scriptPath    string
}

func (host *Host) Probe() bool {
	if hi, err := systemd.GetHostInfo(); err != nil {
		log.Printf("hosts/ubuntu probe failed: %v", err)

		return false
	} else if match := osPrettyNameRegexp.FindStringSubmatch(hi.OperatingSystemPrettyName); match == nil {
		log.Printf("hosts/ubuntu probe mismatch: %v", hi.OperatingSystemPrettyName)

		return false
	} else {
		host.info = hosts.Info{
			OperatingSystem:        OperatingSystem,
			OperatingSystemRelease: match[1],
			Kernel:                 hi.KernelName,
			KernelRelease:          hi.KernelRelease,
		}

		if procStat, err := proc.ReadStat(); err != nil {
			log.Printf("hosts/ubuntu failed stat BootTime: %v", err)
		} else {
			log.Printf("hosts/ubuntu boot time: %v", procStat.BootTime)

			host.info.BootTime = procStat.BootTime
		}

		log.Printf("hosts/ubuntu probe success: %#v", host.info)

		return true
	}
}

func (host *Host) String() string {
	return fmt.Sprintf("%v %v", host.info.OperatingSystem, host.info.OperatingSystemRelease)
}

func (host *Host) Info() hosts.Info {
	return host.info
}

func (host *Host) Config(config hosts.Config) error {
	host.config = config // used for reading output...

	if hostPath := config.HostPath(); hostPath == "" {
		return fmt.Errorf("hosts/ubuntu requires --host-path")
	} else {
		log.Printf("hosts/ubuntu: using host path %v for output files", hostPath)
	}

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

	// Ubuntu has /run mounted noexec, so no point making this executable...
	if path, err := config.WriteHostFile("host-upgrades.sh", bytes.NewReader([]byte(upgradeScript))); err != nil {
		return err
	} else {
		log.Printf("hosts/ubuntu: using generated host-upgrades.sh at %v", path)

		host.scriptPath = path
	}

	return nil
}

func (host *Host) exec(env []string, cmd []string) error {
	if err := systemd.Exec("host-upgrades", systemd.ExecOptions{Env: env, Cmd: cmd}); err != nil {
		return fmt.Errorf("exec %v: %v", cmd, err)
	}

	return nil
}

func (host *Host) readRebootRequired(status *hosts.Status) error {
	var buf bytes.Buffer

	if stat, exists, err := host.config.StatHostFile("reboot-required"); err != nil {
		return err
	} else if !exists {

	} else if err := host.config.ReadHostFile("reboot-required", &buf); err != nil {
		return err
	} else {
		status.RebootRequired = true
		status.RebootRequiredSince = stat.ModTime()
		status.RebootRequiredMessage = buf.String()
	}

	return nil
}

func (host *Host) readUpgradeLog(status *hosts.Status) error {
	var buf bytes.Buffer

	if err := host.config.ReadHostFile("unattended-upgrade.out", &buf); err != nil {
		return err
	} else {
		status.UpgradeLog = buf.String()
	}

	return nil
}

func (host *Host) Upgrade() (hosts.Status, error) {
	var status hosts.Status
	var env = []string{
		"HOST_PATH=" + host.config.HostPath(),
		"APT_CONFIG=" + host.aptConfigPath,
	}
	var cmd = []string{"/bin/sh", "-x", host.scriptPath}

	log.Printf("hosts/ubuntu upgrade...")

	if err := host.exec(env, cmd); err != nil {
		return status, err
	} else if err := host.readUpgradeLog(&status); err != nil {
		return status, err
	} else if err := host.readRebootRequired(&status); err != nil {
		return status, err
	} else {
		return status, nil
	}
}
