package storage

import (
	"os"
	"path/filepath"
	"strings"
)

func extractName(path string) string {
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func extractExt(path string) string {
	return filepath.Ext(path)
}

func size(path string) (int64, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return -1, err // todo: wrap
	}

	return finfo.Size(), nil
}
