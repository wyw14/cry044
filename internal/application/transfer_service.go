package application

import (
	"context"
	"fmt"
	"github.com/wyw14/cry044/internal/domain"
	"time"
)

type TransferService struct {
	repository ReviewRepository
	clock      ReviewClock
	identities IdentityProvider
	reader     Importer
	writer     Exporter
}

func NewTransferService(repository ReviewRepository, clock ReviewClock, identities IdentityProvider, reader Importer, writer Exporter) *TransferService {
	return &TransferService{repository: repository, clock: clock, identities: identities, reader: reader, writer: writer}
}

func (s *TransferService) ImportTemplate(ctx context.Context, payload []byte, idempotencyKey string) (domain.StandardTemplate, error) {
	template, err := s.reader.Read(ctx, payload)
	if err != nil {
		return template, err
	}
	when := s.clock.Now()
	template.Status, template.Revision = domain.TemplateDraft, 1
	template.CreatedAt, template.UpdatedAt = when, when
	return s.repository.SaveTemplate(ctx, template, 0, "import")
}

func (s *TransferService) ExportTemplate(ctx context.Context, templateID string, version int) (string, error) {
	template, err := s.repository.Template(ctx, templateID, version)
	if err != nil {
		return "", err
	}
	return s.writer.WriteTemplate(ctx, template)
}

func (s *TransferService) PrintReviewSheet(ctx context.Context, batchID string) (string, error) {
	batch, err := s.repository.Batch(ctx, batchID)
	if err != nil {
		return "", err
	}
	if batch.Snapshot == nil || len(batch.Results) != len(batch.Materials) {
		return "", fmt.Errorf("batch must be completed before printing")
	}
	reviews, err := s.repository.Reviews(ctx, batchID)
	if err != nil {
		return "", err
	}
	report := BuildAssessmentReport(batch, reviews, s.clock.Now())
	results := make([]domain.MaterialResult, len(report.Lines))
	for index, line := range report.Lines {
		results[index] = domain.MaterialResult{MaterialID: line.MaterialID, WeightedScore: line.Score, Passed: line.Passed, Vetoed: line.Vetoed}
	}
	return s.writer.WriteReviewSheet(ctx, batch, results)
}

func (s *TransferService) Statistics(ctx context.Context, from, to time.Time) (domain.ReviewStatistics, error) {
	if duration := to.Sub(from); duration <= 0 || duration > 366*24*time.Hour {
		return domain.ReviewStatistics{}, fmt.Errorf("statistics range must be within 366 days")
	}
	return s.repository.Statistics(ctx, from, to)
}
