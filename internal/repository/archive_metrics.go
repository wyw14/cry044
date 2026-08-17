package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
	"sort"
	"time"
)

func (a *ReviewArchive) Statistics(ctx context.Context, from, to time.Time) (domain.ReviewStatistics, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := domain.ReviewStatistics{From: from, To: to}
	completed := 0
	var cycleHours float64
	reasonCounts := map[string]int{}
	criterionCounts := map[string]*domain.CriterionEffectiveness{}
	for _, batch := range a.state.batches {
		if batch.CompletedAt != nil && batch.StartedAt != nil && !batch.CompletedAt.Before(from) && batch.CompletedAt.Before(to) {
			completed++
			cycleHours += batch.CompletedAt.Sub(*batch.StartedAt).Hours()
		}
		if batch.ReturnReason != "" {
			reasonCounts[batch.ReturnReason]++
		}
		reviews := a.state.panels[batch.ID]
		for _, review := range reviews {
			if review.SubmittedAt == nil {
				continue
			}
			for _, score := range review.Scores {
				stat := criterionCounts[score.CriterionID]
				if stat == nil {
					stat = &domain.CriterionEffectiveness{CriterionID: score.CriterionID}
					criterionCounts[score.CriterionID] = stat
				}
				stat.UseCount++
				if score.VetoTriggered {
					stat.VetoCount++
				}
			}
		}
		if batch.Snapshot != nil {
			for _, material := range batch.Materials {
				if aggregate, err := domain.Aggregate(*batch.Snapshot, material.ID, reviews); err == nil {
					for _, id := range aggregate.Disagreements {
						stat := criterionCounts[id]
						if stat == nil {
							stat = &domain.CriterionEffectiveness{CriterionID: id}
							criterionCounts[id] = stat
						}
						stat.DisagreementCount++
					}
				}
			}
		}
	}
	if completed > 0 {
		result.AverageCycleHours = cycleHours / float64(completed)
	}
	for reason, count := range reasonCounts {
		result.ReturnReasons = append(result.ReturnReasons, domain.ReturnReasonCount{Reason: reason, Count: count})
	}
	for _, stat := range criterionCounts {
		result.Criteria = append(result.Criteria, *stat)
	}
	sort.Slice(result.ReturnReasons, func(i, j int) bool {
		if result.ReturnReasons[i].Count != result.ReturnReasons[j].Count {
			return result.ReturnReasons[i].Count > result.ReturnReasons[j].Count
		}
		return result.ReturnReasons[i].Reason < result.ReturnReasons[j].Reason
	})
	sort.Slice(result.Criteria, func(i, j int) bool { return result.Criteria[i].CriterionID < result.Criteria[j].CriterionID })
	return result, contextErr(ctx)
}
