package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
)

func (a *ReviewArchive) SaveTemplate(ctx context.Context, value domain.StandardTemplate, expected int64, requestID string) (domain.StandardTemplate, error) {
	return writeArchive(ctx, a, func(state *archiveState) (domain.StandardTemplate, error) {
		if requestID != "" {
			if prior := state.requests[requestID]; prior != "" {
				return state.templates[prior].Clone(), nil
			}
		}
		key := archiveKey(value.ID, value.Version)
		current, exists := state.templates[key]
		if (exists && current.Revision != expected) || (!exists && expected != 0) {
			return value, ErrRevision
		}
		state.templates[key] = value.Clone()
		if requestID != "" {
			state.requests[requestID] = key
		}
		return value.Clone(), nil
	})
}

func (a *ReviewArchive) Template(ctx context.Context, id string, version int) (domain.StandardTemplate, error) {
	return readArchive(ctx, a, func(state *archiveState) (domain.StandardTemplate, error) {
		value, ok := state.templates[archiveKey(id, version)]
		if !ok {
			return value, ErrNotFound
		}
		return value.Clone(), nil
	})
}

func (a *ReviewArchive) TemplateReferenced(ctx context.Context, id string, version int) (bool, error) {
	return readArchive(ctx, a, func(state *archiveState) (bool, error) {
		for _, batch := range state.batches {
			if snapshot := batch.Snapshot; snapshot != nil && snapshot.TemplateID == id && snapshot.Version == version {
				return true, nil
			}
		}
		return false, nil
	})
}
