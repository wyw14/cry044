package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
)

func (p *PGLedger) SaveTemplate(ctx context.Context, value domain.StandardTemplate, expected int64, requestID string) (domain.StandardTemplate, error) {
	raw, err := encode(value)
	if err != nil {
		return value, err
	}
	if expected == 0 {
		_, err = p.pool.Exec(ctx, `insert into templates(id,version,status,revision,idempotency_key,payload)
values($1,$2,$3,$4,$5,$6) on conflict(idempotency_key) do nothing`, value.ID, value.Version, value.Status, value.Revision, requestID, raw)
	} else {
		changed, execErr := p.pool.Exec(ctx, `update templates set status=$3, revision=$4, payload=$5
where id=$1 and version=$2 and revision=$6`, value.ID, value.Version, value.Status, value.Revision, raw, expected)
		err = execErr
		if err == nil && changed.RowsAffected() != 1 {
			return value, ErrRevision
		}
	}
	if err != nil {
		return value, err
	}
	return p.Template(ctx, value.ID, value.Version)
}

func (p *PGLedger) Template(ctx context.Context, id string, version int) (domain.StandardTemplate, error) {
	raw, err := p.payload(ctx, `select payload from templates where id=$1 and version=$2`, id, version)
	if err != nil {
		return domain.StandardTemplate{}, err
	}
	var result domain.StandardTemplate
	return result, decode(raw, &result)
}

func (p *PGLedger) TemplateReferenced(ctx context.Context, id string, version int) (bool, error) {
	var found bool
	err := p.pool.QueryRow(ctx, `select exists(select 1 from batches where template_id=$1 and template_version=$2)`, id, version).Scan(&found)
	return found, err
}
