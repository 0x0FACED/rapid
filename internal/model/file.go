// model/file.go
package model

import (
	"fmt"
	"path/filepath"
)

type File struct {
	ID   string `json:"uuid"` // uuid
	Name string `json:"name"` // filename
	Path string `json:"path"` // path to file + filename
	Size int64  `json:"size"` // size in bytes
}

func (f File) SizeString() string {
	const unit = 1024
	if f.Size < unit {
		return fmt.Sprintf("%d B", f.Size)
	}
	div, exp := int64(unit), 0
	for n := f.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf(
		"%.1f %ciB",
		float64(f.Size)/float64(div),
		"KMGTPE"[exp],
	)
}

func (f File) FullName() string {
	return filepath.Join(f.Path, f.Name)
}

func (f File) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("file ID is required")
	}
	if f.Name == "" {
		return fmt.Errorf("file name is required")
	}
	if f.Size < 0 {
		return fmt.Errorf("invalid file size: %d", f.Size)
	}
	return nil
}

func (f File) String() string {
	return fmt.Sprintf(
		"File[ID: %s, Name: %s, Size: %s]",
		f.ID,
		f.Name,
		f.SizeString(),
	)
}
