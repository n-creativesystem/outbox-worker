package store

import (
	"context"
	"sync"

	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/service/errors"
	"go.opentelemetry.io/otel/attribute"
)

type Store[T any] struct {
	mp sync.Map
}

func (s *Store[T]) Load(ctx context.Context, key string) (_ T, rErr error) {
	ctx = trace.StartSpan(ctx, "Store/Load", attribute.String("key", key))
	defer func() { trace.EndSpan(ctx, rErr) }()
	v, ok := s.mp.Load(key)
	if !ok {
		var v T
		return v, errors.NewNotFoundKeyError(key)
	}
	val, ok := v.(T)
	if !ok {
		var v T
		return v, errors.NewNotFoundKeyError(key)
	}
	return val, nil
}

func (s *Store[T]) Add(ctx context.Context, key string, value T) {
	ctx = trace.StartSpan(ctx, "Store/Load", attribute.String("key", key))
	defer func() { trace.EndSpan(ctx, nil) }()
	s.mp.Store(key, value)
}
