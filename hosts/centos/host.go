package centos

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/kontena/pharos-host-upgrades/hosts"
	"github.com/kontena/pharos-host-upgrades/proc"
	"github.com/kontena/pharos-host-upgrades/systemd"
)

const OperatingSystem = "CentOS"

var osPrettyNameRegexp = regexp.MustCompile(`CentOS Linux (.+?)( \(.+?\))?`)

const upgradeScript = `
set -ue

yum-cron ${CONFIG_PATH:-} | tee $HOST_PATH/yum-cron.out

needs-restarting -r > $HOST_PATH/needs-restarting.out || touch -a $HOST_PATH/needs-restarting.stamp
`

type Host struct {
	info   hosts.Info
	config hosts.Config

	configPath string
	scriptPath string
}

func (host *Host) Probe() bool {
	if hi, err := systemd.GetHostInfo(); err != nil {
		log.Printf("hosts/centos probe failed: %v", err)

		return false
	} else if match := osPrettyNameRegexp.FindStringSubmatch(hi.OperatingSystemPrettyName); match == nil {
		log.Printf("hosts/centos probe mismatch: %v", hi.OperatingSystemPrettyName)

		return false
	} else {
		host.info = hosts.Info{
			OperatingSystem:        OperatingSystem,
			OperatingSystemRelease: match[1],
			Kernel:                 hi.KernelName,
			KernelRelease:          hi.KernelRelease,
		}

		if procStat, err := proc.ReadStat(); err != nil {
			log.Printf("hosts/centos failed stat BootTime: %v", err)
		} else {
			log.Printf("hosts/centos boot time: %v", procStat.BootTime)

			host.info.BootTime = procStat.BootTime
		}

		log.Printf("hosts/centos probe success: %#v", host.info)

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
	host.config = config

	if hostPath := config.HostPath(); hostPath == "" {
		return fmt.Errorf("hosts/centos requires --host-path")
	} else {
		log.Printf("hosts/centos: using host path %v for output files", hostPath)
	}

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

	if path, err := config.WriteHostFile("host-upgrades.sh", bytes.NewReader([]byte(upgradeScript)), hosts.FileModeScript); err != nil {
		return err
	} else {
		log.Printf("hosts/centos: using generated host-upgrades.sh at %v", path)

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

func (host *Host) readNeedsRestarting(status *hosts.Status) error {
	var buf bytes.Buffer

	if stat, exists, err := host.config.StatHostFile("needs-restarting.stamp"); err != nil {
		return err
	} else if !exists {

	} else if err := host.config.ReadHostFile("needs-restarting.out", &buf); err != nil {
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

	if err := host.config.ReadHostFile("yum-cron.out", &buf); err != nil {
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
		"CONFIG_PATH=" + host.configPath,
	}
	var cmd = []string{"/bin/sh", "-x", host.scriptPath}

	log.Printf("hosts/centos upgrade...")

	if err := host.exec(env, cmd); err != nil {
		return status, err
	} else if err := host.readUpgradeLog(&status); err != nil {
		return status, err
	} else if err := host.readNeedsRestarting(&status); err != nil {
		return status, err
	} else {
		return status, nil
	}
}
