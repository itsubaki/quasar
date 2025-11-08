package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
)

var (
	_ Store = (*MemoryStore)(nil)
	_ Store = (*FireStore)(nil)
)

type Snippet struct {
	Code      string
	CreatedAt time.Time
}

type Store interface {
	Put(ctx context.Context, id string, snippet *Snippet) error
	Get(ctx context.Context, id string) (*Snippet, error)
}

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

type FireStore struct {
	Collection string
	Client     *firestore.Client
}

func (s *FireStore) Put(ctx context.Context, id string, snippet *Snippet) error {
	if _, err := s.Client.Collection(s.Collection).Doc(id).Set(ctx, map[string]any{
		"code":       snippet.Code,
		"created_at": snippet.CreatedAt,
	}); err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return nil
}

func (s *FireStore) Get(ctx context.Context, id string) (*Snippet, error) {
	doc, err := s.Client.Collection(s.Collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	code, err := Get[string](doc.Data(), "code")
	if err != nil {
		return nil, err
	}

	createdAt, err := Get[time.Time](doc.Data(), "created_at")
	if err != nil {
		return nil, err
	}

	return &Snippet{
		Code:      code,
		CreatedAt: createdAt,
	}, nil
}

func Get[T any](data map[string]any, key string) (T, error) {
	v, ok := data[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("%v not found", key)
	}

	typed, ok := v.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("invalid type(%T)", v)
	}

	return typed, nil
}
