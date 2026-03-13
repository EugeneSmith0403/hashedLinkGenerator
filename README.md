# LinkShort

A full-stack URL shortener service with subscription billing, built with Go and Nuxt 3.

## Features

- **URL Shortening**: Create and manage shortened links with click tracking
- **Session-based Auth**: JWT tokens tied to `auth_sessions` (IP, user agent, expiry, active flag)
- **Two-Factor Authentication**: TOTP-based 2FA with QR code setup
- **Subscription Billing**: Stripe integration — plans, payment methods, subscriptions, refunds
- **Payment History**: Full audit trail of all payments per user
- **Click Statistics**: Time-range and per-link analytics stored in ClickHouse
- **Async Event Processing**: RabbitMQ consumers for payment intents, subscriptions, invoices, and click stats
- **Transactional Email**: SMTP mailer — welcome email on registration, payment confirmation on invoice paid
- **Redis Caching**: Stats cache with automatic invalidation on new click events
- **GeoIP Country Detection**: MaxMind GeoIP2 lookup for click country metadata
- **Internationalization**: EN / RU / DE (frontend + email templates)

## Tech Stack

### Backend
- **Go** (net/http ServeMux)
- **PostgreSQL** with GORM
- **ClickHouse** (click analytics — `link_clicks` table, GORM driver)
- **Redis** (stats caching with pattern-based invalidation)
- **RabbitMQ** (async event processing via `rabbitmq/amqp091-go`)
- **Stripe Go SDK** (v84)
- **golang-jwt/jwt**
- **pquerna/otp** (TOTP 2FA)
- **golang-migrate** (SQL migrations for both PostgreSQL and ClickHouse)
- **wneessen/go-mail** (SMTP client)
- **nicksnyder/go-i18n** (email template i18n)
- **oschwald/geoip2-golang** (MaxMind GeoIP2 country lookup)

### Frontend
- **Nuxt 3** (Vue 3, TypeScript)
- **Tailwind CSS**
- **Pinia** (auth store)
- **TanStack Vue Query** (data fetching, cache management)
- **vee-validate + zod** (form validation)
- **@stripe/stripe-js** (Stripe Elements)
- **@nuxtjs/i18n** (EN / RU / DE)

## Project Structure

```
.
├── cmd/
│   ├── server/                    # HTTP server entry point
│   ├── consumers/
│   │   ├── paymentIntent/         # PaymentIntent event consumer
│   │   ├── subscription/          # Subscription event consumer
│   │   ├── invoice/               # Invoice event consumer
│   │   └── stats/                 # Click stats consumer (writes to ClickHouse, invalidates Redis cache)
│   └── shared/                    # Shared consumer loop & config loader
├── configs/                       # Configuration loading
├── internal/
│   ├── account/                   # Stripe customer account management + 2FA service
│   ├── auth/                      # Registration & login handlers
│   ├── auth_session/              # Session create/update (token, IP, user agent, is_verify)
│   ├── consts/                    # RabbitMQ exchange/queue/routing constants
│   ├── consumers/                 # Consumer handler logic (paymentIntent, subscription, invoice, stats)
│   ├── jwt/                       # JWT service
│   ├── locales/                   # Embedded email templates & i18n strings (EN/RU/DE)
│   │   ├── auth/register/         # Welcome email (welcome.html + *.toml)
│   │   └── invoice/succeed/       # Payment success email (payment_success.html + *.toml)
│   ├── mailer/                    # SMTP mailer service (go-mail + go-i18n)
│   ├── link/                      # Link CRUD + redirect + stats publishing
│   ├── models/                    # Shared message/event models
│   ├── publishers/                # RabbitMQ publishers (payment, subscription, invoice, stats)
│   ├── stats/                     # Click statistics (ClickHouse repo, handler, service, cache)
│   ├── user/                      # User repository + 2FA handlers
│   └── payments/
│       ├── invoice/               # Invoice repository & service
│       ├── payment/               # Payment repository & handler (GET /payments)
│       ├── plan/                  # Plans (GET /plans)
│       ├── stripe/                # Stripe handlers & services
│       │   └── services/          # PaymentIntent, CustomerAccount helpers
│       ├── subscription/          # Subscription service, handlers
│       └── webhook/               # Stripe webhook handler & service
├── pkg/
│   ├── clickhouse/                # ClickHouse connection wrapper
│   ├── db/                        # PostgreSQL connection
│   ├── event/                     # In-process event bus
│   ├── middleware/                # CORS, auth, logging, subscription check
│   ├── rabbitMq/                  # RabbitMQ client wrapper
│   ├── redis/                     # Redis client wrapper (Get/Set/Incr/SAdd/DelByPattern)
│   ├── request/                   # Body parsing helpers
│   └── response/                  # JSON response helpers
├── migrations/
│   ├── postgres/                  # PostgreSQL migration files
│   ├── clickhouse/                # ClickHouse migration files
│   └── auto.go                    # Migration runner (supports --target postgres|clickhouse|all)
├── frontend/
│   ├── apps/web/                  # Nuxt 3 SPA
│   │   ├── pages/                 # dashboard, billing, payments, account, auth
│   │   ├── components/            # UI, billing, auth, links components
│   │   ├── services/              # API clients (auth, account, subscription, payment, stats)
│   │   ├── stores/                # Pinia stores (auth)
│   │   ├── schemas/               # Zod validation schemas
│   │   └── i18n/locales/          # en.json, ru.json, de.json
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
- PostgreSQL
- ClickHouse
- Redis
- RabbitMQ
- Node.js 20+ and pnpm
- Docker & Docker Compose
- Stripe account (test keys)
- MaxMind GeoLite2 Country `.mmdb` file (optional, for country detection)

### Environment Variables

Copy `.env.example` to `.env`:

```env
DSN="host=localhost user=postgres password=pass dbname=linkshort port=5432 sslmode=disable"
DATABASE_URL="pgx5://postgres:pass@localhost:5432/linkshort"
TOKEN="your_jwt_secret"
REDIS_URL="localhost:6379"
REDIS_CACHE_MINUTES=5
STRIPE_KEY="sk_test_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
RABBITMQ_CONSUMERS=1

