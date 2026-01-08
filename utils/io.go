package utils

import (
	"os"
	"path/filepath"
)

func WriteFileAtomic(path string, contents []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	tmp, err := os.CreateTemp(dir, "."+base+".tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	cleanup := func(e error) error {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return e
	}

	if err := tmp.Chmod(perm); err != nil {
		return cleanup(err)
	}
	if _, err := tmp.Write(contents); err != nil {
		return cleanup(err)
	}
	if err := tmp.Sync(); err != nil {
		return cleanup(err)
	}
	if err := tmp.Close(); err != nil {
		return cleanup(err)
	}

	return os.Rename(tmpName, path)
}
