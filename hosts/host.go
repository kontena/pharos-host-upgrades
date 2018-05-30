package hosts

import (
	"time"
)

// Static host information, determined during probe, not expected to change without restarting
type Info struct {
	OperatingSystem        string
	OperatingSystemRelease string
	Kernel                 string
	KernelRelease          string
	BootTime               time.Time
}

type Status struct {
	RebootRequired        bool
	RebootRequiredSince   time.Time
	RebootRequiredMessage string

	UpgradeLog string
}

type Host interface {
	Probe() (Info, bool)
	Config(Config) error
	Upgrade() (Status, error)
	Reboot() error
}
