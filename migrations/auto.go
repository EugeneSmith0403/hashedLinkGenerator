package main

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	stripeGo "github.com/stripe/stripe-go/v84"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}

	dsn := os.Getenv("DSN")
	dbURL := os.Getenv("DATABASE_URL")

	d, err := iofs.New(sqlFiles, "sql")
	if err != nil {
		panic(err.Error())
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		panic(err.Error())
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err.Error())
	}

	fmt.Println("migrations applied")

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	stripeClient := stripeGo.NewClient(os.Getenv("STRIPE_TOKEN"))
	seedPlans(db, stripeClient)
	fmt.Println("seeds applied")
}
