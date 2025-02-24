package controller

import (
	"sort"
	"sync"

	"github.com/0x0FACED/rapid/internal/model"
)

type ServerState struct {
	Instances  map[string]model.ServiceInstance
	mu         sync.RWMutex
	sortedKeys []string
}

func NewServerState() *ServerState {
	return &ServerState{
		Instances: make(map[string]model.ServiceInstance),
	}
}

func (s *ServerState) AddOrUpdate(instance model.ServiceInstance) bool {
	key := instance.Key()
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Instances[key]; !exists {
		s.sortedKeys = append(s.sortedKeys, key)
	}
	_, ok := s.Instances[key]
	if !ok {
		s.Instances[key] = instance
		return true
	}

	sort.Slice(s.sortedKeys, func(i, j int) bool {
		return s.sortedKeys[i] < s.sortedKeys[j]
	})
	return false
}

func (s *ServerState) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Instances, key)
}

func (s *ServerState) GetAll() []model.ServiceInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.ServiceInstance, 0, len(s.sortedKeys))
	for _, key := range s.sortedKeys {
		if instance, exists := s.Instances[key]; exists {
			result = append(result, instance)
		}
	}
	return result
}
