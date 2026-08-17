package domain

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var ErrPanelQuorumMissing = errors.New("review panel quorum is not met")
var ErrDivergenceUnresolved = errors.New("score divergence requires a resolution rationale")
var ErrVetoCannotBeOverruled = errors.New("veto criterion cannot be overruled by final score")

type PanelPolicy struct {
	MinimumReviewers    int
	DivergenceThreshold int
}
type ReviewerPosition struct {
	ReviewerID string   `json:"reviewerID"`
	Score      int      `json:"score"`
	Comment    string   `json:"comment"`
	Evidence   []string `json:"evidence"`
}
type DivergenceCase struct {
	CriterionID string             `json:"criterionID"`
	Spread      int                `json:"spread"`
	Positions   []ReviewerPosition `json:"positions"`
}

type DeliberationCase struct {
	BatchID            string           `json:"batchID"`
	MaterialID         string           `json:"materialID"`
	TemplateID         string           `json:"templateID"`
	TemplateVersion    int              `json:"templateVersion"`
	AssignedReviewers  int              `json:"assignedReviewers"`
	SubmittedReviewers int              `json:"submittedReviewers"`
	QuorumMet          bool             `json:"quorumMet"`
	Preliminary        MaterialResult   `json:"preliminary"`
	Divergences        []DivergenceCase `json:"divergences"`
	Decision           *FinalDecision   `json:"decision,omitempty"`
}

type FinalDecision struct {
	ResolvedAt time.Time `json:"resolvedAt"`
	MaterialID string    `json:"materialID"`
	Conclusion string    `json:"conclusion"`
	Rationale  string    `json:"rationale"`
	ResolvedBy string    `json:"resolvedBy"`
	Passed     bool      `json:"passed"`
}

type panelCensus struct {
	assigned  map[string]struct{}
	submitted map[string]struct{}
}

func countPanel(batch Batch, materialID string, reviews []Review) panelCensus {
	census := panelCensus{assigned: make(map[string]struct{}), submitted: make(map[string]struct{})}
	for _, seat := range batch.Assignments {
		if seat.MaterialID == materialID {
			census.assigned[seat.ReviewerID] = struct{}{}
		}
	}
	for _, review := range reviews {
		_, belongs := census.assigned[review.ReviewerID]
		if belongs && review.MaterialID == materialID && review.SubmittedAt != nil {
			census.submitted[review.ReviewerID] = struct{}{}
		}
	}
	return census
}

func normalizedPolicy(policy PanelPolicy) PanelPolicy {
	if policy.MinimumReviewers < 1 {
		policy.MinimumReviewers = 2
	}
	if policy.DivergenceThreshold < 1 {
		policy.DivergenceThreshold = 2
	}
	return policy
}

func BuildDeliberation(batch Batch, materialID string, reviews []Review, policy PanelPolicy) (DeliberationCase, error) {
	if batch.Snapshot == nil {
		return DeliberationCase{}, ErrSnapshotMissing
	}
	policy = normalizedPolicy(policy)
	panel := countPanel(batch, materialID, reviews)
	preliminary, err := Aggregate(*batch.Snapshot, materialID, reviews)
	if err != nil {
		return DeliberationCase{}, err
	}
	caseFile := DeliberationCase{BatchID: batch.ID, MaterialID: materialID, TemplateID: batch.Snapshot.TemplateID, TemplateVersion: batch.Snapshot.Version, AssignedReviewers: len(panel.assigned), SubmittedReviewers: len(panel.submitted), QuorumMet: len(panel.submitted) >= policy.MinimumReviewers, Preliminary: preliminary}
	for _, criterion := range batch.Snapshot.Criteria {
		positions := positionsFor(reviews, materialID, criterion.ID, panel.assigned)
		spread := positionSpread(positions)
		if spread > policy.DivergenceThreshold {
			caseFile.Divergences = append(caseFile.Divergences, DivergenceCase{CriterionID: criterion.ID, Spread: spread, Positions: positions})
		}
	}
	for _, decision := range batch.Decisions {
		if decision.MaterialID == materialID {
			copy := decision
			caseFile.Decision = &copy
			break
		}
	}
	return caseFile, nil
}

func (batch *Batch) RecordDecision(caseFile DeliberationCase, pass bool, conclusion, rationale, resolver string, at time.Time) error {
	switch {
	case batch.Status != BatchSubmitted && batch.Status != BatchReReview:
		return ErrInvalidBatchFlow
	case !caseFile.QuorumMet:
		return ErrPanelQuorumMissing
	case caseFile.Preliminary.Vetoed && pass:
		return ErrVetoCannotBeOverruled
	case len(caseFile.Divergences) > 0 && strings.TrimSpace(rationale) == "":
		return ErrDivergenceUnresolved
	case strings.TrimSpace(conclusion) == "" || strings.TrimSpace(resolver) == "":
		return errors.New("conclusion and resolver are required")
	}
	decision := FinalDecision{MaterialID: caseFile.MaterialID, Passed: pass, Conclusion: conclusion, Rationale: rationale, ResolvedBy: resolver, ResolvedAt: at}
	replaced := false
	for index := range batch.Decisions {
		if batch.Decisions[index].MaterialID == caseFile.MaterialID {
			batch.Decisions[index], replaced = decision, true
			break
		}
	}
	if !replaced {
		batch.Decisions = append(batch.Decisions, decision)
	}
	batch.Revision += 1
	batch.UpdatedAt = at
	return nil
}

func positionsFor(reviews []Review, materialID, criterionID string, assigned map[string]struct{}) []ReviewerPosition {
	positions := make([]ReviewerPosition, 0)
	for _, review := range reviews {
		if _, ok := assigned[review.ReviewerID]; !ok || review.MaterialID != materialID || review.SubmittedAt == nil {
			continue
		}
		for _, score := range review.Scores {
			if score.CriterionID == criterionID {
				positions = append(positions, ReviewerPosition{ReviewerID: review.ReviewerID, Score: score.Value, Comment: score.Comment, Evidence: slices.Clone(score.EvidenceIDs)})
			}
		}
	}
	slices.SortFunc(positions, func(left, right ReviewerPosition) int { return strings.Compare(left.ReviewerID, right.ReviewerID) })
	return positions
}

func positionSpread(positions []ReviewerPosition) int {
	if len(positions) < 2 {
		return 0
	}
	minimum, maximum := positions[0].Score, positions[0].Score
	for _, position := range positions[1:] {
		minimum = min(minimum, position.Score)
		maximum = max(maximum, position.Score)
	}
	return maximum - minimum
}
