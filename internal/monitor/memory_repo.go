package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryRepo struct {
	mu       sync.RWMutex
	monitors map[string]*Monitor
	checks   map[string][]*Check
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		monitors: make(map[string]*Monitor),
		checks:   make(map[string][]*Check),
	}
}

func (r *MemoryRepo) Create(ctx context.Context, m *Monitor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m.ID        = uuid.New().String()
	m.Status    = StatusActive
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	r.monitors[m.ID] = m
	return nil
}

func (r *MemoryRepo) GetByID(ctx context.Context, id string) (*Monitor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.monitors[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

func (r *MemoryRepo) List(ctx context.Context) ([]*Monitor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Monitor, 0, len(r.monitors))
	for _, m := range r.monitors {
		result = append(result, m)
	}
	return result, nil
}

func (r *MemoryRepo) Update(ctx context.Context, m *Monitor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.monitors[m.ID]; !ok {
		return ErrNotFound
	}
	m.UpdatedAt = time.Now()
	r.monitors[m.ID] = m
	return nil
}

func (r *MemoryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.monitors[id]; !ok {
		return ErrNotFound
	}
	delete(r.monitors, id)
	return nil
}

func (r *MemoryRepo) SaveCheck(ctx context.Context, check *Check) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	check.ID        = uuid.New().String()
	check.CheckedAt = time.Now()
	r.checks[check.MonitorID] = append(r.checks[check.MonitorID], check)
	return nil
}

func (r *MemoryRepo) ListChecks(ctx context.Context, monitorID string) ([]*Check, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.checks[monitorID], nil
}

func (r *MemoryRepo) ListDueMonitors(ctx context.Context) ([]*Monitor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var due []*Monitor
	now := time.Now()
	for _, m := range r.monitors {
		if m.Status != StatusActive {
			continue
		}
		if m.LastCheckedAt == nil {
			due = append(due, m)
			continue
		}
		next := m.LastCheckedAt.Add(time.Duration(m.IntervalS) * time.Second)
		if now.After(next) {
			due = append(due, m)
		}
	}
	return due, nil
}