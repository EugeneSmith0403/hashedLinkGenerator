# LinkShort

A full-stack URL shortener service with subscription billing, built with Go and Nuxt 3.

## Features

- **URL Shortening**: Create and manage shortened links with click tracking
- **Session-based Auth**: JWT tokens tied to `auth_sessions` (IP, user agent, expiry, active flag, `is_active` filter on lookup)
- **Two-Factor Authentication**: TOTP-based 2FA with QR code setup
- **Subscription Billing**: Stripe integration — plans, payment methods, subscriptions, refunds
- **Payment History**: Full audit trail of all payments per user
- **Click Statistics**: Time-range and per-link analytics stored in ClickHouse
- **Async Event Processing**: RabbitMQ consumers for payment intents, subscriptions, invoices, and click stats
- **Transactional Email**: SMTP mailer — welcome email on registration, payment confirmation on invoice paid
- **Redis Caching**: Grouped stats cached per `(linkHash, date-range)` with exact key invalidation via Redis Set tracking
- **Rate Limiting**: Token-bucket limiter via Redis Lua script — IP-based (auth, redirect) and per-account (authenticated endpoints), `429 Too Many Requests` with configurable capacity and refill rate
- **GeoIP Country Detection**: MaxMind GeoIP2 lookup for click country metadata
- **Internationalization**: EN / RU / DE (frontend + email templates)
- **Observability**: Prometheus metrics, structured JSON logs, Grafana dashboards, Loki log aggregation

## Tech Stack

### Backend
- **Go** (net/http ServeMux)
- **PostgreSQL** with GORM
- **ClickHouse** (click analytics — `link_clicks` table, GORM driver)
- **Redis** (stats caching + token-bucket rate limiting)
- **RabbitMQ** (async event processing via `rabbitmq/amqp091-go`)
- **Stripe Go SDK** (v84)
- **golang-jwt/jwt**
- **pquerna/otp** (TOTP 2FA)
- **golang-migrate** (SQL migrations for both PostgreSQL and ClickHouse)
- **wneessen/go-mail** (SMTP client)
- **nicksnyder/go-i18n** (email template i18n)
- **oschwald/geoip2-golang** (MaxMind GeoIP2 country lookup)
- **prometheus/client_golang** (metrics)

### Frontend
- **Nuxt 3** (Vue 3, TypeScript)
- **Tailwind CSS**
- **Pinia** (auth store)
- **TanStack Vue Query** (data fetching, cache management)
- **vee-validate + zod** (form validation)
- **@stripe/stripe-js** (Stripe Elements)
- **@nuxtjs/i18n** (EN / RU / DE)

### Infrastructure
- **Docker Compose** (local development)
- **Kubernetes** (production) — see [k8s/README.md](k8s/README.md)
- **GitHub Actions** (CI/CD — test → build → deploy)
- **Prometheus + Grafana** (metrics & dashboards)
- **Loki + Promtail** (log aggregation)

## Project Structure

