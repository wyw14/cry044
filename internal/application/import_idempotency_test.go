package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
)

type importClock struct{ now time.Time }

func (c importClock) Now() time.Time { return c.now }

type varyingImporter struct{}

func (varyingImporter) Read(_ context.Context, payload []byte) (domain.StandardTemplate, error) {
	id := "draft-one"
	if string(payload) == "second" {
		id = "draft-two"
	}
	return domain.StandardTemplate{ID: id, Version: 1, Status: domain.TemplatePublished, Revision: 1, Criteria: []domain.Criterion{{ID: "quality", Scale: domain.Scale{Max: 5}}}}, nil
}

type discardExporter struct{}

func (discardExporter) WriteTemplate(context.Context, domain.StandardTemplate) (string, error) {
	return "", nil
}
func (discardExporter) WriteReviewSheet(context.Context, domain.Batch, []domain.MaterialResult) (string, error) {
	return "", nil
}

func TestImportWithoutHeaderDoesNotCollapseDistinctPayloads(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewReviewArchive()
	service := NewTransferService(repo, importClock{time.Date(2026, 8, 18, 11, 0, 0, 0, time.UTC)}, nil, varyingImporter{}, discardExporter{})
	first, err := service.ImportTemplate(ctx, []byte("first"), "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.ImportTemplate(ctx, []byte("second"), "")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatalf("empty idempotency header collapsed distinct imports: first=%q second=%q", first.ID, second.ID)
	}
}
