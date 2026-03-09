package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode     string
	Db       DbConfig
	Auth     AuthConfig
	Stripe   StripeConfig
	Redis    Redis
	RabbitMq RabbitMq
}

type Redis struct {
	Addr     string
	Username string
	Password string
	Cache    string
}
type RabbitMq struct {
	Amqp      string
	User      string
	Password  string
	Consumers string
}

type StripeConfig struct {
	Mode          string
	ApiKey        string
	WebhookSecret string
	ReturnURL     string
}

type DbConfig struct {
	Dsn string
}

type AuthConfig struct {
	Secret    string
	ExpiredAt string
}

func LoadConfig(envFiles ...string) *Config {
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
			Secret:    os.Getenv("TOKEN"),
			ExpiredAt: os.Getenv("EXPIRED_AT"),
		},
		Stripe: StripeConfig{
			Mode:          os.Getenv("MODE"),
			ApiKey:        os.Getenv("STRIPE_TOKEN"),
			WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
			ReturnURL:     os.Getenv("STRIPE_RETURN_URL"),
		},
		Redis: Redis{
			Addr:     os.Getenv("REDIS_ADDR"),
			Username: os.Getenv("REDIS_USER"),
			Password: os.Getenv("REDIS_USER_PASSWORD"),
			Cache:    os.Getenv("REDIS_CACHE"),
		},
		RabbitMq: RabbitMq{
			User:      os.Getenv("RABBITMQ_USER"),
			Password:  os.Getenv("RABBITMQ_PASSWORD"),
			Consumers: os.Getenv("RABBITMQ_CONSUMERS"),
			Amqp:      os.Getenv("RABBITNQ_AMQP"),
		},
	}
}
