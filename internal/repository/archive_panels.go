package repository

import (
	"context"
	"errors"
	"github.com/wyw14/cry044/internal/domain"
)

var ErrDuplicateReview = errors.New("review already submitted for material")

func (a *ReviewArchive) AppendReview(ctx context.Context, review domain.Review) error {
	_, err := writeArchive(ctx, a, func(state *archiveState) (struct{}, error) {
		current := state.panels[review.BatchID]
		for _, prior := range current {
			if !prior.MatchesSubmission(review) {
				continue
			}
			return struct{}{}, ErrDuplicateReview
		}
		state.panels[review.BatchID] = append(current, review)
		return struct{}{}, nil
	})
	return err
}

func (a *ReviewArchive) Reviews(ctx context.Context, batchID string) ([]domain.Review, error) {
	return readArchive(ctx, a, func(state *archiveState) ([]domain.Review, error) {
		return append([]domain.Review(nil), state.panels[batchID]...), nil
	})
}
