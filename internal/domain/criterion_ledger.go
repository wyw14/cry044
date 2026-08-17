package domain

import (
	"sort"
	"strings"
)

// CriterionLedger is the explainable counterpart of a numeric total. It is
// deliberately kept as a domain object so statistics and printable sheets
// use the same evidence rather than reimplementing score math in handlers.
type CriterionLedger struct {
	CriterionID       string
	Title             string
	Uses              int
	Vetoes            int
	Disagreements     int
	Average           float64
	EvidenceCoverage  float64
	ReviewerBreakdown map[string]int
}

func BuildCriterionLedger(snapshot TemplateSnapshot, reviews []Review, materialID string) []CriterionLedger {
	ledgers := make(map[string]*CriterionLedger, len(snapshot.Criteria))
	for _, criterion := range snapshot.Criteria {
		ledgers[criterion.ID] = &CriterionLedger{CriterionID: criterion.ID, Title: criterion.Title, ReviewerBreakdown: map[string]int{}}
	}
	for _, review := range reviews {
		if review.MaterialID != materialID || review.SubmittedAt == nil {
			continue
		}
		for _, score := range review.Scores {
			ledger := ledgers[score.CriterionID]
			if ledger == nil {
				continue
			}
			ledger.Uses++
			ledger.Average += float64(score.Value)
			ledger.ReviewerBreakdown[review.ReviewerID] = score.Value
			if score.VetoTriggered {
				ledger.Vetoes++
			}
			if len(score.EvidenceIDs) > 0 {
				ledger.EvidenceCoverage++
			}
		}
	}
	for _, ledger := range ledgers {
		if ledger.Uses > 0 {
			ledger.Average /= float64(ledger.Uses)
			ledger.EvidenceCoverage /= float64(ledger.Uses)
		}
		values := make([]int, 0, len(ledger.ReviewerBreakdown))
		for _, value := range ledger.ReviewerBreakdown {
			values = append(values, value)
		}
		if len(values) > 1 {
			sort.Ints(values)
			if values[len(values)-1]-values[0] > 2 {
				ledger.Disagreements = 1
			}
		}
	}
	result := make([]CriterionLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		result = append(result, *ledger)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].CriterionID) < strings.ToLower(result[j].CriterionID)
	})
	return result
}

func (l CriterionLedger) Health() string {
	if l.Vetoes > 0 {
		return "veto"
	}
	if l.Disagreements > 0 {
		return "disputed"
	}
	if l.Uses == 0 {
		return "unused"
	}
	if l.EvidenceCoverage < 1 {
		return "evidence_gap"
	}
	return "stable"
}
