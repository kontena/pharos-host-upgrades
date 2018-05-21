package systemd

import (
	"fmt"

	"github.com/coreos/go-systemd/hostname1"
)

func setString(s *string, m map[string]interface{}, k string) {
	if v, ok := m[k]; !ok {

	} else if sv, ok := v.(string); !ok {

	} else {
		*s = sv
	}
}

type HostInfo struct {
	KernelName                string
	Hostname                  string
	OperatingSystemPrettyName string
	KernelVersion             string
	KernelRelease             string
}

func (i *HostInfo) fromProperties(properties map[string]interface{}) {
	setString(&i.KernelName, properties, "KernelName")
	setString(&i.Hostname, properties, "Hostname")
	setString(&i.OperatingSystemPrettyName, properties, "OperatingSystemPrettyName")
	setString(&i.KernelVersion, properties, "KernelVersion")
	setString(&i.KernelRelease, properties, "KernelRelease")
}

func GetHostInfo() (HostInfo, error) {
	var hostInfo HostInfo

	conn, err := hostname1.New()
	if err != nil {
		return hostInfo, fmt.Errorf("hostname1.New: %v", err)
	} else {
		defer conn.Close()
	}

	if properties, err := conn.GetProperties(); err != nil {
		return hostInfo, fmt.Errorf("hostname1.GetProperties: %v", err)
	} else {
		hostInfo.fromProperties(properties)
	}

	return hostInfo, nil
}
