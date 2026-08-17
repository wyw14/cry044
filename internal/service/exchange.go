package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/platform/localfiles"
)

type LocalExchange struct{ files *localfiles.Store }

func NewLocalExchange(root string) *LocalExchange {
	return &LocalExchange{files: localfiles.New(root, 2<<20, ".json", ".txt")}
}

func (e *LocalExchange) Read(ctx context.Context, payload []byte) (domain.StandardTemplate, error) {
	if err := ctx.Err(); err != nil {
		return domain.StandardTemplate{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var template domain.StandardTemplate
	if err := decoder.Decode(&template); err != nil {
		return template, err
	}
	if strings.TrimSpace(template.ID) == "" || template.Version < 1 || len(template.Criteria) == 0 {
		return template, fmt.Errorf("template id, version and criteria are required")
	}
	return template, nil
}

func (e *LocalExchange) WriteTemplate(ctx context.Context, template domain.StandardTemplate) (string, error) {
	payload, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "", err
	}
	stored, err := e.files.Save(ctx, fmt.Sprintf("template-%s-v%d.json", template.ID, template.Version), payload)
	return stored.Path, err
}

func (e *LocalExchange) WriteReviewSheet(ctx context.Context, batch domain.Batch, results []domain.MaterialResult) (string, error) {
	var sheet strings.Builder
	fmt.Fprintf(&sheet, "评议批次：%s\n模板：%s V%d\n状态：%s\n\n", batch.Name, batch.Snapshot.TemplateID, batch.Snapshot.Version, batch.Status)
	for _, result := range results {
		fmt.Fprintf(&sheet, "材料 %s｜得分 %.2f｜通过 %t｜否决 %t｜评审 %d\n", result.MaterialID, result.WeightedScore, result.Passed, result.Vetoed, result.ReviewerCount)
	}
	stored, err := e.files.Save(ctx, fmt.Sprintf("review-sheet-%s.txt", batch.ID), []byte(sheet.String()))
	return stored.Path, err
}
