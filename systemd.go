package main

import (
	"fmt"

	"github.com/coreos/go-systemd/dbus"
	godbus "github.com/godbus/dbus"
)

type systemdExec struct {
	conn *dbus.Conn
}

func SystemdExec(cmd []string) error {
	var se systemdExec
	var ch = make(chan string)
	var properties = []dbus.Property{
		dbus.PropExecStart(cmd, false), // XXX: bool flag meaning is inverted?
		dbus.PropType("oneshot"),
	}

	if conn, err := dbus.NewSystemConnection(); err != nil {
		return fmt.Errorf("dbus.NewSystemConnection: %v", err)
	} else {
		defer conn.Close()

		se.conn = conn
	}

	if err := se.conn.ResetFailedUnit("test.service"); err == nil {

	} else if dbusErr, ok := err.(godbus.Error); ok && dbusErr.Name == "org.freedesktop.systemd1.NoSuchUnit" {
		// ignore
	} else {
		return fmt.Errorf("dbus.ResetFailedUnit: %v", err)
	}

	if _, err := se.conn.StartTransientUnit("test.service", "replace", properties, ch); err != nil {
		return fmt.Errorf("dbus.StartTransientUnit: %v", err)
	}

	if status := <-ch; status != "done" {
		return fmt.Errorf("Job status: %v", status)
	}

	/*
		if properties, err := se.conn.GetUnitTypeProperties("test.service", "Service"); err != nil {
			return fmt.Errorf("dbus.GetUnitProperties: %v", err)
		} else {
			log.Printf("Unit properties:")
			for name, value := range properties {
				log.Printf("\t%v: %v", name, value)
			}
		}
	*/

	return nil
}