CLICKHOUSE_ADDR=localhost:9000
CLICKHOUSE_DB=default
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

# GeoIP — optional, country detection is disabled if unset
GEOIP_PATH=/path/to/GeoLite2-Country.mmdb

# Mailer (SMTP) — optional, emails are skipped if SMTP_HOST is empty
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=no-reply@example.com
```

`RABBITMQ_CONSUMERS` controls how many worker instances start per consumer type in `make dev` / `make consumers`.

If `SMTP_HOST` is left empty, the mailer is disabled and no emails are sent.

If `GEOIP_PATH` is not set, country detection is silently skipped and the `country` field is stored as an empty string.

### Running with Docker + Make

```bash
# Start infrastructure (Postgres, ClickHouse, Redis, RabbitMQ) + server + all consumers + frontend
make dev

# Or step by step:
make docker-up           # start infrastructure
make migrate             # run all migrations (postgres + clickhouse)
make migrate-postgres    # postgres only
make migrate-clickhouse  # clickhouse only
make server              # HTTP server only (port 8081)
make consumers           # all consumer workers
make consumer-payment
make consumer-subscription
make consumer-invoice
make consumer-stats
make frontend            # Nuxt dev server
make stripe-webhook      # listen for Stripe webhooks (Stripe CLI)

# Scale consumers (e.g. 3 workers each)
make consumers WORKERS=3
```

> In `make dev`, consumers wait for the HTTP server to be ready on port 8081 before starting.

### Running Locally (without Make)

```bash
# Backend server
go run ./cmd/server

# Consumers (each in separate terminal)
go run ./cmd/consumers/paymentIntent
go run ./cmd/consumers/subscription
go run ./cmd/consumers/invoice
go run ./cmd/consumers/stats

# Frontend
cd frontend && pnpm install && pnpm dev
```

### Build Binaries

```bash
make build
# Produces: bin/server, bin/consumer-payment, bin/consumer-subscription, bin/consumer-invoice, bin/consumer-stats
```

### Database Migrations

```bash
# Run all (postgres + clickhouse)
go run ./migrations

# Postgres only
go run ./migrations -target=postgres

# ClickHouse only
go run ./migrations -target=clickhouse
```

## Async Event Architecture

Stripe webhook events and click tracking are handled asynchronously via RabbitMQ:

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
                                                          (upsert invoice + payment record,
                                                           send payment confirmation email)

GET /{hash} → redirect
           │
           └── link.visited event → Stats exchange → statsQueue
                                                          │
                                                    Stats consumer
                                                    (insert into ClickHouse link_clicks,
                                                     invalidate Redis stats cache)
```

Each consumer binary:
1. Creates its own exchange and queue on startup (idempotent)
2. Processes messages with manual Ack/Nack
3. Nacks on error (message re-queued for retry), except Stripe 400 errors (discarded)

## Click Statistics

Click events are captured on every redirect (`GET /{hash}`) and published to RabbitMQ. The stats consumer writes them to ClickHouse and invalidates the Redis cache for that link.

Each `link_clicks` row captures:
- `link_id`, `clicked_at`
- Network: `ip`, `country`, `forwarded_for`, `real_ip`, `remote_addr`
- Headers: `user_agent`, `referer`, `accept_language`, `origin`
- Device: `device_type`, `os`, `browser`
- Security: `fingerprint`, `request_id`, `scheme`

The `country` field is populated via MaxMind GeoLite2 if `GEOIP_PATH` is configured.

Stats are cached in Redis per `(email, linkID, date-range)`. The cache is invalidated by pattern `*:link:{id}:*` whenever new clicks are inserted.

## API Endpoints

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register → `{email, token}` |
| POST | `/auth/login` | No | Login → `{email, token, is2faEnabled}` |

### User & 2FA

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/users/me` | Yes | Current user info |
| POST | `/users/me/2fa/setup` | Yes | Generate TOTP secret → `{qrCode}` |
| POST | `/users/2fa/verify` | No | Verify TOTP code → `{token}` |

### Account

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/account` | Yes | Create Stripe customer account |
| PATCH | `/account` | Yes | Update account name/email |

