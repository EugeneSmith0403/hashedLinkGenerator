package clickhouse

import (
	"context"
	"fmt"
	"link-generator/configs"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Clickhouse struct {
	client driver.Conn
	ctx    context.Context
}

func NewCliсkhouse(cfg *configs.ClickHouseConfig) (*Clickhouse, error) {
	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.DB,
			Username: cfg.User,
			Password: cfg.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	return &Clickhouse{client: conn, ctx: ctx}, nil
}

func (c Clickhouse) Exec(sql string) error {
	err := c.client.Exec(c.ctx, sql)

	if err != nil {
		return err
	}

	return nil
}

func (c Clickhouse) Close() error {
	return c.client.Close()
}
