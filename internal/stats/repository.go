package stats

import (
	"link-generator/pkg/db"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatsRepository struct {
	db *db.Db
}

type StatsQuery struct {
	to     time.Time
	from   time.Time
	linkID *uint
}

func NewStatsRepository(db *db.Db) *StatsRepository {
	return &StatsRepository{
		db,
	}
}

func (r *StatsRepository) GetStats(query *StatsQuery) ([]Stats, error) {
	var stats []Stats

	q := r.db.Preload("Link").Order("date asc")

	if query.from.Unix() > 0 {
		q = q.Where("date >= ?", query.from)
	}

	if query.to.Unix() > 0 {
		q = q.Where("date <= ?", query.to)
	}

	if query.linkID != nil {
		q = q.Where("link_id = ?", *query.linkID)
	}

	res := q.Find(&stats)

	if res.Error != nil {
		return nil, res.Error
	}

	return stats, nil
}

func (r *StatsRepository) GetStatsGroupByDate(query *StatsQuery) ([]StatsGroupByDate, error) {
	var stats []StatsGroupByDate

	q := r.db.Model(&Stats{}).Select("date, sum(clicks) as clicks").Order("date desc")

	if query.from.Unix() > 0 && query.to.Unix() > 0 {
		q = q.Where("date BETWEEN ? AND ?", query.from, query.to)
	}

	res := q.Group("date").Find(&stats)

	if res.Error != nil {
		return nil, res.Error
	}

	return stats, nil
}

func (r *StatsRepository) GetStatById(id int) (*Stats, error) {
	var result Stats

	res := r.db.Preload("Link").Where("id=?", id).First(&result)

	if res.Error != nil {
		return nil, res.Error
	}

	return &result, nil
}

func (r *StatsRepository) UpdateLinkClicks(linkID int) (*Stats, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var stats Stats
	err := r.db.Transaction(func(tx *gorm.DB) error {
		stats = Stats{
			LinkID: uint(linkID),
			Date:   today,
			Clicks: 1,
		}

		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "link_id"}, {Name: "date"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"clicks": gorm.Expr("stats.clicks + 1"),
			}),
		}).Create(&stats)

		if result.Error != nil {
			return result.Error
		}

		if err := tx.Where("link_id = ? AND date = ?", linkID, today).First(&stats).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}
