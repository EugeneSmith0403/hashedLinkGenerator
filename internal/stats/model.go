package stats

import (
	"time"

	"link-generator/internal/models"

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

type GetStatByLink struct {
	Date          time.Time `json:"date"          gorm:"column:date"`
	AmountClicks  uint64    `json:"amountClicks"  gorm:"column:amount_clicks"`
}

type ClientContext struct {
	// Network
	IP           string
	ForwardedFor string
	RealIP       string
	RemoteAddr   string
	RemotePort   string
	Country      string

	// Headers
	UserAgent      string
	Accept         string
	AcceptLanguage string
	AcceptEncoding string
	Origin         string
	Referer        string

	// Device
	DeviceType string
	OS         string
	Browser    string

	// Security
	Fingerprint    string
	RequestID      string
	ForwardedProto string
	ForwardedHost  string
	ForwardedPort  string

	Timestamp time.Time
	Scheme    string
}
