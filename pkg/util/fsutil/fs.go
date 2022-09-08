package fsutil

import (
	"os"
	"path/filepath"
)

func CreateOrOpen(filename string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		if os.IsExist(err) {
			return os.Open(filename)
		}
		return nil, err
	}
	return f, nil
}

func WriteFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return os.WriteFile(filename, data, os.ModePerm)
}
