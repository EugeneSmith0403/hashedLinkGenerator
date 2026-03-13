package clickhouse

import (
	"fmt"
	"link-generator/configs"

	gormClickhouse "gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type Clickhouse struct {
	DB *gorm.DB
}

func NewCliсkhouse(cfg *configs.ClickHouseConfig) (*Clickhouse, error) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s", cfg.User, cfg.Password, cfg.Addr, cfg.DB)

	db, err := gorm.Open(gormClickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	return &Clickhouse{DB: db}, nil
}

func (db Clickhouse) Close() {
	ch, err := db.DB.DB()
	if err == nil {
		ch.Close()
	}
}
