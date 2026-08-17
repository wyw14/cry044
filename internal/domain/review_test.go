package domain

import (
	"errors"
	"testing"
	"time"
)

func TestStartedBatchUsesImmutableSnapshot(t *testing.T) {
	now := time.Now()
	template := StandardTemplate{ID: "t", Version: 1, Criteria: []Criterion{{ID: "quality", Title: "质量", Required: true, Weight: 1, Scale: Scale{Max: 5}}}}
	batch := Batch{Status: BatchDraft, Revision: 1}
	snap := Snapshot(template, nil, now)
	if err := batch.Start(snap, now); err != nil {
		t.Fatal(err)
	}
	template.Criteria[0].Title = "changed"
	if batch.Snapshot.Criteria[0].Title != "质量" {
		t.Fatal("snapshot mutated")
	}
}
func TestVetoOverridesHighScore(t *testing.T) {
	now := time.Now()
	snapshot := TemplateSnapshot{Criteria: []Criterion{{ID: "score", Required: true, Weight: 1, Scale: Scale{Max: 5}}, {ID: "veto", Required: true, Veto: true, Weight: 0, Scale: Scale{Max: 1}}}}
	reviews := []Review{{MaterialID: "m", ReviewerID: "r", SubmittedAt: &now, Scores: []Score{{CriterionID: "score", Value: 5}, {CriterionID: "veto", Value: 1, VetoTriggered: true}}}}
	result, err := Aggregate(snapshot, "m", reviews)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || !result.Vetoed || result.WeightedScore != 100 {
		t.Fatalf("result %#v", result)
	}
}
func TestRequiredCriteriaAndDisagreement(t *testing.T) {
	now := time.Now()
	snapshot := TemplateSnapshot{Criteria: []Criterion{{ID: "c", Required: true, Weight: 1, Scale: Scale{Max: 5}}}}
	if _, err := Aggregate(snapshot, "m", nil); !errors.Is(err, ErrIncompleteReview) {
		t.Fatalf("got %v", err)
	}
	reviews := []Review{{MaterialID: "m", ReviewerID: "a", SubmittedAt: &now, Scores: []Score{{CriterionID: "c", Value: 1}}}, {MaterialID: "m", ReviewerID: "b", SubmittedAt: &now, Scores: []Score{{CriterionID: "c", Value: 5}}}}
	result, err := Aggregate(snapshot, "m", reviews)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Disagreements) != 1 || result.ReviewerCount != 2 {
		t.Fatalf("result %#v", result)
	}
}
func TestPublishedTemplateCannotChangeInPlace(t *testing.T) {
	template := StandardTemplate{Status: TemplatePublished}
	if err := template.ReplaceCriteria(nil, time.Now()); !errors.Is(err, ErrPublishedTemplateImmutable) {
		t.Fatalf("got %v", err)
	}
}
