package repository

import (
	"context"
	"errors"
	"github.com/wyw14/cry044/internal/domain"
)

func (a *ReviewArchive) AppendAudit(ctx context.Context, event domain.AuditEvent) error {
	_, err := writeArchive(ctx, a, func(state *archiveState) (struct{}, error) {
		chain := state.trail[event.AggregateID]
		if len(chain) != 0 && chain[len(chain)-1].Hash != event.PreviousHash {
			return struct{}{}, errors.New("audit chain conflict")
		}
		state.trail[event.AggregateID] = append(chain, event)
		return struct{}{}, nil
	})
	return err
}

func (a *ReviewArchive) AuditHead(ctx context.Context, aggregateID string) (string, error) {
	return readArchive(ctx, a, func(state *archiveState) (string, error) {
		chain := state.trail[aggregateID]
		if len(chain) == 0 {
			return "", nil
		}
		return chain[len(chain)-1].Hash, nil
	})
}
