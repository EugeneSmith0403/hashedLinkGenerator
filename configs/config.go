package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode       string
	GeoIPPath  string
	Db         DbConfig
	Auth       AuthConfig
	TOTP       TOTPConfig
	RateLimit  RateLimitConfig
	Stripe     StripeConfig
	Redis      Redis
	RabbitMq   RabbitMq
	Mailer     MailerConfig
	ClickHouse ClickHouseConfig
}

type MailerConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
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

type ClickHouseConfig struct {
	Addr     string
	DB       string
	User     string
	Password string
}

type AuthConfig struct {
	Secret    string
	ExpiredAt string
}

type TOTPConfig struct {
	Issuer string
	Period uint
	Skew   uint
}

type RateLimiterConfig struct {
	Capacity   int
	RefillRate float64
}

type RateLimitConfig struct {
	Account RateLimiterConfig
	IP      RateLimiterConfig
}

func smtpPort(s string) int {
	port, err := strconv.Atoi(s)
	if err != nil {
		return 587
	}
	return port
}

func envUint(key string, defaultVal uint) uint {
	v, err := strconv.ParseUint(os.Getenv(key), 10, 64)
	if err != nil {
		return defaultVal
	}
	return uint(v)
}

func envString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultVal
	}
	return v
}

func envFloat(key string, defaultVal float64) float64 {
	v, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil {
		return defaultVal
	}
	return v
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
		GeoIPPath: os.Getenv("GEOIP_PATH"),
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret:    os.Getenv("TOKEN"),
			ExpiredAt: os.Getenv("EXPIRED_AT"),
		},
		TOTP: TOTPConfig{
			Issuer: envString("TOTP_ISSUER", "LinkShort"),
			Period: envUint("TOTP_PERIOD", 30),
			Skew:   envUint("TOTP_SKEW", 5),
		},
		RateLimit: RateLimitConfig{
			Account: RateLimiterConfig{
				Capacity:   envInt("RATE_LIMIT_ACCOUNT_CAPACITY", 100),
				RefillRate: envFloat("RATE_LIMIT_ACCOUNT_REFILL_RATE", 100),
			},
			IP: RateLimiterConfig{
				Capacity:   envInt("RATE_LIMIT_IP_CAPACITY", 100),
				RefillRate: envFloat("RATE_LIMIT_IP_REFILL_RATE", 100),
			},
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
		Mailer: MailerConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     smtpPort(os.Getenv("SMTP_PORT")),
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
		ClickHouse: ClickHouseConfig{
			Addr:     os.Getenv("CLICKHOUSE_ADDR"),
			DB:       os.Getenv("CLICKHOUSE_DB"),
			User:     os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
	}
}
