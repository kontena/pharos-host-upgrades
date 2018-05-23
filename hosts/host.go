package hosts

import (
	"time"
)

type HostInfo struct {
	OperatingSystem        string
	OperatingSystemRelease string
	Kernel                 string
	KernelRelease          string
}

type Status struct {
	RebootRequired        bool
	RebootRequiredSince   time.Time
	RebootRequiredMessage string

	UpgradeLog string
}

type Host interface {
	Probe() (HostInfo, bool)
	Config(Config) error
	Upgrade() (Status, error)
}
