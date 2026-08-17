package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry044/internal/domain"
)

func (p *PGLedger) AppendReview(ctx context.Context, review domain.Review) error {
	return pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		var status domain.BatchStatus
		if err := tx.QueryRow(ctx, `select status from batches where id=$1 for update`, review.BatchID).Scan(&status); err != nil {
			return err
		}
		if status != domain.BatchSubmitted && status != domain.BatchReReview {
			return domain.ErrInvalidBatchFlow
		}
		document, err := encode(review)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `insert into reviews(payload,reviewer_id,material_id,batch_id,id) values($1,$2,$3,$4,$5)`, document, review.ReviewerID, review.MaterialID, review.BatchID, review.ID)
		return err
	})
}

func (p *PGLedger) Reviews(ctx context.Context, batchID string) ([]domain.Review, error) {
	rows, err := p.pool.Query(ctx, `select payload from reviews where batch_id=$1 order by id`, batchID)
	if err != nil {
		return nil, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Review, error) {
		var document []byte
		if err := row.Scan(&document); err != nil {
			return domain.Review{}, err
		}
		var review domain.Review
		return review, decode(document, &review)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return []domain.Review{}, nil
	}
	return items, err
}
