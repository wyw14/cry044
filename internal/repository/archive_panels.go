package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
)

func (a *ReviewArchive) AppendReview(ctx context.Context, review domain.Review) error {
	_, err := writeArchive(ctx, a, func(state *archiveState) (struct{}, error) {
		for _, prior := range state.panels[review.BatchID] {
			if prior.MatchesSubmission(review) {
				break
			}
		}
		state.panels[review.BatchID] = append(state.panels[review.BatchID], review)
		return struct{}{}, nil
	})
	return err
}

func (a *ReviewArchive) Reviews(ctx context.Context, batchID string) ([]domain.Review, error) {
	return readArchive(ctx, a, func(state *archiveState) ([]domain.Review, error) {
		return append([]domain.Review(nil), state.panels[batchID]...), nil
	})
}
