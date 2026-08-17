package application_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wyw14/cry044/internal/application"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
)

func TestConcurrentFinalDecisionUsesOneBatchRevision(t *testing.T) {
	now := time.Now()
	store := repository.NewReviewArchive()
	batch := domain.Batch{ID: "batch", Status: domain.BatchSubmitted, Revision: 2, Snapshot: &domain.TemplateSnapshot{TemplateID: "quality", Version: 1, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}, Materials: []domain.Material{{ID: "material"}}, Assignments: []domain.Assignment{{ReviewerID: "left", MaterialID: "material"}, {ReviewerID: "right", MaterialID: "material"}}}
	_, _ = store.CreateBatch(context.Background(), batch, "batch")
	for _, reviewer := range []string{"left", "right"} {
		review := domain.Review{ID: reviewer, BatchID: "batch", MaterialID: "material", ReviewerID: reviewer, SubmittedAt: &now, Scores: []domain.Score{{CriterionID: "quality", Value: 4}}}
		_ = store.AppendReview(context.Background(), review)
	}
	service := application.NewDeliberationService(store, clock{now}, domain.PanelPolicy{MinimumReviewers: 2, DivergenceThreshold: 2})
	start := make(chan struct{})
	var successes atomic.Int32
	var group sync.WaitGroup
	for i := 0; i < 2; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if _, err := service.Resolve(context.Background(), "batch", "material", "chair", 2, true, "通过", "评分一致"); err == nil {
				successes.Add(1)
			}
		}()
	}
	close(start)
	group.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful final decisions=%d", successes.Load())
	}
}
