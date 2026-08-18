package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/wyw14/cry044/internal/domain"
)

type ReviewService struct {
	repository ReviewRepository
	clock      ReviewClock
	identities IdentityProvider
}

func NewReviewService(repository ReviewRepository, clock ReviewClock, identities IdentityProvider) *ReviewService {
	return &ReviewService{repository: repository, clock: clock, identities: identities}
}

func (s *ReviewService) CloneTemplate(ctx context.Context, sourceID string, version int, targetID, actor string) (domain.StandardTemplate, error) {
	source, err := s.repository.Template(ctx, sourceID, version)
	if err != nil {
		return source, err
	}
	draft := source.CloneAs(targetID, 1, s.clock.Now())
	requestKey := strings.Join([]string{"clone", sourceID, strconv.Itoa(version), targetID}, ":")
	stored, err := s.repository.SaveTemplate(ctx, draft, 0, requestKey)
	if err == nil {
		err = s.audit(ctx, targetID, actor, "template.cloned", sourceID)
	}
	return stored, err
}

func (s *ReviewService) PublishTemplate(ctx context.Context, id string, version int, expected int64, actor string) error {
	template, err := s.repository.Template(ctx, id, version)
	if err != nil {
		return err
	}
	if err = template.Transition(domain.TemplatePublished, false, s.clock.Now()); err != nil {
		return err
	}
	key := strings.Join([]string{"publish", id, strconv.Itoa(version)}, ":")
	if _, err = s.repository.SaveTemplate(ctx, template, expected, key); err != nil {
		return err
	}
	return s.audit(ctx, id, actor, "template.published", strconv.Itoa(version))
}

func (s *ReviewService) StartBatch(ctx context.Context, batchID, templateID string, version int, values map[string]string, actor string) error {
	batch, err := s.repository.Batch(ctx, batchID)
	if err != nil {
		return err
	}
	template, err := s.repository.Template(ctx, templateID, version)
	if err != nil {
		return err
	}
	if template.Status != domain.TemplatePublished {
		return errors.New("published template required")
	}
	when := s.clock.Now()
	if err = batch.Start(domain.Snapshot(template, values, when), when); err != nil {
		return err
	}
	if err = s.repository.SaveBatch(ctx, batch, batch.Revision-1); err != nil {
		return err
	}
	return s.audit(ctx, batchID, actor, "batch.started", templateID+strconv.Itoa(version))
}

func (s *ReviewService) SubmitReview(ctx context.Context, batchID, materialID, reviewer string, scores []domain.Score, opinion string) error {
	batch, err := s.repository.Batch(ctx, batchID)
	if err != nil {
		return err
	}
	if batch.Status != domain.BatchSubmitted && batch.Status != domain.BatchReReview {
		return domain.ErrInvalidBatchFlow
	}
	if !assignedTo(batch, materialID, reviewer) {
		return errors.New("reviewer is not assigned")
	}
	prior, err := s.repository.Reviews(ctx, batchID)
	if err != nil {
		return err
	}
	for _, review := range prior {
		if review.MaterialID == materialID && review.ReviewerID == reviewer && review.SubmittedAt != nil {
			return errors.New("review already submitted for material")
		}
	}
	opinion = strings.TrimSpace(opinion)
	submitted := s.clock.Now()
	entry := domain.Review{ID: s.identities.NewID(), BatchID: batchID, MaterialID: materialID, ReviewerID: reviewer, Scores: scores, Opinion: opinion, SubmittedAt: &submitted}
	if err = s.repository.AppendReview(ctx, entry); err != nil {
		return err
	}
	return s.audit(ctx, batchID, reviewer, "review.submitted", materialID)
}

func (s *ReviewService) Complete(ctx context.Context, batchID, actor string) ([]domain.MaterialResult, error) {
	batch, err := s.repository.Batch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	reviews, err := s.repository.Reviews(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch.Snapshot == nil {
		return nil, domain.ErrSnapshotMissing
	}
	results := make([]domain.MaterialResult, 0, len(batch.Materials))
	for _, material := range batch.Materials {
		result, aggregateErr := domain.Aggregate(*batch.Snapshot, material.ID, reviews)
		if aggregateErr != nil {
			return nil, aggregateErr
		}
		decision, found := decisionFor(batch.Decisions, material.ID)
		if (result.Vetoed || len(result.Disagreements) != 0) && !found {
			return nil, errors.New("veto or disagreement requires a final deliberation decision")
		}
		if found {
			result.Passed = decision.Passed && !result.Vetoed
		}
		results = append(results, result)
	}
	expected := batch.Revision
	batch.Results = append([]domain.MaterialResult(nil), results...)
	if err = batch.Transition(domain.BatchCompleted, "", s.clock.Now()); err != nil {
		return nil, err
	}
	if err = s.repository.SaveBatch(ctx, batch, expected); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, batchID, actor, "batch.completed", strconv.Itoa(len(results)))
	return results, nil
}

func assignedTo(batch domain.Batch, materialID, reviewer string) bool {
	for _, assignment := range batch.Assignments {
		if assignment.MaterialID == materialID && assignment.ReviewerID == reviewer {
			return true
		}
	}
	return false
}

func decisionFor(decisions []domain.FinalDecision, materialID string) (domain.FinalDecision, bool) {
	for _, decision := range decisions {
		if decision.MaterialID == materialID {
			return decision, true
		}
	}
	return domain.FinalDecision{}, false
}

func (s *ReviewService) audit(ctx context.Context, aggregateID, actor, action, detail string) error {
	previous, _ := s.repository.AuditHead(ctx, aggregateID)
	when := s.clock.Now()
	hasher := sha256.New()
	for _, part := range []string{previous, aggregateID, actor, action, detail, when.UTC().Format(time.RFC3339Nano)} {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}
	event := domain.AuditEvent{ID: s.identities.NewID(), AggregateID: aggregateID, ActorID: actor, Action: action, Detail: detail, PreviousHash: previous, Hash: hex.EncodeToString(hasher.Sum(nil)), At: when}
	return s.repository.AppendAudit(ctx, event)
}
