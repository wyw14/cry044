package application_test

import (
	"context"
	"github.com/wyw14/cry044/internal/application"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
	"github.com/wyw14/cry044/internal/service"
	"testing"
	"time"
)

type clock struct{ v time.Time }

func (c clock) Now() time.Time { return c.v }
func TestBatchSnapshotSurvivesLaterTemplateVersion(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	repo := repository.NewReviewArchive()
	template := domain.StandardTemplate{ID: "t", Version: 1, Name: "one", Status: domain.TemplatePublished, Revision: 1, Criteria: []domain.Criterion{{ID: "c", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	_, _ = repo.SaveTemplate(ctx, template, 0, "t1")
	batch := domain.Batch{ID: "b", Status: domain.BatchDraft, Revision: 1, Materials: []domain.Material{{ID: "m"}}, Assignments: []domain.Assignment{{ReviewerID: "r", MaterialID: "m"}}, CreatedAt: now, UpdatedAt: now}
	_, _ = repo.CreateBatch(ctx, batch, "b")
	svc := application.NewReviewService(repo, clock{now}, &service.ReviewIdentitySequence{})
	if err := svc.StartBatch(ctx, "b", "t", 1, nil, "c"); err != nil {
		t.Fatal(err)
	}
	next := template.CloneAs("t", 2, now)
	next.Name = "two"
	next.Status = domain.TemplatePublished
	_, _ = repo.SaveTemplate(ctx, next, 0, "t2")
	stored, _ := repo.Batch(ctx, "b")
	if stored.Snapshot.Name != "one" || stored.Snapshot.Version != 1 {
		t.Fatalf("snapshot %#v", stored.Snapshot)
	}
}