```
.
├── cmd/
│   ├── server/                    # HTTP server entry point
│   ├── consumers/
│   │   ├── paymentIntent/         # PaymentIntent event consumer
│   │   ├── subscription/          # Subscription event consumer
│   │   ├── invoice/               # Invoice event consumer
│   │   └── stats/                 # Click stats consumer
│   └── shared/                    # Shared consumer loop & config loader
├── configs/                       # Configuration loading
├── internal/
│   ├── account/                   # Stripe customer account management + 2FA service
│   ├── auth/                      # Registration & login handlers
│   ├── auth_session/              # Session create/update
│   ├── consts/                    # RabbitMQ exchange/queue/routing constants
│   ├── consumers/                 # Consumer handler logic
│   ├── jwt/                       # JWT service
│   ├── locales/                   # Embedded email templates & i18n strings (EN/RU/DE)
│   ├── mailer/                    # SMTP mailer service
│   ├── link/                      # Link CRUD + redirect + stats publishing
│   ├── models/                    # Shared message/event models
│   ├── publishers/                # RabbitMQ publishers
│   ├── stats/                     # Click statistics (ClickHouse repo, handler, service, cache)
│   ├── user/                      # User repository + 2FA handlers
│   └── payments/
│       ├── invoice/               # Invoice repository & service
│       ├── payment/               # Payment repository & handler
│       ├── plan/                  # Plans
│       ├── stripe/                # Stripe handlers & services
│       ├── subscription/          # Subscription service, handlers
│       └── webhook/               # Stripe webhook handler & service
├── pkg/
│   ├── clickhouse/                # ClickHouse connection wrapper
│   ├── db/                        # PostgreSQL connection + Prometheus metrics
│   ├── event/                     # In-process event bus
│   ├── limiter/                   # Token-bucket rate limiter
│   ├── middleware/                # CORS, auth, logging (JSON), metrics, rate limit
│   ├── rabbitMq/                  # RabbitMQ client wrapper
│   ├── redis/                     # Redis client wrapper
│   ├── request/                   # Body parsing helpers
│   └── response/                  # JSON response helpers
├── migrations/
│   ├── postgres/                  # PostgreSQL migration files
│   ├── clickhouse/                # ClickHouse migration files
│   └── auto.go                    # Migration runner
├── docker/
│   ├── Dockerfile.server          # Server image (multi-stage)
│   └── Dockerfile.consumer        # Consumers image (ARG CONSUMER)
├── monitoring/
│   ├── prometheus.yml             # Prometheus scrape config (dev)
│   ├── loki/loki.yml              # Loki config
│   ├── promtail/promtail.yml      # Promtail — Docker log collection
│   └── grafana/
│       ├── datasources/           # Prometheus + Loki auto-provisioning
│       ├── dashboards/            # Dashboard provider config
│       └── dashboard-files/       # Pre-built dashboards (server, postgres, redis, rabbitmq, clickhouse)
├── k8s/                           # Production Kubernetes manifests
│   └── README.md                  # K8s deploy guide
├── .github/
│   └── workflows/
│       └── deploy.yml             # CI/CD: test → build → deploy
├── frontend/
│   ├── apps/web/                  # Nuxt 3 SPA
│   └── packages/
│       ├── ui/                    # Shared component library + Storybook
│       └── eslint-config/         # Shared ESLint config
├── Makefile
├── docker-compose.yml
└── .env.example
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Node.js 20+ and pnpm
- Stripe account (test keys)
- MaxMind GeoLite2 Country `.mmdb` file (optional)

### Environment Variables

Copy `.env.example` to `.env`:

```env
DSN="host=localhost user=postgres password=pass dbname=linkshort port=5432 sslmode=disable"
TOKEN="your_jwt_secret"

REDIS_ADDR="localhost:6379"
REDIS_USER=""
REDIS_USER_PASSWORD=""
REDIS_PASSWORD=""
REDIS_CACHE=5

RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_CONSUMERS=1
RABBITNQ_AMQP="amqp://guest:guest@localhost:5672/"

CLICKHOUSE_ADDR=localhost:9000
CLICKHOUSE_DB=default
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

STRIPE_TOKEN="sk_test_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
STRIPE_RETURN_URL="http://localhost:3000"

# GeoIP — optional, country detection is disabled if unset
GEOIP_PATH=/path/to/GeoLite2-Country.mmdb

# Mailer — optional, emails are skipped if SMTP_HOST is empty
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=no-reply@example.com
```

### Running with Docker Compose

```bash
# Start all infrastructure + monitoring stack
docker-compose up -d

# Run migrations
make migrate

# Start server
make server

# Start all consumers
make consumers

# Start frontend
make frontend
```

After startup:

| Service    | URL                   |
|------------|-----------------------|
| API server | http://localhost:8081  |
| Frontend   | http://localhost:3000  |
| Grafana    | http://localhost:3001 (admin / admin) |
| Prometheus | http://localhost:9090  |

### Running with Make

```bash
make dev                  # infrastructure + server + consumers + frontend

make docker-up            # infrastructure only
make migrate              # run all migrations (postgres + clickhouse)
make migrate-postgres     # postgres only
make migrate-clickhouse   # clickhouse only
make server               # HTTP server (port 8081)
make consumers            # all consumer workers
make consumer-payment
make consumer-subscription
make consumer-invoice
make consumer-stats
make frontend             # Nuxt dev server
make stripe-webhook       # Stripe CLI webhook listener

# Scale consumers
make consumers WORKERS=3
```

### Running Locally (without Make)

```bash
go run ./cmd/server

# Each in a separate terminal
go run ./cmd/consumers/paymentIntent
go run ./cmd/consumers/subscription
go run ./cmd/consumers/invoice
go run ./cmd/consumers/stats

