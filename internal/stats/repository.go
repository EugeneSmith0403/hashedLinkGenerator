package stats

import (
	"context"
	"link-generator/internal/models"
	"link-generator/pkg/clickhouse"
	"time"

	"gorm.io/gorm"
)

type StatsRepository struct {
	ch *clickhouse.Clickhouse
}

type StatsQuery struct {
	to     time.Time
	from   time.Time
	linkID *uint
}

const TableName = "link_clicks"

func NewStatsRepository(ch *clickhouse.Clickhouse) *StatsRepository {
	return &StatsRepository{
		ch: ch,
	}
}

func (r *StatsRepository) getTableWithContext() *gorm.DB {
	ctx := context.Background()
	return r.ch.DB.WithContext(ctx).Table(TableName)
}

func (r *StatsRepository) GetStats(query *StatsQuery) ([]models.LinkTransition, error) {
	var stats []models.LinkTransition

	q := r.getTableWithContext()

	if query.from.Unix() > 0 {
		q = q.Where("date >= ?", query.from)
	}

	if query.to.Unix() > 0 {
		q = q.Where("date <= ?", query.to)
	}

	if query.linkID != nil {
		q = q.Where("link_id = ?", *query.linkID)
	}

	res := q.
		Order("date asc").
		Find(&stats)

	if res.Error != nil {
		return nil, res.Error
	}

	return stats, nil
}

func (r *StatsRepository) Insert(events []models.LinkTransition) error {
	return r.getTableWithContext().CreateInBatches(events, len(events)).Error
}

func (r *StatsRepository) GetStatByLink(query *StatsQuery) ([]GetStatByLink, error) {
	var result []GetStatByLink

	q := r.getTableWithContext().
		Select("toDate(clicked_at) as date, count() as amount_clicks").
		Where("link_id=?", query.linkID)

	if query.from.Unix() > 0 && query.to.Unix() > 0 {
		q = q.Where("date BETWEEN ? AND ?", query.from, query.to)
	}

	q = q.
		Group("toDate(clicked_at)").
		Order("toDate(clicked_at) desc").
		Find(&result)

	if q.Error != nil {
		return nil, q.Error
	}

	return result, nil
}
