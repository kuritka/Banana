package utils

import (
	"github.com/ahmetb/go-linq"
	"os"
	"path/filepath"
)

func RemoveAllExcept(pattern string) error {

	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	allFiles, err := filepath.Glob(filepath.Join(filepath.Dir(pattern), "*.*"))
	if err != nil {
		return err
	}
	var remove []string
	linq.From(allFiles).Except(linq.From(files)).ToSlice(&remove)

	for _, f := range remove {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
