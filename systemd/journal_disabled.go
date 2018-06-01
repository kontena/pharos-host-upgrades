// +build !cgo

package systemd

func OpenJournal(options JournalOptions) (JournalReader, error) {
	return nil, nil
}
