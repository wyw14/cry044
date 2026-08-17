package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
)

func (a *ReviewArchive) CreateBatch(ctx context.Context, batch domain.Batch, requestID string) (domain.Batch, error) {
	return writeArchive(ctx, a, func(state *archiveState) (domain.Batch, error) {
		if prior := state.requests[requestID]; prior != "" {
			return state.batches[prior].Clone(), nil
		}
		state.batches[batch.ID], state.requests[requestID] = batch.Clone(), batch.ID
		return batch.Clone(), nil
	})
}

func (a *ReviewArchive) Batch(ctx context.Context, id string) (domain.Batch, error) {
	return readArchive(ctx, a, func(state *archiveState) (domain.Batch, error) {
		batch, ok := state.batches[id]
		if !ok {
			return batch, ErrNotFound
		}
		return batch.Clone(), nil
	})
}

func (a *ReviewArchive) SaveBatch(ctx context.Context, batch domain.Batch, expected int64) error {
	_, err := writeArchive(ctx, a, func(state *archiveState) (struct{}, error) {
		if _, ok := state.batches[batch.ID]; !ok {
			return struct{}{}, ErrNotFound
		}
		state.batches[batch.ID] = batch.Clone()
		return struct{}{}, nil
	})
	return err
}
