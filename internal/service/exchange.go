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
	fmt.Fprintf(&sheet, "评议批次：%s\n", batch.Name)
	fmt.Fprintf(&sheet, "模板：%s V%d\n", batch.Snapshot.TemplateID, batch.Snapshot.Version)
	fmt.Fprintf(&sheet, "状态：%s\n\n", batch.Status)
	for _, result := range results {
		fmt.Fprintf(&sheet, "材料：%s\n", result.MaterialID)
		fmt.Fprintf(&sheet, "聚合得分：%.2f\n", result.WeightedScore)
		fmt.Fprintf(&sheet, "通过：%t\n", result.Passed)
		fmt.Fprintf(&sheet, "否决：%t\n", result.Vetoed)
		fmt.Fprintf(&sheet, "评审人数：%d\n", result.ReviewerCount)
		if len(result.Disagreements) == 0 {
			sheet.WriteString("分歧条目：无\n\n")
		} else {
			fmt.Fprintf(&sheet, "分歧条目：%s\n\n", strings.Join(result.Disagreements, ", "))
		}
	}
	stored, err := e.files.Save(ctx, fmt.Sprintf("review-sheet-%s.txt", batch.ID), []byte(sheet.String()))
	return stored.Path, err
}
