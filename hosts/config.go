package hosts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

func testDir(path string) (bool, error) {
	if stat, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else if !stat.IsDir() {
		return true, fmt.Errorf("not a directory")
	} else {
		return true, nil
	}
}

type Config struct {
	path  string
	mount string
}

// Set path to config files
func (config *Config) UsePath(path string) (bool, error) {
	if exists, err := testDir(path); err != nil {
		return exists, err
	} else if exists {
		config.path = path

		return true, nil
	} else {
		return false, nil
	}
}

// set path to host mount
func (config *Config) UseMount(path string) (bool, error) {
	if exists, err := testDir(path); err != nil {
		return exists, err
	} else if exists {
		config.mount = path

		return true, nil
	} else {
		return false, nil
	}
}

func (config *Config) Path(name string) string {
	return filepath.Join(config.path, name)
}

func (config *Config) MountPath(name string) string {
	return filepath.Join(config.mount, name)
}

func (config *Config) Stat(name string) (os.FileInfo, error) {
	return os.Stat(config.Path(name))
}

func (config *Config) FileExists(name string) (bool, error) {
	if stat, err := config.Stat(name); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else if !stat.Mode().IsRegular() {
		return false, fmt.Errorf("Not a file: %v/%v", config.path, name)
	} else {
		return true, nil
	}
}

func (config *Config) Open(name string) (*os.File, error) {
	return os.Open(config.Path(name))
}

func (config *Config) WriteHostFile(name string, src io.Reader) (string, error) {
	if config.mount == "" {
		return "", fmt.Errorf("No host mount given")
	}

	var mountPath = config.MountPath(name)
	var tempPath = mountPath + ".tmp"
	var hostPath = mountPath // assume same for now

	if tempFile, err := os.Create(tempPath); err != nil {
		return "", fmt.Errorf("Create host %v file: %v", name, err)
	} else if _, err := io.Copy(tempFile, src); err != nil {
		return "", fmt.Errorf("Copy to host %v file: %v", name, err)
	} else if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("Close host %v file: %v", name, err)
	} else if err := os.Rename(tempPath, mountPath); err != nil {
		return "", fmt.Errorf("Rename host %v file: %v", name, err)
	}

	return hostPath, nil
}

// Copy config file to host mount, returning host path to copied file
func (config *Config) CopyHostFile(name string) (string, error) {
	if configFile, err := config.Open(name); err != nil {
		return "", fmt.Errorf("Open config %v file: %v", name, err)
	} else {
		defer configFile.Close()

		return config.WriteHostFile(name, configFile)
	}
}

func (config *Config) GenerateFile(name string, text string, data interface{}) (string, error) {
	var t = template.New(name)
	var buf bytes.Buffer

	if _, err := t.Parse(text); err != nil {
		return "", fmt.Errorf("Invalid template for %v: %v", name, err)
	}

	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("Failed template for %v: %v", name, err)
	}

	return config.WriteHostFile(name, &buf)
}
