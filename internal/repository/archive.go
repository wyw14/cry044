package repository

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/wyw14/cry044/internal/domain"
)

// ErrNotFound and ErrRevision are intentionally small domain errors.  The
// archive uses them for both the in-memory test ledger and the SQL adapter so
// application services do not need to know which persistence mode is active.
var (
	ErrNotFound = errors.New("review record not found")
	ErrRevision = errors.New("review record revision conflict")
)

// ReviewArchive is a concurrency-safe projection of the review room.  The
// maps are grouped by business aggregate instead of one generic CRUD table;
// that makes snapshot, panel, and audit lifecycles explicit in tests.
type ReviewArchive struct {
	mu    sync.RWMutex
	state archiveState
}

type archiveState struct {
	templates map[string]domain.StandardTemplate
	batches   map[string]domain.Batch
	requests  map[string]string
	panels    map[string][]domain.Review
	trail     map[string][]domain.AuditEvent
}

func NewReviewArchive() *ReviewArchive {
	return &ReviewArchive{
		state: archiveState{
			templates: make(map[string]domain.StandardTemplate),
			batches:   make(map[string]domain.Batch),
			requests:  make(map[string]string),
			panels:    make(map[string][]domain.Review),
			trail:     make(map[string][]domain.AuditEvent),
		},
	}
}

func archiveKey(id string, version int) string {
	return id + "#" + strconv.Itoa(version)
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func readArchive[T any](ctx context.Context, archive *ReviewArchive, inspect func(*archiveState) (T, error)) (T, error) {
	var zero T
	if err := contextErr(ctx); err != nil {
		return zero, err
	}
	archive.mu.RLock()
	defer archive.mu.RUnlock()
	return inspect(&archive.state)
}

func writeArchive[T any](ctx context.Context, archive *ReviewArchive, change func(*archiveState) (T, error)) (T, error) {
	var zero T
	if err := contextErr(ctx); err != nil {
		return zero, err
	}
	archive.mu.Lock()
	defer archive.mu.Unlock()
	return change(&archive.state)
}
