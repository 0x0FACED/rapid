package controller

import (
	"sort"
	"strings"
	"sync"

	"github.com/0x0FACED/rapid/internal/model"
)

type FileState struct {
	Files         map[string]model.File
	FilteredFiles []model.File
	SearchQuery   string
	mu            sync.RWMutex
	sortedIDs     []string
}

func NewFileState() *FileState {
	return &FileState{
		Files:         make(map[string]model.File),
		FilteredFiles: make([]model.File, 0),
	}
}

func (f *FileState) Add(id string, file model.File) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.Files[id]; !exists {
		f.sortedIDs = append(f.sortedIDs, id)
		sort.Slice(f.sortedIDs, func(i, j int) bool {
			return f.sortedIDs[i] < f.sortedIDs[j]
		})
	}
	f.Files[id] = file

	if f.SearchQuery != "" {
		f.Filter(f.SearchQuery)
	}
}

func (f *FileState) GetAll() []model.File {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.SearchQuery == "" {
		result := make([]model.File, 0, len(f.sortedIDs))
		for _, id := range f.sortedIDs {
			if file, exists := f.Files[id]; exists {
				result = append(result, file)
			}
		}
		return result
	}

	return f.FilteredFiles
}

func (f *FileState) Filter(query string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.SearchQuery = strings.ToLower(query)
	f.FilteredFiles = make([]model.File, 0)

	for _, id := range f.sortedIDs {
		file := f.Files[id]
		if strings.Contains(strings.ToLower(file.Name), f.SearchQuery) {
			f.FilteredFiles = append(f.FilteredFiles, file)
		}
	}
}
