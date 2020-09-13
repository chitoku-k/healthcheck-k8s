package service

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type HealthCheck interface {
	Do(ctx context.Context, nodeName string) (bool, error)
}
