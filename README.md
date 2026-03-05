# LinkShort

A full-stack URL shortener service with subscription billing, built with Go and Nuxt 3.

## Features

- **URL Shortening**: Create and manage shortened links with click tracking
- **JWT Authentication**: Secure authentication with per-request authorization
- **Subscription Billing**: Stripe integration — plans, payment methods, subscriptions, refunds
- **Payment History**: Full audit trail of all payments per user
- **Click Statistics**: Time-range and per-link analytics
- **Event-Driven Architecture**: Asynchronous statistics via event bus
- **Redis Caching**: Stats cache for performance
- **Internationalization**: EN / RU / DE (frontend)

## Tech Stack

### Backend
- **Go** (net/http ServeMux)
- **PostgreSQL** with GORM
- **Redis** (stats caching)
- **Stripe Go SDK** (v84)
- **golang-jwt/jwt**
- **golang-migrate** (SQL migrations)

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
│   └── main.go                    # App entry point, dependency wiring
├── configs/                       # Configuration loading
├── internal/
│   ├── account/                   # Stripe customer account management
│   ├── auth/                      # Registration & login
│   ├── jwt/                       # JWT service
│   ├── link/                      # Link CRUD + redirect
│   ├── stats/                     # Click statistics
│   ├── user/                      # User repository
│   └── payments/
│       ├── models/                # Payment DB model
│       ├── payment/               # Payment repository & handler (GET /payments)
│       ├── plan/                  # Plans (GET /plans)
│       ├── stripe/                # Stripe handlers & services
│       │   └── services/          # PaymentIntent, Refund helpers
│       ├── subscription/          # Subscription service, handlers
│       └── webhook/               # Stripe webhook handler
├── pkg/
│   ├── db/                        # Database connection
│   ├── event/                     # Event bus
│   ├── middleware/                 # CORS, auth, subscription check
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
├── docker-compose.yml
└── .env.example
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- Redis
- Node.js 20+ and pnpm
- Docker & Docker Compose (optional)
- Stripe account (test keys)

### Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```env
DSN="host=localhost user=postgres password=pass dbname=linkshort port=5432 sslmode=disable"
TOKEN="your_jwt_secret"
REDIS_URL="localhost:6379"
STRIPE_KEY="sk_test_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
```

### Running with Docker

```bash
docker-compose up -d
```

The API starts on `http://localhost:8081`.

### Running Locally

```bash
# Backend
go mod download
go run cmd/main.go

# Frontend (from project root)
cd frontend
pnpm install
pnpm dev
```

Frontend runs on `http://localhost:3000`.

### Database Migrations

```bash
migrate -path migrations/sql -database "$DSN" up
```

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
| GET | `/subscriptions/me` | Yes | Current subscription (returns `isPaymentIntent` flag) |
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

### PaymentIntent flow (one-time)

1. `POST /stripe/paymentIntent` — create PaymentIntent
2. `stripe.confirmCardPayment(clientSecret, {card})` — confirm payment
3. Webhook `payment_intent.succeeded` → subscription activated

### Cancellation

- **Subscription** (`isPaymentIntent: false`): `PATCH /subscriptions/cancel`
- **PaymentIntent payment** (`isPaymentIntent: true`): `POST /stripe/paymentIntent/cancel` (issues Stripe Refund, then marks subscription/payment canceled)

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
- **Query cache** — cleared on login and logout to prevent stale data across user sessions
- **`GET /subscriptions/me`** — returns `204 No Content` when user has no subscription
- **`isPaymentIntent`** field — `true` when `billing_id` starts with `pi_`, drives cancel button behavior

## Roadmap

### 1. RabbitMQ — Email Invoice Delivery

Send invoice emails asynchronously after successful payments.

- Add RabbitMQ as a message broker (Docker service)
- Publish `invoice.created` event from the Stripe webhook handler on `payment_intent.succeeded` / `invoice.paid`
- Create a dedicated consumer service that reads the queue and sends emails via SMTP / SendGrid
- Email template: payment amount, plan name, period, PDF invoice link from Stripe
- Dead-letter queue for failed delivery retries

