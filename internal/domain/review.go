package domain

import (
	"errors"
	"maps"
	"math"
	"slices"
	"time"
)

type BatchStatus string

const (
	BatchDraft     BatchStatus = "draft"
	BatchSubmitted BatchStatus = "submitted"
	BatchReturned  BatchStatus = "returned"
	BatchReReview  BatchStatus = "re_review"
	BatchCompleted BatchStatus = "completed"
	BatchVoided    BatchStatus = "voided"
)

var ErrInvalidBatchFlow = errors.New("invalid review batch flow")
var ErrSnapshotMissing = errors.New("review batch requires template snapshot")
var ErrIncompleteReview = errors.New("required criteria are incomplete")

type Material struct {
	ID        string
	Title     string
	LocalPath string
	Checksum  string
}
type Assignment struct {
	ReviewerID string
	MaterialID string
	AssignedAt time.Time
	DueAt      time.Time
}
type Score struct {
	CriterionID   string
	Value         int
	Comment       string
	EvidenceIDs   []string
	VetoTriggered bool
}
type Review struct {
	Scores      []Score
	SubmittedAt *time.Time
	ID          string
	BatchID     string
	MaterialID  string
	ReviewerID  string
	Opinion     string
}

func (review Review) MatchesSubmission(other Review) bool {
	return review.BatchID == other.BatchID && review.MaterialID == other.MaterialID && review.ReviewerID == other.ReviewerID && review.SubmittedAt != nil
}

type Batch struct {
	Snapshot     *TemplateSnapshot
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Materials    []Material
	Assignments  []Assignment
	Reviews      []Review
	Decisions    []FinalDecision
	Results      []MaterialResult
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       BatchStatus
	ID           string
	ProjectID    string
	Name         string
	ReturnReason string
	Revision     int64
}

func cloneSnapshot(source *TemplateSnapshot) *TemplateSnapshot {
	if source == nil {
		return nil
	}
	copy := *source
	copy.Variables = maps.Clone(source.Variables)
	copy.Criteria = make([]Criterion, len(source.Criteria))
	for index, criterion := range source.Criteria {
		copy.Criteria[index] = copyCriterion(criterion)
	}
	return &copy
}

func (batch Batch) Clone() Batch {
	copy := batch
	copy.Snapshot = cloneSnapshot(batch.Snapshot)
	copy.Materials = slices.Clone(batch.Materials)
	copy.Assignments = slices.Clone(batch.Assignments)
	copy.Reviews = slices.Clone(batch.Reviews)
	copy.Decisions = slices.Clone(batch.Decisions)
	copy.Results = slices.Clone(batch.Results)
	return copy
}

func (batch *Batch) Start(snapshot TemplateSnapshot, now time.Time) error {
	if batch.Status != BatchDraft {
		return ErrInvalidBatchFlow
	}
	batch.Snapshot = cloneSnapshot(&snapshot)
	batch.Status, batch.StartedAt = BatchSubmitted, &now
	batch.Revision += 1
	batch.UpdatedAt = now
	return nil
}

func permittedBatchMove(from, to BatchStatus) bool {
	switch from {
	case BatchSubmitted:
		return to == BatchReturned || to == BatchCompleted || to == BatchVoided
	case BatchReturned:
		return to == BatchReReview || to == BatchVoided
	case BatchReReview:
		return to == BatchReturned || to == BatchCompleted || to == BatchVoided
	default:
		return false
	}
}

func (batch *Batch) Transition(target BatchStatus, reason string, now time.Time) error {
	if batch.Snapshot == nil {
		return ErrSnapshotMissing
	}
	if !permittedBatchMove(batch.Status, target) {
		return ErrInvalidBatchFlow
	}
	if target == BatchReturned && reason == "" {
		return errors.New("return reason required")
	}
	batch.Status, batch.ReturnReason = target, reason
	batch.Revision += 1
	batch.UpdatedAt = now
	if target == BatchCompleted {
		batch.CompletedAt = &now
	}
	return nil
}

type MaterialResult struct {
	MaterialID    string   `json:"materialID"`
	WeightedScore float64  `json:"weightedScore"`
	Vetoed        bool     `json:"vetoed"`
	Passed        bool     `json:"passed"`
	ReviewerCount int      `json:"reviewerCount"`
	Disagreements []string `json:"disagreements"`
}

type criterionTally struct {
	sum     int
	count   int
	minimum int
	maximum int
	vetoed  bool
}
type materialScoreboard struct {
	criteria  map[string]Criterion
	tallies   map[string]criterionTally
	reviewers map[string]struct{}
}

func newScoreboard(snapshot TemplateSnapshot) materialScoreboard {
	criteria := make(map[string]Criterion, len(snapshot.Criteria))
	for _, criterion := range snapshot.Criteria {
		criteria[criterion.ID] = criterion
	}
	return materialScoreboard{criteria: criteria, tallies: make(map[string]criterionTally), reviewers: make(map[string]struct{})}
}

func (board *materialScoreboard) include(review Review, materialID string) {
	if review.MaterialID != materialID || review.SubmittedAt == nil {
		return
	}
	board.reviewers[review.ReviewerID] = struct{}{}
	for _, score := range review.Scores {
		criterion, exists := board.criteria[score.CriterionID]
		if !exists {
			continue
		}
		tally := board.tallies[score.CriterionID]
		if tally.count == 0 || score.Value < tally.minimum {
			tally.minimum = score.Value
		}
		if tally.count == 0 || score.Value > tally.maximum {
			tally.maximum = score.Value
		}
		tally.sum += score.Value
		tally.count += 1
		tally.vetoed = tally.vetoed || (criterion.Veto && score.VetoTriggered)
		board.tallies[score.CriterionID] = tally
	}
}

func (board materialScoreboard) result(materialID string) (MaterialResult, error) {
	weighted, capacity, vetoed := 0.0, 0.0, false
	disagreements := make([]string, 0)
	for id, criterion := range board.criteria {
		tally := board.tallies[id]
		if criterion.Required && tally.count == 0 {
			return MaterialResult{}, ErrIncompleteReview
		}
		if tally.count == 0 {
			continue
		}
		weighted += (float64(tally.sum) / float64(tally.count)) * criterion.Weight
		capacity += float64(criterion.Scale.Max) * criterion.Weight
		vetoed = vetoed || tally.vetoed
		if tally.count > 1 && tally.maximum-tally.minimum > 2 {
			disagreements = append(disagreements, id)
		}
	}
	percentage := 0.0
	if capacity > 0 {
		percentage = math.Round(weighted/capacity*10000) / 100
	}
	slices.Sort(disagreements)
	return MaterialResult{MaterialID: materialID, WeightedScore: percentage, Vetoed: vetoed, Passed: percentage >= 60 && !vetoed, ReviewerCount: len(board.reviewers), Disagreements: disagreements}, nil
}

func Aggregate(snapshot TemplateSnapshot, materialID string, reviews []Review) (MaterialResult, error) {
	board := newScoreboard(snapshot)
	for _, review := range reviews {
		board.include(review, materialID)
	}
	return board.result(materialID)
}

type AuditEvent struct {
	At           time.Time
	ID           string
	AggregateID  string
	ActorID      string
	Action       string
	Detail       string
	PreviousHash string
	Hash         string
}
