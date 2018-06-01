// +build cgo

package systemd

import (
	"fmt"
	"log"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

type journalReader struct {
	journal *sdjournal.Journal
}

// open journal for reading, if available
func OpenJournal(options JournalOptions) (JournalReader, error) {
	var jr journalReader

	// this silently succeeds if no journal files are available
	if journal, err := sdjournal.NewJournal(); err != nil {
		return nil, fmt.Errorf("sdjournal.NewJournal: %v", err)
	} else {
		jr.journal = journal
	}

	if options.Unit == "" {

	} else if err := jr.journal.AddMatch("_SYSTEMD_UNIT=" + options.Unit); err != nil {
		return nil, fmt.Errorf("sdjournal.AddMatch: %v", err)
	}

	if options.StartTimestamp == 0 {

	} else if err := jr.journal.SeekRealtimeUsec(options.StartTimestamp); err != nil {
		return nil, fmt.Errorf("sdjournal.SeekRealtimeUsec: %v", err)
	}

	return &jr, nil
}

func (jr *journalReader) Read() ([]string, error) {
	var lines []string

	log.Printf("systemd/journal: read journal...")

	for {
		if n, err := jr.journal.Next(); err != nil {
			return lines, fmt.Errorf("sdjournal.Next: %v", err)
		} else if n == 0 {
			break
		} else if entry, err := jr.journal.GetEntry(); err != nil {
			return lines, fmt.Errorf("sdjournal.GetEntry: %v", err)
		} else {
			var t = time.Unix(int64(entry.RealtimeTimestamp/1e6), int64(entry.RealtimeTimestamp%1e6*1e3))
			var message = entry.Fields["MESSAGE"]

			log.Printf("systemd/journal: %v %v", t, message)

			lines = append(lines, message)
		}
	}

	return lines, nil
}

func (jr *journalReader) Close() error {
	return jr.journal.Close()
}
