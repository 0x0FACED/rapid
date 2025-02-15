package storage

import (
	"sync"

	"github.com/0x0FACED/rapid/internal/model"
	"github.com/google/uuid"
)

type FileStorage struct {
	// Key - id
	// Value - RapidFile
	files map[string]model.RapidFile

	// Lock() for write
	//
	// RLock() for read
	rwmu sync.RWMutex
}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) Add(path string) error {
	_uuid := uuid.NewString()
	s, err := size(path)
	if err != nil {
		return err
	}

	file := model.RapidFile{
		ID:   _uuid,
		Name: extractName(path),
		Ext:  extractExt(path),
		Path: path,
		Size: s,
	}
	fs.rwmu.Lock()
	fs.files[_uuid] = file
	fs.rwmu.Unlock()
	return nil
}

func (fs *FileStorage) ListFiles() map[string]model.RapidFile {
	return fs.files
}
