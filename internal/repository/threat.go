package repository

import (
	"errors"
	"sync"

	"threat-service/api"
)

var ErrNotFound = errors.New("threat not found")

type InMemoryThreatRepository struct {
	mu      sync.RWMutex
	threats map[string]api.Threat
}

var _ ThreatRepository = (*InMemoryThreatRepository)(nil)

func NewInMemoryThreatRepository() *InMemoryThreatRepository {
	return &InMemoryThreatRepository{
		threats: make(map[string]api.Threat),
	}
}

func (r *InMemoryThreatRepository) Update(md5 string, threat api.Threat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threats[md5]; !exists {
		return ErrNotFound
	}

	threat.Md5 = md5
	r.threats[md5] = threat

	return nil
}

var ErrAlreadyExists = errors.New("threat already exists")

func (r *InMemoryThreatRepository) Create(threat api.Threat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.threats[threat.Md5]; exists {
		return ErrAlreadyExists
	}

	r.threats[threat.Md5] = threat
	return nil
}

func (r *InMemoryThreatRepository) Get(md5 string) (api.Threat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	threat, exists := r.threats[md5]
	if !exists {
		return api.Threat{}, ErrNotFound
	}

	return threat, nil
}
