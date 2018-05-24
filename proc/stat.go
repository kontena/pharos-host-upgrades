package proc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Stat struct {
	BootTime time.Time
}

func (stat *Stat) set(field string, args []string) error {
	switch field {
	case "btime":
		var sec int64

		if len(args) < 1 {
			return fmt.Errorf("Missing arg")
		} else if _, err := fmt.Sscanf(args[0], "%d", &sec); err != nil {
			return err
		} else {
			stat.BootTime = time.Unix(sec, 0)
		}
	}

	return nil
}

func ReadStat() (Stat, error) {
	var stat Stat

	if file, err := os.Open("/proc/stat"); err != nil {
		return stat, err
	} else {
		for scanner := bufio.NewScanner(file); scanner.Scan(); {
			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) < 1 {
				continue
			} else if err := stat.set(fields[0], fields[1:]); err != nil {
				return stat, fmt.Errorf("Invalid /proc/stat field %v: %v", fields[0], err)
			}
		}
	}

	return stat, nil
}
