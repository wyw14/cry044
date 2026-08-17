package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
)

func (p *PGLedger) CreateBatch(ctx context.Context, batch domain.Batch, requestID string) (domain.Batch, error) {
	raw, err := encode(batch)
	if err != nil {
		return batch, err
	}
	_, err = p.pool.Exec(ctx, `insert into batches(id,status,revision,idempotency_key,payload)
values($1,$2,$3,$4,$5) on conflict(idempotency_key) do nothing`, batch.ID, batch.Status, batch.Revision, requestID, raw)
	if err != nil {
		return batch, err
	}
	return p.Batch(ctx, batch.ID)
}

func (p *PGLedger) Batch(ctx context.Context, id string) (domain.Batch, error) {
	raw, err := p.payload(ctx, `select payload from batches where id=$1`, id)
	if err != nil {
		return domain.Batch{}, err
	}
	var result domain.Batch
	return result, decode(raw, &result)
}

func (p *PGLedger) SaveBatch(ctx context.Context, batch domain.Batch, expected int64) error {
	raw, err := encode(batch)
	if err != nil {
		return err
	}
	templateID, version := "", 0
	if batch.Snapshot != nil {
		templateID, version = batch.Snapshot.TemplateID, batch.Snapshot.Version
	}
	_, execErr := p.pool.Exec(ctx, `update batches set status=$2, revision=$3, template_id=$4,
template_version=$5, payload=$6 where id=$1`, batch.ID, batch.Status, batch.Revision, templateID, version, raw)
	return execErr
}
