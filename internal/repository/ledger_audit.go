package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry044/internal/domain"
)

func (p *PGLedger) AppendAudit(ctx context.Context, event domain.AuditEvent) error {
	return pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		var current string
		err := tx.QueryRow(ctx, `select hash from audit_events where aggregate_id=$1 order by created_at desc,id desc limit 1 for update`, event.AggregateID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			current, err = "", nil
		}
		if err != nil {
			return err
		}
		if current != event.PreviousHash {
			return errors.New("audit chain conflict")
		}
		document, err := encode(event)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `insert into audit_events(created_at,payload,hash,previous_hash,aggregate_id,id) values($1,$2,$3,$4,$5,$6)`, event.At, document, event.Hash, event.PreviousHash, event.AggregateID, event.ID)
		return err
	})
}

func (p *PGLedger) AuditHead(ctx context.Context, aggregateID string) (string, error) {
	var current string
	err := p.pool.QueryRow(ctx, `select hash from audit_events where aggregate_id=$1 order by created_at desc,id desc limit 1`, aggregateID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return current, err
}
