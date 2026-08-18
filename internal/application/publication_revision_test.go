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

func TestPublishRejectsStaleExpectedRevision(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := repository.NewReviewArchive()
	template := domain.StandardTemplate{ID: "quality", Version: 1, Revision: 1, Status: domain.TemplateReview, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	_, _ = store.SaveTemplate(ctx, template, 0, "template")
	svc := application.NewReviewService(store, clock{now}, &service.ReviewIdentitySequence{})
	if err := svc.PublishTemplate(ctx, "quality", 1, 0, "chair"); err == nil {
		t.Fatal("stale publication was accepted")
	}
	stored, _ := store.Template(ctx, "quality", 1)
	if stored.Status != domain.TemplateReview {
		t.Fatalf("template status changed to %s", stored.Status)
	}
}
