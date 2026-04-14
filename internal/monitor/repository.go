package repository

import(
	"context"
)

type Repository interface {
	SaveCheck(ctx context.Context, check *Check) error //it handles peristant of a *Check(configuration and current state of a monitor)
	ListDueMonitors(ctx context.Context) ([]*Monitor, error)
	
}