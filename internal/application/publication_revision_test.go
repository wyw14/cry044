package application_test

import (
	"context"
	"errors"
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

func TestPublishRejectsStaleRevisionInsteadOfNormalizing(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := repository.NewReviewArchive()
	template := domain.StandardTemplate{ID: "quality", Version: 1, Revision: 1, Status: domain.TemplateReview, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	_, _ = store.SaveTemplate(ctx, template, 0, "template")
	svc := application.NewReviewService(store, clock{now}, &service.ReviewIdentitySequence{})

	t.Run("stale revision below current is rejected without normalization", func(t *testing.T) {
		before, _ := store.Template(ctx, "quality", 1)
		err := svc.PublishTemplate(ctx, "quality", 1, before.Revision-1, "chair")
		if err == nil {
			t.Fatal("stale publication was accepted")
		}
		if !errors.Is(err, application.ErrStalePublicationRevision) {
			t.Fatalf("expected ErrStalePublicationRevision, got %v", err)
		}
		after, _ := store.Template(ctx, "quality", 1)
		if after.Revision != before.Revision {
			t.Fatalf("revision jumped from %d to %d", before.Revision, after.Revision)
		}
		if after.Status != domain.TemplateReview {
			t.Fatalf("status changed to %s", after.Status)
		}
	})

	t.Run("future revision above current is rejected without normalization", func(t *testing.T) {
		before, _ := store.Template(ctx, "quality", 1)
		err := svc.PublishTemplate(ctx, "quality", 1, before.Revision+1, "chair")
		if err == nil {
			t.Fatal("future publication was accepted")
		}
		if !errors.Is(err, application.ErrStalePublicationRevision) {
			t.Fatalf("expected ErrStalePublicationRevision, got %v", err)
		}
		after, _ := store.Template(ctx, "quality", 1)
		if after.Revision != before.Revision {
			t.Fatalf("revision jumped from %d to %d", before.Revision, after.Revision)
		}
		if after.Status != domain.TemplateReview {
			t.Fatalf("status changed to %s", after.Status)
		}
	})

	t.Run("current revision publishes and remains immutable", func(t *testing.T) {
		current, _ := store.Template(ctx, "quality", 1)
		if err := svc.PublishTemplate(ctx, "quality", 1, current.Revision, "chair"); err != nil {
			t.Fatalf("publish with current revision failed: %v", err)
		}
		published, _ := store.Template(ctx, "quality", 1)
		if published.Status != domain.TemplatePublished {
			t.Fatalf("status=%s", published.Status)
		}
		if published.Revision != current.Revision+1 {
			t.Fatalf("revision=%d, expected %d", published.Revision, current.Revision+1)
		}
		if err := svc.PublishTemplate(ctx, "quality", 1, published.Revision, "chair"); err == nil {
			t.Fatal("already-published template was republished")
		}
	})
}

