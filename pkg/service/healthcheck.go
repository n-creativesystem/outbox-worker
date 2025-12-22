package service

import "context"

type HealthCheck interface {
	Start(ctx context.Context) error
}