### Links

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/links` | Yes | List user's links |
| POST | `/link` | Yes + Active Sub | Create shortened link |
| PATCH | `/link/{id}` | Yes | Update link |
| DELETE | `/link/{id}` | Yes | Delete link |
| GET | `/{hash}` | No | Redirect to original URL (publishes click event) |

### Plans & Subscriptions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/plans` | Yes | List available plans |
| GET | `/subscriptions/me` | Yes | Current subscription |
| POST | `/subscriptions/method` | Yes | Create SetupIntent → `{clientSecret}` |
| POST | `/subscriptions` | Yes | Create subscription `{planId}` |
| PATCH | `/subscriptions/cancel` | Yes | Cancel subscription |

### Payments

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/payments` | Yes | List all payments for current user |
| POST | `/stripe/paymentIntent` | Yes | Create PaymentIntent |
| POST | `/stripe/paymentIntent/confirm` | Yes | Confirm PaymentIntent |
| POST | `/stripe/paymentIntent/cancel` | Yes | Refund a PaymentIntent-based payment |
| POST | `/stripe/webhook` | No | Stripe webhook handler |

### Statistics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/stats` | Yes | Raw click events `?from=&to=&linkId=` |
| GET | `/stats/link/{id}` | Yes | Clicks grouped by date `?from=&to=` |

## Auth Flow

### Login without 2FA

1. `POST /auth/login` → returns `{email, token, is2faEnabled: false}`
2. Token is stored — user is authenticated

### Login with 2FA

1. `POST /auth/login` → returns `{email, is2faEnabled: true}`, token is empty
2. Frontend prompts for TOTP code
3. `POST /users/2fa/verify {email, code}` → returns `{token}`
4. Token is stored — user is authenticated

### 2FA Setup

1. `POST /users/me/2fa/setup` (authenticated) → returns base64 QR code PNG
2. User scans QR code with authenticator app
3. 2FA is active at next login

### Session Tracking

Every login creates an `auth_sessions` record with token, expiry, IP address, user agent, and `is_verify` flag. The `IsAuthed` middleware validates the session on every protected request.

## Stripe Payment Flows

### Subscription flow

1. `POST /account` — create Stripe customer
2. `POST /subscriptions/method` — create SetupIntent, get `clientSecret`
3. `stripe.confirmCardSetup(clientSecret, {card})` — confirm card on frontend
4. `POST /subscriptions {planId}` — create subscription
5. Stripe fires `customer.subscription.created` → consumer creates local subscription + initial invoice
6. Stripe fires `invoice.payment_succeeded` → consumer upserts invoice + payment record + sends payment confirmation email

### PaymentIntent flow (one-time)

1. `POST /stripe/paymentIntent` — create PaymentIntent
2. `stripe.confirmCardPayment(clientSecret, {card})` — confirm payment
3. Stripe fires `payment_intent.succeeded` → consumer syncs payment record

### Cancellation

- **Subscription** (`isPaymentIntent: false`): `PATCH /subscriptions/cancel`
- **PaymentIntent payment** (`isPaymentIntent: true`): `POST /stripe/paymentIntent/cancel` (issues Stripe Refund)

> **Note**: A succeeded PaymentIntent cannot be canceled — the Refund API must be used instead.

## Frontend Pages

| Route | Description |
|-------|-------------|
| `/auth/login` | Login form (with 2FA code step if enabled) |
| `/auth/register` | Registration form (auto-creates account after signup) |
| `/dashboard` | User's links list with per-link stats drawer |
| `/billing` | Plans, subscription status, cancel button |
| `/payments` | Payment history table |
| `/account` | Profile info and 2FA setup |

## Development Notes

- **Auth middleware** (`middleware/auth.global.ts`) — redirects unauthenticated users
- **Query cache** — cleared on login/logout to prevent stale data across sessions
- **`GET /subscriptions/me`** — returns `204 No Content` when user has no subscription
- **`isPaymentIntent`** field — `true` when `billing_id` starts with `pi_`, drives cancel button behavior
- **Consumer startup** — each consumer creates its own RabbitMQ exchange/queue on init, safe to start independently
- **Mailer** — `internal/mailer/` is a shared SMTP client; `AuthMailer` sends welcome email on registration; `InvoiceConsumer` sends payment confirmation asynchronously (goroutine) after invoice paid; if `SMTP_HOST` is empty, all sends are silently skipped
- **Email i18n** — templates live in `internal/locales/` (embedded via `embed.FS`), translated via `go-i18n` + TOML files (EN/RU/DE)
- **TOTP** — 30-second window, ±5 second skew tolerance, SHA1, 6-digit codes via `pquerna/otp`
- **ClickHouse table** — `link_clicks` uses `ReplacingMergeTree`, partitioned by month, ordered by `(link_id, clicked_at, request_id)`; deduplication is eventual
- **Stats cache invalidation** — when the stats consumer inserts into ClickHouse, it calls `redis.DelByPattern("*:link:{id}:*")` to clear all cached reports for that link

## License

MIT
