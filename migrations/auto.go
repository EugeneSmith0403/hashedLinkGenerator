package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	stripeGo "github.com/stripe/stripe-go/v84"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed postgres/*.sql
var pgFiles embed.FS

//go:embed clickhouse/*.sql
var chFiles embed.FS

func main() {
	target := flag.String("target", "all", "which migrations to run: all, postgres, clickhouse")
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}

	switch *target {
	case "postgres":
		runPostgresMigrations()
		return
	case "clickhouse":
		runClickHouseMigrations()
		return
	default:
		runPostgresMigrations()
		runClickHouseMigrations()
	}

	dsn := os.Getenv("DSN")
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	stripeClient := stripeGo.NewClient(os.Getenv("STRIPE_TOKEN"))
	seedPlans(db, stripeClient)
	fmt.Println("seeds applied")
}

func runPostgresMigrations() {
	dbURL := os.Getenv("DATABASE_URL")

	d, err := iofs.New(pgFiles, "postgres")
	if err != nil {
		panic(fmt.Errorf("postgres iofs: %w", err))
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		panic(fmt.Errorf("postgres migrate init: %w", err))
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Errorf("postgres migrate up: %w", err))
	}

	fmt.Println("postgres migrations applied")
}

func runClickHouseMigrations() {
	addr := os.Getenv("CLICKHOUSE_ADDR")
	db := os.Getenv("CLICKHOUSE_DB")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	chURL := fmt.Sprintf("clickhouse://%s/%s?username=%s&password=%s&x-multi-statement=true", addr, db, user, password)

	d, err := iofs.New(chFiles, "clickhouse")
	if err != nil {
		panic(fmt.Errorf("clickhouse iofs: %w", err))
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, chURL)
	if err != nil {
		panic(fmt.Errorf("clickhouse migrate init: %w", err))
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Errorf("clickhouse migrate up: %w", err))
	}

	fmt.Println("clickhouse migrations applied")
}
