package catalog

import (
	"errors"
	"sync"
)

// ErrNotFound is returned when a service isn't in the catalog.
var ErrNotFound = errors.New("service not found")
var ErrExists = errors.New("service already exists")

// Service is the domain model (separate from the protobuf type on purpose).
type Service struct {
	Name     string
	Owner    string
	Language string
	Replicas int32
}

// Store is an in-memory, concurrency-safe catalog.
// (Week 2+ you could swap this for a real DB behind the same interface.)
type Store struct {
	mu       sync.RWMutex
	services map[string]Service
}

func NewStore() *Store {
	return &Store{services: make(map[string]Service)}
}

func (s *Store) Register(svc Service) (Service, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[svc.Name]; ok {
		return Service{}, ErrExists
	}
	s.services[svc.Name] = svc
	return svc, nil
}

func (s *Store) Get(name string) (Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[name]
	if !ok {
		return Service{}, ErrNotFound
	}
	return svc, nil
}

func (s *Store) List() []Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Service, 0, len(s.services))
	for _, svc := range s.services {
		out = append(out, svc)
	}
	return out
}
