package repository

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGLedger is the durable review-room adapter.  Payloads remain JSON so the
// domain can evolve independently from the audit and reporting projections.
type PGLedger struct{ pool *pgxpool.Pool }

func NewPGLedger(pool *pgxpool.Pool) *PGLedger { return &PGLedger{pool: pool} }

func encode(value any) ([]byte, error) { return json.Marshal(value) }

func (p *PGLedger) payload(ctx context.Context, query string, args ...any) ([]byte, error) {
	var raw []byte
	if err := p.pool.QueryRow(ctx, query, args...).Scan(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func decode(raw []byte, target any) error { return json.Unmarshal(raw, target) }
