package application

import (
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry044/internal/domain"
)

type AssessmentReport struct {
	BatchID         string
	BatchName       string
	Template        string
	TemplateVersion int
	GeneratedAt     time.Time
	Lines           []AssessmentLine
}

type AssessmentLine struct {
	MaterialID string
	Title      string
	Score      float64
	Passed     bool
	Vetoed     bool
	Health     []string
}

func BuildAssessmentReport(batch domain.Batch, reviews []domain.Review, generatedAt time.Time) AssessmentReport {
	report := AssessmentReport{BatchID: batch.ID, BatchName: batch.Name, GeneratedAt: generatedAt}
	if batch.Snapshot == nil {
		return report
	}
	report.Template, report.TemplateVersion = batch.Snapshot.TemplateID, batch.Snapshot.Version
	for _, material := range batch.Materials {
		result, err := domain.Aggregate(*batch.Snapshot, material.ID, reviews)
		if err != nil {
			continue
		}
		line := AssessmentLine{MaterialID: material.ID, Title: material.Title, Score: result.WeightedScore, Passed: result.Passed, Vetoed: result.Vetoed}
		for _, ledger := range domain.BuildCriterionLedger(*batch.Snapshot, reviews, material.ID) {
			line.Health = append(line.Health, ledger.CriterionID+":"+ledger.Health())
		}
		report.Lines = append(report.Lines, line)
	}
	sort.Slice(report.Lines, func(i, j int) bool {
		return strings.ToLower(report.Lines[i].MaterialID) < strings.ToLower(report.Lines[j].MaterialID)
	})
	return report
}
