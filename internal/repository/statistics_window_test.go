package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
	"testing"
	"time"
)

func TestStatisticsWindowExcludesStaleReturnsAndScores(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(31 * 24 * time.Hour)
	submitted := from.Add(10 * 24 * time.Hour)
	store := NewReviewArchive()
	snapshot := domain.TemplateSnapshot{TemplateID: "quality", Version: 1, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	recent := domain.Batch{ID: "recent", Status: domain.BatchReturned, Revision: 1, Snapshot: &snapshot, ReturnReason: "evidence missing", UpdatedAt: submitted}
	stale := domain.Batch{ID: "stale", Status: domain.BatchReturned, Revision: 1, Snapshot: &snapshot, ReturnReason: "legacy reason", UpdatedAt: from.Add(-24 * time.Hour)}
	_, _ = store.CreateBatch(ctx, recent, "recent")
	_, _ = store.CreateBatch(ctx, stale, "stale")
	_ = store.AppendReview(ctx, domain.Review{ID: "new-review", BatchID: "recent", MaterialID: "part", ReviewerID: "new", Scores: []domain.Score{{CriterionID: "quality", Value: 4}}, SubmittedAt: &submitted})
	oldSubmitted := from.Add(-48 * time.Hour)
	_ = store.AppendReview(ctx, domain.Review{ID: "old-review", BatchID: "stale", MaterialID: "part", ReviewerID: "old", Scores: []domain.Score{{CriterionID: "quality", Value: 1, VetoTriggered: true}}, SubmittedAt: &oldSubmitted})
	statistics, err := store.Statistics(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(statistics.ReturnReasons) != 1 || statistics.ReturnReasons[0].Reason != "evidence missing" {
		t.Fatalf("stale return reason leaked into window: %#v", statistics.ReturnReasons)
	}
	if len(statistics.Criteria) != 1 || statistics.Criteria[0].UseCount != 1 || statistics.Criteria[0].VetoCount != 0 {
		t.Fatalf("stale scores leaked into window: %#v", statistics.Criteria)
	}
}
