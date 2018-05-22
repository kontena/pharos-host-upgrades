package systemd

import (
	"fmt"
	"log"

	"github.com/coreos/go-systemd/dbus"
	godbus "github.com/godbus/dbus"
)

type systemdExec struct {
	unit      string
	cmd       []string
	conn      *dbus.Conn
	ch        chan string
	jobStatus string
}

func (se *systemdExec) connect() error {
	if conn, err := dbus.NewSystemConnection(); err != nil {
		return fmt.Errorf("dbus.NewSystemConnection: %v", err)
	} else {
		se.conn = conn
	}

	return nil
}

func (se *systemdExec) reset() error {
	log.Printf("systemd/exec %v: reset", se.unit)

	if err := se.conn.ResetFailedUnit(se.unit); err == nil {

	} else if dbusErr, ok := err.(godbus.Error); ok && dbusErr.Name == "org.freedesktop.systemd1.NoSuchUnit" {
		return nil
	} else {
		return fmt.Errorf("dbus.ResetFailedUnit %v: %v", se.unit, err)
	}

	return nil
}

// fails if unit is already running
func (se *systemdExec) start() error {
	var properties = []dbus.Property{
		dbus.PropExecStart(se.cmd, false), // XXX: bool flag meaning is inverted?
		dbus.PropType("oneshot"),
	}

	log.Printf("systemd/exec %v: start %#v", se.unit, properties)

	if _, err := se.conn.StartTransientUnit(se.unit, "fail", properties, se.ch); err != nil {
		return fmt.Errorf("dbus.StartTransientUnit %v: %v", se.unit, err)
	}

	return nil
}

func (se *systemdExec) wait() error {
	log.Printf("systemd/exec %v: wait", se.unit)

	se.jobStatus = <-se.ch

	return nil
}

func (se *systemdExec) show() error {
	if properties, err := se.conn.GetUnitTypeProperties(se.unit, "Service"); err != nil {
		return fmt.Errorf("dbus.GetUnitProperties %v: %v", se.unit, err)
	} else {
		log.Printf("Unit properties:")
		for name, value := range properties {
			log.Printf("\t%v: %v", name, value)
		}
	}

	return nil
}

func (se *systemdExec) close() {
	se.conn.Close()
}

func Exec(name string, cmd []string) error {
	var se = systemdExec{
		unit: fmt.Sprintf("%v.service", name),
		cmd:  cmd,
		ch:   make(chan string),
	}

	log.Printf("systemd/exec %v: cmd=%v", se.unit, cmd)

	if err := se.connect(); err != nil {
		return err
	} else {
		defer se.close()
	}

	if err := se.reset(); err != nil {
		return err
	}

	if err := se.start(); err != nil {
		return err
	}

	if err := se.wait(); err != nil {
		return err
	}

	if se.jobStatus == "done" {
		log.Printf("systemd/exec %v: done", se.unit)
	} else {
		return fmt.Errorf("Job %v", se.jobStatus)
	}

	return nil
}
