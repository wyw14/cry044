package domain

import (
	"testing"
	"time"
)

func TestCriterionLedgerLabelsEvidenceGapsAndDisagreements(t *testing.T) {
	now := time.Now()
	snapshot := TemplateSnapshot{Criteria: []Criterion{{ID: "trace", Title: "证据", Required: true}, {ID: "safety", Title: "安全", Veto: true}}}
	reviews := []Review{{MaterialID: "m", ReviewerID: "a", SubmittedAt: &now, Scores: []Score{{CriterionID: "trace", Value: 1, EvidenceIDs: []string{"photo"}}}}, {MaterialID: "m", ReviewerID: "b", SubmittedAt: &now, Scores: []Score{{CriterionID: "trace", Value: 5}}}, {MaterialID: "m", ReviewerID: "a", SubmittedAt: &now, Scores: []Score{{CriterionID: "safety", VetoTriggered: true}}}}
	ledgers := BuildCriterionLedger(snapshot, reviews, "m")
	if len(ledgers) != 2 {
		t.Fatalf("ledgers %#v", ledgers)
	}
	for _, ledger := range ledgers {
		if ledger.CriterionID == "trace" && ledger.Health() != "disputed" {
			t.Fatalf("trace ledger %#v", ledger)
		}
		if ledger.CriterionID == "safety" && ledger.Health() != "veto" {
			t.Fatalf("safety ledger %#v", ledger)
		}
	}
}
