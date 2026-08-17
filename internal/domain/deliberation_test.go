package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDeliberationRequiresQuorumRationaleAndPreservesVeto(t *testing.T) {
	now := time.Now()
	batch := Batch{ID: "batch", Status: BatchSubmitted, Revision: 2, Snapshot: &TemplateSnapshot{TemplateID: "quality", Version: 4, Criteria: []Criterion{{ID: "score", Required: true, Weight: 1, Scale: Scale{Max: 5}}, {ID: "fatal", Required: true, Veto: true, Scale: Scale{Max: 1}}}}, Materials: []Material{{ID: "sample"}}, Assignments: []Assignment{{ReviewerID: "left", MaterialID: "sample"}, {ReviewerID: "right", MaterialID: "sample"}}}
	reviews := []Review{
		{MaterialID: "sample", ReviewerID: "left", SubmittedAt: &now, Scores: []Score{{CriterionID: "score", Value: 1}, {CriterionID: "fatal", VetoTriggered: true}}},
		{MaterialID: "sample", ReviewerID: "right", SubmittedAt: &now, Scores: []Score{{CriterionID: "score", Value: 5}, {CriterionID: "fatal"}}},
	}
	caseFile, err := BuildDeliberation(batch, "sample", reviews, PanelPolicy{MinimumReviewers: 2, DivergenceThreshold: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !caseFile.QuorumMet || len(caseFile.Divergences) != 1 || !caseFile.Preliminary.Vetoed {
		t.Fatalf("unexpected case %#v", caseFile)
	}
	if err := batch.RecordDecision(caseFile, true, "通过", "已讨论", "chair", now); !errors.Is(err, ErrVetoCannotBeOverruled) {
		t.Fatalf("veto was overruled: %v", err)
	}
	if err := batch.RecordDecision(caseFile, false, "退回", "", "chair", now); !errors.Is(err, ErrDivergenceUnresolved) {
		t.Fatalf("divergence closed without rationale: %v", err)
	}
	if err := batch.RecordDecision(caseFile, false, "退回", "复核证据后维持否决", "chair", now); err != nil {
		t.Fatal(err)
	}
}
