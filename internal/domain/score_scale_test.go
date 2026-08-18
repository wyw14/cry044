package domain

import (
	"testing"
	"time"
)

func TestAggregateRejectsScoresOutsideCriterionScale(t *testing.T) {
	now := time.Now()
	snapshot := TemplateSnapshot{Criteria: []Criterion{{ID: "quality", Required: true, Weight: 1, Scale: Scale{Min: 1, Max: 5}}}}
	_, err := Aggregate(snapshot, "material", []Review{{MaterialID: "material", ReviewerID: "reviewer", SubmittedAt: &now, Scores: []Score{{CriterionID: "quality", Value: 9}}}})
	if err == nil {
		t.Fatal("out-of-range score was accepted")
	}
}
