package store

import (
	"context"
	"errors"
	"sync"
)

var ErrNoSuchEntity = errors.New("no such entity")

type MemoryStore struct {
	m map[string]*Snippet
	sync.RWMutex
}

func (s *MemoryStore) Put(_ context.Context, id string, snippet *Snippet) error {
	s.Lock()
	defer s.Unlock()

	if s.m == nil {
		s.m = make(map[string]*Snippet)
	}

	s.m[id] = snippet
	return nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (*Snippet, error) {
	s.RLock()
	defer s.RUnlock()

	snippet, ok := s.m[id]
	if !ok {
		return nil, ErrNoSuchEntity
	}

	return snippet, nil
}
