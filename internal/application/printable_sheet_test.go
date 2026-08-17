package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
)

type sheetClock struct{ now time.Time }

func (c sheetClock) Now() time.Time { return c.now }

type sheetImporter struct{}

func (sheetImporter) Read(context.Context, []byte) (domain.StandardTemplate, error) {
	return domain.StandardTemplate{}, nil
}

type sheetExporter struct{ results []domain.MaterialResult }

func (e *sheetExporter) WriteTemplate(context.Context, domain.StandardTemplate) (string, error) {
	return "", nil
}
func (e *sheetExporter) WriteReviewSheet(_ context.Context, _ domain.Batch, results []domain.MaterialResult) (string, error) {
	e.results = results
	return "sheet", nil
}

func TestPrintableSheetKeepsPanelCountsAndDisagreements(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)
	repo := repository.NewReviewArchive()
	template := domain.StandardTemplate{ID: "quality", Version: 1, Status: domain.TemplatePublished, Revision: 1, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	_, _ = repo.SaveTemplate(ctx, template, 0, "template")
	snapshot := domain.Snapshot(template, nil, now)
	submitted := now
	batch := domain.Batch{ID: "batch", Name: "first", Status: domain.BatchCompleted, Revision: 2, Snapshot: &snapshot, Materials: []domain.Material{{ID: "part"}}, Results: []domain.MaterialResult{{MaterialID: "part"}}, CreatedAt: now, UpdatedAt: now}
	_, _ = repo.CreateBatch(ctx, batch, "batch")
	_ = repo.AppendReview(ctx, domain.Review{ID: "r1", BatchID: "batch", MaterialID: "part", ReviewerID: "a", Scores: []domain.Score{{CriterionID: "quality", Value: 1}}, SubmittedAt: &submitted})
	_ = repo.AppendReview(ctx, domain.Review{ID: "r2", BatchID: "batch", MaterialID: "part", ReviewerID: "b", Scores: []domain.Score{{CriterionID: "quality", Value: 5}}, SubmittedAt: &submitted})
	exporter := &sheetExporter{}
	service := NewTransferService(repo, sheetClock{now}, nil, sheetImporter{}, exporter)
	if _, err := service.PrintReviewSheet(ctx, "batch"); err != nil {
		t.Fatal(err)
	}
	if len(exporter.results) != 1 || exporter.results[0].ReviewerCount != 2 || len(exporter.results[0].Disagreements) != 1 {
		t.Fatalf("printable data lost panel detail: %#v", exporter.results)
	}
}
