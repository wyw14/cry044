package application

import (
	"context"
	"errors"
	"github.com/wyw14/cry044/internal/domain"
)

type DeliberationService struct {
	repository ReviewRepository
	clock      ReviewClock
	policy     domain.PanelPolicy
}
type loadedPanel struct {
	batch   domain.Batch
	reviews []domain.Review
}

func NewDeliberationService(repository ReviewRepository, clock ReviewClock, policy domain.PanelPolicy) *DeliberationService {
	return &DeliberationService{repository: repository, clock: clock, policy: policy}
}

func (s *DeliberationService) load(ctx context.Context, batchID string) (loadedPanel, error) {
	panel := loadedPanel{}
	batch, err := s.repository.Batch(ctx, batchID)
	if err != nil {
		return panel, err
	}
	panel.batch = batch
	reviews, err := s.repository.Reviews(ctx, batchID)
	if err != nil {
		return panel, err
	}
	panel.reviews = reviews
	return panel, nil
}

func (s *DeliberationService) Board(ctx context.Context, batchID string) ([]domain.DeliberationCase, error) {
	panel, err := s.load(ctx, batchID)
	if err != nil {
		return nil, err
	}
	board := make([]domain.DeliberationCase, 0, len(panel.batch.Materials))
	for _, material := range panel.batch.Materials {
		caseFile, buildErr := domain.BuildDeliberation(panel.batch, material.ID, panel.reviews, s.policy)
		switch {
		case buildErr == nil:
			board = append(board, caseFile)
		case errors.Is(buildErr, domain.ErrIncompleteReview):
			continue
		default:
			return nil, buildErr
		}
	}
	return board, nil
}

func (s *DeliberationService) Resolve(ctx context.Context, batchID, materialID, resolver string, expected int64, pass bool, conclusion, rationale string) (domain.FinalDecision, error) {
	panel, err := s.load(ctx, batchID)
	if err != nil {
		return domain.FinalDecision{}, err
	}
	caseFile, err := domain.BuildDeliberation(panel.batch, materialID, panel.reviews, s.policy)
	if err != nil {
		return domain.FinalDecision{}, err
	}
	if panel.batch.Revision != expected {
		return domain.FinalDecision{}, errors.New("deliberation revision conflict")
	}
	if err = panel.batch.RecordDecision(caseFile, pass, conclusion, rationale, resolver, s.clock.Now()); err != nil {
		return domain.FinalDecision{}, err
	}
	if err = s.repository.SaveBatch(ctx, panel.batch, expected); err != nil {
		return domain.FinalDecision{}, err
	}
	decision, _ := decisionFor(panel.batch.Decisions, materialID)
	return decision, nil
}
