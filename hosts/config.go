package hosts

import (
	"fmt"
	"os"
)

type Config struct {
	path string
}

func (config *Config) Load(path string) error {
	if stat, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("not a directory")
	} else {
		config.path = path
	}

	return nil
}
