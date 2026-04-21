package monitor

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("monitor not found")

type Repository interface {
	Create(ctx context.Context, m *Monitor) error
	GetByID(ctx context.Context, id string) (*Monitor, error)
	List(ctx context.Context) ([]*Monitor, error)
	Update(ctx context.Context, m *Monitor) error
	Delete(ctx context.Context, id string) error
	SaveCheck(ctx context.Context, check *Check) error
	ListChecks(ctx context.Context, monitorID string) ([]*Check, error)
	ListDueMonitors(ctx context.Context) ([]*Monitor, error)
}