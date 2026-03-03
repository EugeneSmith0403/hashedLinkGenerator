package stats

import (
	"time"

	"adv/go-http/internal/models"

	"gorm.io/gorm"
)

type Stats struct {
	gorm.Model
	LinkID uint        `json:"linkId" gorm:"uniqueIndex:idx_stats_link_date"`
	Link   models.Link `gorm:"foreignKey:LinkID"`
	Clicks int32       `json:"clicks" default:"0"`
	Date   time.Time   `json:"date" gorm:"uniqueIndex:idx_stats_link_date;type:date"`
}

type StatsGroupByDate struct {
	Clicks int32     `json:"clicks" default:"0"`
	Date   time.Time `json:"date" gorm:"index:idx_stats_date,unique;type:date"`
}
