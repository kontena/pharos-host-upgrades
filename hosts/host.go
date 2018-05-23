package hosts

type HostInfo struct {
	OperatingSystem        string
	OperatingSystemRelease string
	Kernel                 string
	KernelRelease          string
}

type Host interface {
	Probe() (HostInfo, bool)
	Config(Config) error
	Upgrade() error
}
