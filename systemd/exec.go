package systemd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/sdjournal"
	godbus "github.com/godbus/dbus"
)

type ExecOptions struct {
	Unit string // set to <name>.service
	Cmd  []string
	Env  []string
}

type ExecResult struct {
	Status   string // job status
	LogLines []string
}

func (result ExecResult) LogString() string {
	return strings.Join(result.LogLines, "\n")
}

type ExecError struct {
	Options ExecOptions
	Result  ExecResult
}

func (err ExecError) Error() string {
	return fmt.Sprintf("systemd/exec %v %v: %v\n%v", err.Options.Unit, err.Result.Status, err.Options.Cmd, err.Result.LogString())
}

func propEnvironment(envs []string) dbus.Property {
	return dbus.Property{
		Name:  "Environment",
		Value: godbus.MakeVariant(envs),
	}
}

type systemdExec struct {
	unit    string
	options ExecOptions
	conn    *dbus.Conn
	journal *sdjournal.Journal
	ch      chan string
}

func (se *systemdExec) connect() error {
	if conn, err := dbus.NewSystemConnection(); err != nil {
		return fmt.Errorf("dbus.NewSystemConnection: %v", err)
	} else {
		se.conn = conn
	}

	return nil
}

// clear any existing failed unit, no-op if not loaded
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
		dbus.PropType("oneshot"),
		propEnvironment(se.options.Env),
		dbus.PropExecStart(se.options.Cmd, false), // XXX: bool flag meaning is inverted?
	}

	log.Printf("systemd/exec %v: start %#v", se.unit, properties)

	// fail => error if unit already exists because it's still running
	if _, err := se.conn.StartTransientUnit(se.unit, "fail", properties, se.ch); err != nil {
		return fmt.Errorf("dbus.StartTransientUnit %v: %v", se.unit, err)
	}

	return nil
}

func (se *systemdExec) getServiceTimestamp(propertyName string) (uint64, error) {
	if property, err := se.conn.GetUnitTypeProperty(se.unit, "Service", propertyName); err != nil {
		return 0, fmt.Errorf("dbus.GetUnitTypeProperty %v: %v", propertyName, err)
	} else if value, ok := property.Value.Value().(uint64); !ok {
		return 0, fmt.Errorf("Invalid property value: %#v", property.Value)
	} else {
		return value, nil
	}
}

// open journal for reading, if available
func (se *systemdExec) openJournal() error {
	// this silently succeeds if no journal files are available
	if journal, err := sdjournal.NewJournal(); err != nil {
		return fmt.Errorf("sdjournal.NewJournal: %v", err)
	} else {
		se.journal = journal
	}

	if err := se.journal.AddMatch("_SYSTEMD_UNIT=" + se.unit); err != nil {
		return fmt.Errorf("sdjournal.AddMatch: %v", err)
	}

	// XXX: returns 0 without error if unit does not exist?
	if startTimestamp, err := se.getServiceTimestamp("ExecMainStartTimestamp"); err != nil {
		return err
	} else if err := se.journal.SeekRealtimeUsec(startTimestamp); err != nil {
		return fmt.Errorf("sdjournal.SeekRealtimeUsec: %v", err)
	}

	return nil
}

func (se *systemdExec) readJournal() ([]string, error) {
	var lines []string

	log.Printf("systemd/exec %v: read journal...", se.unit)

	for {
		if n, err := se.journal.Next(); err != nil {
			return lines, fmt.Errorf("sdjournal.Next: %v", err)
		} else if n == 0 {
			break
		} else if entry, err := se.journal.GetEntry(); err != nil {
			return lines, fmt.Errorf("sdjournal.GetEntry: %v", err)
		} else {
			var t = time.Unix(int64(entry.RealtimeTimestamp/1e6), int64(entry.RealtimeTimestamp%1e6*1e3))
			var message = entry.Fields["MESSAGE"]

			log.Printf("systemd/exec %v: journal %v: %v", se.unit, t, message)

			lines = append(lines, message)
		}
	}

	return lines, nil
}

func (se *systemdExec) wait() (string, error) {
	log.Printf("systemd/exec %v: wait...", se.unit)

	status := <-se.ch

	log.Printf("systemd/exec %v: done, status=%v", se.unit, status)

	return status, nil
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
	if se.conn != nil {
		se.conn.Close()
	}

	if se.journal != nil {
		se.journal.Close()
	}
}

func Exec(name string, options ExecOptions) (ExecResult, error) {
	options.Unit = fmt.Sprintf("%v.service", name)

	var se = systemdExec{
		unit:    options.Unit,
		options: options,
		ch:      make(chan string),
	}
	var result ExecResult

	defer se.close()

	log.Printf("systemd/exec %v: %#v", se.unit, options)

	if err := se.connect(); err != nil {
		return result, err
	}

	if err := se.reset(); err != nil {
		return result, err
	}

	if err := se.start(); err != nil {
		return result, err
	}

	// XXX: race if unit exit with success?
	if err := se.openJournal(); err != nil {
		return result, err
	}

	if status, err := se.wait(); err != nil {
		return result, err
	} else {
		result.Status = status
	}

	if lines, err := se.readJournal(); err != nil {
		return result, err
	} else {
		result.LogLines = lines
	}

	if result.Status == "done" {
		log.Printf("systemd/exec %v: done", se.unit)
	} else {
		return result, ExecError{
			Options: options,
			Result:  result,
		}
	}

	return result, nil
}
