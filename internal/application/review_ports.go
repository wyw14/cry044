package application

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
	"time"
)

// ReviewRepository is deliberately expressed as business capabilities.  A
// service can ask for a standard snapshot, a panel transcript, or an audit
// head without depending on the storage technology behind it.
type ReviewRepository interface {
	SaveTemplate(ctx context.Context, template domain.StandardTemplate, expectedRevision int64, requestKey string) (stored domain.StandardTemplate, err error)
	Template(ctx context.Context, templateID string, version int) (stored domain.StandardTemplate, err error)
	TemplateReferenced(ctx context.Context, templateID string, version int) (referenced bool, err error)
	CreateBatch(ctx context.Context, batch domain.Batch, requestKey string) (stored domain.Batch, err error)
	Batch(ctx context.Context, batchID string) (stored domain.Batch, err error)
	SaveBatch(ctx context.Context, batch domain.Batch, expectedRevision int64) (err error)
	AppendReview(ctx context.Context, review domain.Review) (err error)
	Reviews(ctx context.Context, batchID string) (items []domain.Review, err error)
	AppendAudit(ctx context.Context, event domain.AuditEvent) (err error)
	AuditHead(ctx context.Context, aggregateID string) (hash string, err error)
	Statistics(ctx context.Context, from time.Time, to time.Time) (statistics domain.ReviewStatistics, err error)
}

type ReviewClock interface{ Now() (current time.Time) }
type IdentityProvider interface{ NewID() (identity string) }

type Importer interface {
	Read(ctx context.Context, document []byte) (template domain.StandardTemplate, err error)
}

type Exporter interface {
	WriteTemplate(ctx context.Context, template domain.StandardTemplate) (path string, err error)
	WriteReviewSheet(ctx context.Context, batch domain.Batch, results []domain.MaterialResult) (path string, err error)
}
