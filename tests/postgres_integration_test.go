package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/repository"
)

func TestTemplateVersionRoundTrip(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := repository.NewPGLedger(pool)
	id := "integration-" + time.Now().Format("150405.000000")
	want := domain.StandardTemplate{ID: id, Version: 1, Name: "集成标准", Status: domain.TemplateDraft, Revision: 1, Criteria: []domain.Criterion{{ID: "quality", Required: true, Weight: 1, Scale: domain.Scale{Max: 5}}}}
	got, err := store.SaveTemplate(ctx, want, 0, id)
	if err != nil || got.ID != want.ID || got.Version != 1 {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}
