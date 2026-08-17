package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry044/internal/application"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
	"github.com/wyw14/cry044/internal/service"
)

type decisionClock struct{ now time.Time }

func (c decisionClock) Now() time.Time { return c.now }

func TestConcurrentFinalDecisionUsesOneBatchRevision(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	store := repository.NewReviewArchive()
	template := domain.StandardTemplate{ID: "quality", Version: 1, Status: domain.TemplatePublished, Revision: 1, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	if _, err := store.SaveTemplate(ctx, template, 0, "template"); err != nil {
		t.Fatal(err)
	}
	batch := domain.Batch{ID: "batch", Status: domain.BatchDraft, Revision: 1, Materials: []domain.Material{{ID: "part"}}, Assignments: []domain.Assignment{{ReviewerID: "r1", MaterialID: "part"}, {ReviewerID: "r2", MaterialID: "part"}}, CreatedAt: now, UpdatedAt: now}
	if _, err := store.CreateBatch(ctx, batch, "batch"); err != nil {
		t.Fatal(err)
	}
	clock := decisionClock{now}
	serviceLayer := application.NewReviewService(store, clock, &service.ReviewIdentitySequence{})
	if err := serviceLayer.StartBatch(ctx, "batch", "quality", 1, nil, "coordinator"); err != nil {
		t.Fatal(err)
	}
	if err := serviceLayer.SubmitReview(ctx, "batch", "part", "r1", []domain.Score{{CriterionID: "quality", Value: 4}}, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := serviceLayer.SubmitReview(ctx, "batch", "part", "r2", []domain.Score{{CriterionID: "quality", Value: 4}}, "ok"); err != nil {
		t.Fatal(err)
	}
	current, err := store.Batch(ctx, "batch")
	if err != nil {
		t.Fatal(err)
	}
	deliberation := application.NewDeliberationService(store, clock, domain.PanelPolicy{MinimumReviewers: 2, DivergenceThreshold: 2})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for _, conclusion := range []string{"first", "second"} {
		go func(conclusion string) {
			defer wg.Done()
			_, callErr := deliberation.Resolve(ctx, "batch", "part", conclusion, current.Revision, true, conclusion, "agreed")
			results <- callErr
		}(conclusion)
	}
	wg.Wait()
	close(results)
	succeeded := 0
	for callErr := range results {
		if callErr == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("expected exactly one concurrent decision to commit, got %d", succeeded)
	}
}
