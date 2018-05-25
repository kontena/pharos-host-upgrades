package systemd

import (
	"fmt"

	"github.com/coreos/go-systemd/login1"
)

func Reboot() error {
	conn, err := login1.New()

	if err != nil {
		return fmt.Errorf("login1.New: %v", err)
	} else {
		defer conn.Close()
	}

	// XXX: WTF... no error return?
	conn.Reboot(false)

	return nil
}
