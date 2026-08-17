package repository

import (
	"context"
	"github.com/wyw14/cry044/internal/domain"
	"time"
)

func (p *PGLedger) Statistics(ctx context.Context, from, to time.Time) (domain.ReviewStatistics, error) {
	var average float64
	if err := p.pool.QueryRow(ctx, `select coalesce(avg(extract(epoch from
((payload->>'CompletedAt')::timestamptz-(payload->>'StartedAt')::timestamptz))/3600),0)
from batches where status='completed' and (payload->>'CompletedAt')::timestamptz >= $1
and (payload->>'CompletedAt')::timestamptz < $2`, from, to).Scan(&average); err != nil {
		return domain.ReviewStatistics{}, err
	}
	result := domain.ReviewStatistics{AverageCycleHours: average, From: from, To: to}
	reasons, err := p.pool.Query(ctx, `select payload->>'ReturnReason', count(*) from batches
where coalesce(payload->>'ReturnReason','')<>'' and (payload->>'UpdatedAt')::timestamptz >= $1
and (payload->>'UpdatedAt')::timestamptz < $2 group by payload->>'ReturnReason' order by count(*) desc`, from, to)
	if err != nil {
		return result, err
	}
	for reasons.Next() {
		var reason domain.ReturnReasonCount
		if err = reasons.Scan(&reason.Reason, &reason.Count); err != nil {
			reasons.Close()
			return result, err
		}
		result.ReturnReasons = append(result.ReturnReasons, reason)
	}
	if err = reasons.Err(); err != nil {
		reasons.Close()
		return result, err
	}
	reasons.Close()
	rows, err := p.pool.Query(ctx, `select score->>'CriterionID', count(*),
	count(*) filter(where coalesce((score->>'VetoTriggered')::boolean,false))
	from reviews join batches on batches.id=reviews.batch_id
	cross join lateral jsonb_array_elements(reviews.payload->'Scores') score
	where (batches.payload->>'UpdatedAt')::timestamptz >= $1
	and (batches.payload->>'UpdatedAt')::timestamptz < $2
	group by score->>'CriterionID' order by score->>'CriterionID'`, from, to)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.CriterionEffectiveness
		if err = rows.Scan(&item.CriterionID, &item.UseCount, &item.VetoCount); err != nil {
			return result, err
		}
		result.Criteria = append(result.Criteria, item)
	}
	return result, rows.Err()
}
