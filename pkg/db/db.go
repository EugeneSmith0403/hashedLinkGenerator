package db

import (
	"adv/go-http/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	*gorm.DB
}

func NewDb(config *configs.Config) *Db {
	db, err := gorm.Open(postgres.Open(config.Db.Dsn), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	return &Db{
		db,
	}
}

func (db Db) Close() {
	sqlDB, err := db.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}
