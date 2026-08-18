package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry044/internal/application"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
	"github.com/wyw14/cry044/internal/service"
)

func TestReviewerCannotSubmitTheSameMaterialTwice(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := repository.NewReviewArchive()
	_, _ = store.CreateBatch(ctx, domain.Batch{ID: "batch", Status: domain.BatchSubmitted, Revision: 2, Materials: []domain.Material{{ID: "material"}}, Assignments: []domain.Assignment{{ReviewerID: "reviewer", MaterialID: "material"}}}, "create")
	svc := application.NewReviewService(store, clock{now}, &service.ReviewIdentitySequence{})
	score := []domain.Score{{CriterionID: "quality", Value: 4}}
	if err := svc.SubmitReview(ctx, "batch", "material", "reviewer", score, "first"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SubmitReview(ctx, "batch", "material", "reviewer", score, "second"); err == nil {
		t.Fatal("duplicate review was accepted")
	}
	reviews, err := store.Reviews(ctx, "batch")
	if err != nil || len(reviews) != 1 {
		t.Fatalf("reviews=%d err=%v", len(reviews), err)
	}
}