cd frontend && pnpm install && pnpm dev
```

### Build Binaries

```bash
make build
# Produces: bin/server, bin/consumer-payment, bin/consumer-subscription, bin/consumer-invoice, bin/consumer-stats
```

## Observability

### Metrics (Prometheus + Grafana)

Prometheus scrapes metrics from the server and all infrastructure exporters. Pre-built Grafana dashboards are provisioned automatically on startup:

| Dashboard      | Panels |
|----------------|--------|
| Server Overview | Request rate, error rate, latency p50/p95/p99, PostgreSQL query duration, connection pool, logs |
| PostgreSQL     | Connections, cache hit ratio, row ops, transactions, deadlocks, locks |
| Redis          | Memory, hit ratio, commands/s, evictions |
| RabbitMQ       | Queue depth, ready/unacked messages, publish/deliver rate |
| ClickHouse     | Queries/s, inserts, memory, merges |

HTTP metrics exposed at `GET /metrics`:
- `http_requests_total{method, path, status}` — request counter
- `http_request_duration_seconds{method, path, status}` — latency histogram (p50/p95/p99)
- `pg_query_duration_seconds{operation}` — PostgreSQL query duration per operation type
- `go_sql_*{db_name="link_generator"}` — connection pool stats

### Logs (Loki + Promtail)

The server outputs structured JSON logs via `log/slog`:

```json
{"time":"2026-01-01T12:00:00Z","level":"INFO","msg":"request","method":"GET","path":"/links","status":200,"duration_ms":12,"ip":"127.0.0.1:54321"}
```

Promtail collects logs from all Docker containers and ships them to Loki. View logs in Grafana → Explore → Loki datasource.

## Production Deployment

See [k8s/README.md](k8s/README.md) for the full Kubernetes setup and deployment guide.

### CI/CD

Every push to `master` triggers the GitHub Actions pipeline:

1. **Test** — `go test ./...`
2. **Build** — Docker images pushed to `ghcr.io` (tagged with `latest` + commit SHA)
3. **Deploy** — `kubectl apply` + rollout status check

Required GitHub secret: `KUBE_CONFIG` (base64-encoded kubeconfig).

## Async Event Architecture

```
Stripe → POST /stripe/webhook
           │
           ├── PaymentIntent events → paymentIntent exchange → paymentIntentQueue
           │                                                        │
           │                                              PaymentIntent consumer
           │                                              (sync payment record)
           │
           ├── Subscription events → Subscription exchange → subscriptionQueue
           │                                                       │
           │                                             Subscription consumer
           │                                             (create/update/cancel sub,
           │                                              create initial invoice)
           │
           └── invoice.payment_succeeded → Invoice exchange → invoiceQueue
                                                                    │
                                                          Invoice consumer
                                                          (upsert invoice + payment,
                                                           send confirmation email)

GET /{hash} → redirect
           │
           └── link.visited event → Stats exchange → statsQueue
                                                          │
                                                    Stats consumer
                                                    (insert into ClickHouse link_clicks,
                                                     invalidate Redis stats cache)
```

## API Endpoints

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register → creates user + Stripe account + auth session |
| POST | `/auth/login` | No | Login → `{email, token, is2faEnabled}` |

### User & 2FA

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/users/me` | Yes | Current user info |
| POST | `/users/me/2fa/setup` | Yes | Generate TOTP secret → `{qrCode}` |
| POST | `/users/2fa/verify` | No | Verify TOTP code → `{token}` |

### Links

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/links` | Yes | List user's links |
| POST | `/link` | Yes + Active Sub | Create shortened link |
| PATCH | `/link/{id}` | Yes | Update link |
| DELETE | `/link/{id}` | Yes | Delete link |
| GET | `/{hash}` | No | Redirect + publish click event |

### Plans & Subscriptions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/plans` | Yes | List available plans |
| GET | `/subscriptions/me` | Yes | Current subscription |
| POST | `/subscriptions/method` | Yes | Create SetupIntent → `{clientSecret}` |
| POST | `/subscriptions` | Yes | Create subscription |
| PATCH | `/subscriptions/cancel` | Yes | Cancel subscription |

### Payments

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/payments` | Yes | Payment history |
| POST | `/stripe/paymentIntent` | Yes | Create PaymentIntent |
| POST | `/stripe/paymentIntent/confirm` | Yes | Confirm PaymentIntent |
| POST | `/stripe/paymentIntent/cancel` | Yes | Refund |
| POST | `/stripe/webhook` | No | Stripe webhook handler |

### Statistics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/stats` | Yes | Raw click events `?from=&to=&linkId=` |
| GET | `/stats/link/{id}` | Yes | Clicks grouped by date `?from=&to=` |

### Monitoring

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/metrics` | No | Prometheus metrics |

## License

MIT
