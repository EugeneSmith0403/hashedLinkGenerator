# LinkShort

A full-stack URL shortener service with subscription billing, built with Go and Nuxt 3.

## Features

- **URL Shortening**: Create and manage shortened links with click tracking
- **JWT Authentication**: Secure authentication with per-request authorization
- **Subscription Billing**: Stripe integration — plans, payment methods, subscriptions, refunds
- **Payment History**: Full audit trail of all payments per user
- **Click Statistics**: Time-range and per-link analytics
- **Async Event Processing**: RabbitMQ consumers for payment intents, subscriptions, and invoices
- **Transactional Email**: SMTP mailer — welcome email on registration, payment confirmation on invoice paid
- **Redis Caching**: Stats cache for performance
- **Internationalization**: EN / RU / DE (frontend + email templates)

## Tech Stack

### Backend
- **Go** (net/http ServeMux)
- **PostgreSQL** with GORM
- **Redis** (stats caching)
- **RabbitMQ** (async event processing via `rabbitmq/amqp091-go`)
- **Stripe Go SDK** (v84)
- **golang-jwt/jwt**
- **golang-migrate** (SQL migrations)
- **wneessen/go-mail** (SMTP client)
- **nicksnyder/go-i18n** (email template i18n)

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
│   │   └── invoice/               # Invoice event consumer
│   └── shared/                    # Shared consumer loop & config loader
├── configs/                       # Configuration loading
├── internal/
│   ├── account/                   # Stripe customer account management
│   ├── auth/                      # Registration & login
│   ├── consts/                    # RabbitMQ exchange/queue/routing constants
│   ├── consumers/                 # Consumer handler logic (paymentIntent, subscription, invoice)
│   ├── jwt/                       # JWT service
│   ├── locales/                   # Embedded email templates & i18n strings (EN/RU/DE)
│   │   ├── auth/register/         # Welcome email (welcome.html + *.toml)
│   │   └── invoice/succeed/       # Payment success email (payment_success.html + *.toml)
│   ├── mailer/                    # SMTP mailer service (go-mail + go-i18n)
│   ├── link/                      # Link CRUD + redirect
│   ├── models/                    # Shared message/event models
│   ├── publishers/                # RabbitMQ publishers (payment, subscription, invoice)
│   ├── stats/                     # Click statistics
│   ├── user/                      # User repository
│   └── payments/
│       ├── invoice/               # Invoice repository & service
│       ├── payment/               # Payment repository & handler (GET /payments)
│       ├── plan/                  # Plans (GET /plans)
│       ├── stripe/                # Stripe handlers & services
│       │   └── services/          # PaymentIntent, CustomerAccount helpers
│       ├── subscription/          # Subscription service, handlers
│       └── webhook/               # Stripe webhook handler & service
├── pkg/
│   ├── db/                        # Database connection
│   ├── event/                     # In-process event bus
│   ├── middleware/                 # CORS, auth, subscription check
│   ├── rabbitMq/                  # RabbitMQ client wrapper
│   ├── request/                   # Body parsing helpers
│   └── response/                  # JSON response helpers
├── migrations/sql/                # SQL migration files
├── frontend/
│   ├── apps/web/                  # Nuxt 3 SPA
│   │   ├── pages/                 # dashboard, billing, payments, auth
│   │   ├── components/            # UI, billing, auth components
│   │   ├── services/              # API clients (auth, account, subscription, payment)
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

- Go 1.21+
- PostgreSQL
- Redis
- RabbitMQ
- Node.js 20+ and pnpm
- Docker & Docker Compose
- Stripe account (test keys)

### Environment Variables

Copy `.env.example` to `.env`:

```env
DSN="host=localhost user=postgres password=pass dbname=linkshort port=5432 sslmode=disable"
TOKEN="your_jwt_secret"
REDIS_URL="localhost:6379"
STRIPE_KEY="sk_test_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
RABBITMQ_CONSUMERS=1

# Mailer (SMTP) — optional, emails are skipped if SMTP_HOST is empty
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=no-reply@example.com
```

`RABBITMQ_CONSUMERS` controls how many worker instances start per consumer type in `make dev` / `make consumers`.

