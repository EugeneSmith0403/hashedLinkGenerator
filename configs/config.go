package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Db     DbConfig
	Auth   AuthConfig
	Stripe StripeConfig
}

type StripeConfig struct {
	ApiKey        string
	WebhookSecret string
	ReturnURL     string
}

type DbConfig struct {
	Dsn string
}

type AuthConfig struct {
	Secret string
}

func LoadConfig(envFiles ...string) *Config {
	// If no env files specified, try to load default .env
	if len(envFiles) == 0 {
		envFiles = []string{".env"}
	}

	err := godotenv.Load(envFiles...)

	if err != nil {
		log.Println("Error loading .env file, using default config")
	}

	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("TOKEN"),
		},
		Stripe: StripeConfig{
			ApiKey:        os.Getenv("STRIPE_TOKEN"),
			WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
			ReturnURL:     os.Getenv("STRIPE_RETURN_URL"),
		},
	}
}
