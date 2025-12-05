package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

type Firestore struct {
	Collection string
	Client     *firestore.Client
}

func (s *Firestore) Put(ctx context.Context, id string, snippet *Snippet) error {
	if _, err := s.Client.Collection(s.Collection).Doc(id).Set(ctx, map[string]any{
		"id":         id,
		"code":       snippet.Code,
		"created_at": snippet.CreatedAt,
	}); err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return nil
}

func (s *Firestore) Get(ctx context.Context, id string) (*Snippet, error) {
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
