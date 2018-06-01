package systemd

type JournalOptions struct {
	Unit           string
	StartTimestamp uint64
}

type JournalReader interface {
	Read() ([]string, error)
	Close() error
}