If `SMTP_HOST` is left empty, the mailer is disabled and no emails are sent.

### Running with Docker + Make

```bash
# Start infrastructure (Postgres, Redis, RabbitMQ) + server + all consumers + frontend
make dev

# Or step by step:
make docker-up       # start infrastructure
make server          # HTTP server only (port 8081)
make consumers       # all consumer workers
make consumer-payment
make consumer-subscription
make consumer-invoice

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

# Frontend
cd frontend && pnpm install && pnpm dev
```

### Build Binaries

```bash
make build
# Produces: bin/server, bin/consumer-payment, bin/consumer-subscription, bin/consumer-invoice
```

### Database Migrations

```bash
migrate -path migrations/sql -database "$DSN" up
```

## Async Event Architecture

Stripe webhook events are handled asynchronously via RabbitMQ:

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
```

Each consumer binary:
1. Creates its own exchange and queue on startup (idempotent)
2. Processes messages with manual Ack/Nack
3. Nacks on error (message re-queued for retry), except Stripe 400 errors (discarded)

## API Endpoints

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register → `{email, token}` |
| POST | `/auth/login` | No | Login → `{email, token}` |

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
| GET | `/{hash}` | No | Redirect to original URL |

### Plans & Subscriptions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/plans` | No | List available plans |
| GET | `/subscriptions/me` | Yes | Current subscription |
| POST | `/subscriptions/method` | Yes | Create SetupIntent → `{clientSecret}` |
| POST | `/subscriptions` | Yes | Create subscription `{planId}` |
| PATCH | `/subscriptions/cancel` | Yes | Cancel subscription |

### Payments

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/payments` | Yes | List all payments for current user |
| POST | `/stripe/paymentIntent` | Yes | Create PaymentIntent |
| POST | `/stripe/paymentIntent/cancel` | Yes | Refund a PaymentIntent-based payment |
| POST | `/stripe/webhook` | No | Stripe webhook handler |

### Statistics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/stats` | Yes | Link stats `?from=&to=&linkId=` |
| GET | `/stats/clicks` | Yes | Total clicks `?from=&to=` |

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
| `/auth/login` | Login form |
| `/auth/register` | Registration form (auto-creates account after signup) |
| `/dashboard` | User's links list |
| `/billing` | Plans, subscription status, cancel button |
| `/payments` | Payment history table |

## Development Notes

- **Auth middleware** (`middleware/auth.global.ts`) — redirects unauthenticated users
- **Query cache** — cleared on login/logout to prevent stale data across sessions
- **`GET /subscriptions/me`** — returns `204 No Content` when user has no subscription
- **`isPaymentIntent`** field — `true` when `billing_id` starts with `pi_`, drives cancel button behavior
- **Consumer startup** — each consumer creates its own RabbitMQ exchange/queue on init, safe to start independently
- **Mailer** — `internal/mailer/` is a shared SMTP client; `AuthMailer` sends welcome email on registration; `InvoiceConsumer` sends payment confirmation asynchronously (goroutine) after invoice paid; if `SMTP_HOST` is empty, all sends are silently skipped
- **Email i18n** — templates live in `internal/locales/` (embedded via `embed.FS`), translated via `go-i18n` + TOML files (EN/RU/DE)

## Roadmap

### 1. Extended Auth — User Agent, IP, Session Table

- Create `auth_sessions` table: `id`, `user_id`, `ip`, `user_agent`, `created_at`
- Parse headers in login/register handlers
- Expose `GET /auth/sessions` for security audit UI

---

### 2. Columnar DB for Click Statistics

- Introduce **ClickHouse** or **TimescaleDB** for analytical queries
- Publish click events with geo/device metadata
- Rewrite `GET /stats` to query ClickHouse

---

### 3. Rate Limiting (RPS)

- Token-bucket limiter via Redis
- Global IP-based + per-user quotas configurable per plan
- `429 Too Many Requests` with `Retry-After`

---

### 4. Scaling & Monitoring

- Horizontal scaling with stateless Go instances
- Prometheus metrics endpoint, OpenTelemetry tracing
- Grafana dashboards — latency, queue depth, payment success rate

---

## License

MIT
