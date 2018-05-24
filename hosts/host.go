package hosts

import (
	"time"
)

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
	Probe() bool
	Info() Info
	Config(Config) error
	Upgrade() (Status, error)
}