**Key files to touch**: `internal/payments/webhook/service.go`, new `internal/mailer/`

---

### 2. Extended Auth — User Agent, IP, Session Table

Capture device and network context on every login and registration.

- Create `auth_sessions` table: `id`, `user_id`, `ip`, `user_agent`, `created_at`
- Parse `X-Forwarded-For` / `RemoteAddr` and `User-Agent` headers in login and register handlers
- Save a session row on each successful auth event
- Expose `GET /auth/sessions` — list of recent sessions for the current user (useful for security audit UI)
- Consider adding `revoked_at` for active-session management / forced logout

**Key files to touch**: `internal/auth/`, new `internal/auth/session/`, `migrations/sql/`

---

### 3. Columnar DB for Click Statistics

Replace PostgreSQL click writes with a columnar store for analytical queries.

- Introduce **ClickHouse** (or **TimescaleDB** as a lighter alternative) as a second data store
- On every redirect (`GET /{hash}`), publish a click event: `link_id`, `timestamp`, `ip`, `user_agent`, `referrer`, `country` (via IP geolocation)
- Write to ClickHouse via the existing event bus (new consumer alongside the current PostgreSQL consumer)
- Rewrite `GET /stats` and `GET /stats/clicks` to query ClickHouse — aggregations become orders of magnitude faster at scale
- Keep PostgreSQL only for relational data (users, links, subscriptions, payments)

**Key files to touch**: `internal/stats/`, `pkg/event/`, new `pkg/clickhouse/`

---

### 4. Rate Limiting (RPS)

Protect the API from abuse and enforce fair-use quotas.

- Add a **token-bucket** rate limiter middleware using Redis (`go-redis` + Lua script or `golang.org/x/time/rate` for in-process)
- Two tiers:
  - **Global**: e.g. 1000 req/s per IP on public endpoints (`/{hash}`, `/auth/*`)
  - **Per-user**: e.g. 60 req/min on authenticated endpoints, configurable per plan
- Return `429 Too Many Requests` with `Retry-After` header
- Store counters in Redis with TTL — stateless across instances

**Key files to touch**: `pkg/middleware/`, `cmd/main.go`

---

### 5. Scaling Strategy, Deployment & Monitoring

Make the service production-ready and observable.

#### Scaling Strategy

- **Horizontal scaling**: stateless Go instances behind a load balancer (HAProxy / Nginx / AWS ALB)
- **Database**: PostgreSQL read replicas for analytics queries; connection pooling via PgBouncer
- **Redis**: Redis Cluster or Sentinel for HA; separate Redis instance per environment
- **ClickHouse**: single-node for MVP, ClickHouse Keeper + sharding for high load
- **RabbitMQ**: mirrored queues or RabbitMQ Cluster for durability
- **CDN**: static frontend assets via CDN (Cloudflare / AWS CloudFront); redirect endpoint behind CDN with short cache TTL

#### Deployment

- **Containerization**: multi-stage Dockerfile for the Go binary (scratch base image, ~10 MB)
- **Orchestration**: Docker Compose for local dev; Kubernetes manifests (Deployment, Service, HPA) for production
- **CI/CD**: GitHub Actions pipeline — lint → test → build image → push to registry → rolling deploy
- **Migrations**: run as a Kubernetes Job / init container before app pods start
- **Secrets**: environment variables injected via Kubernetes Secrets or HashiCorp Vault

#### Monitoring & Observability

- **Metrics**: expose `GET /metrics` (Prometheus format) — request count, latency histograms, error rates, active subscriptions
- **Tracing**: OpenTelemetry SDK → Jaeger / Tempo for distributed traces across Go + RabbitMQ consumers
- **Logging**: structured JSON logs (`log/slog`) with `request_id`, `user_id`, `duration` fields → aggregated in Loki / CloudWatch
- **Dashboards**: Grafana — API latency P99, click throughput, payment success rate, queue depth
- **Alerting**: Alertmanager rules — error rate > 1%, P99 > 500ms, queue depth > 10k, failed payments spike

---

## License

MIT
