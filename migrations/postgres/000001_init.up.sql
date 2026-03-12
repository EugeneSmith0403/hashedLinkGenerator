-- links
CREATE TABLE links (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    url        TEXT,
    hash       TEXT
);
CREATE UNIQUE INDEX idx_links_hash ON links (hash);
CREATE INDEX idx_links_deleted_at ON links (deleted_at);

-- users
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    email      TEXT,
    password   TEXT,
    name       TEXT
);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

-- stats
CREATE TABLE stats (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    link_id    BIGINT REFERENCES links (id),
    clicks     INTEGER DEFAULT 0,
    date       DATE
);
CREATE UNIQUE INDEX idx_stats_link_date ON stats (link_id, date);
CREATE INDEX idx_stats_deleted_at ON stats (deleted_at);

-- accounts
CREATE TABLE accounts (
    id             BIGSERIAL PRIMARY KEY,
    created_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ,
    user_id        BIGINT,
    account_status VARCHAR(20) NOT NULL,
    provider       VARCHAR(20) NOT NULL,
    customer_id    TEXT,
    banned_by      TEXT,
    banned_at      TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_account_user_id ON accounts (user_id);
CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at);

-- plans
CREATE TABLE plans (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    is_active  BOOLEAN,
    period     TIMESTAMPTZ,
    name       TEXT,
    cost       REAL,
    currency   TEXT,
    features   JSONB
);
CREATE INDEX idx_plans_deleted_at ON plans (deleted_at);

-- subscriptions
CREATE TABLE subscriptions (
    id                   BIGSERIAL PRIMARY KEY,
    created_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ,
    deleted_at           TIMESTAMPTZ,
    user_id              BIGINT      NOT NULL,
    plan_id              BIGINT      NOT NULL,
    billing_id           TEXT        NOT NULL,
    customer_id          TEXT        NOT NULL,
    status               VARCHAR(30) NOT NULL,
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    cancel_at            TIMESTAMPTZ,
    canceled_at          TIMESTAMPTZ,
    trial_start          TIMESTAMPTZ,
    trial_end            TIMESTAMPTZ,
    provider_metadata    JSONB
);
CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions (plan_id);
CREATE UNIQUE INDEX idx_subscriptions_billing_id ON subscriptions (billing_id);
CREATE INDEX idx_subscriptions_customer_id ON subscriptions (customer_id);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions (deleted_at);

-- invoices
CREATE TABLE invoices (
    id                BIGSERIAL PRIMARY KEY,
    created_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ,
    account_id        BIGINT      NOT NULL,
    subscription_id   BIGINT,
    billing_id        TEXT        NOT NULL,
    status            VARCHAR(20) NOT NULL,
    amount_due        BIGINT,
    amount_paid       BIGINT,
    amount_remaining  BIGINT,
    currency          VARCHAR(3)  NOT NULL,
    due_date          TIMESTAMPTZ,
    paid_at           TIMESTAMPTZ,
    hosted_invoice_url TEXT,
    invoice_pdf       TEXT,
    provider_metadata JSONB
);
CREATE INDEX idx_invoices_account_id ON invoices (account_id);
CREATE INDEX idx_invoices_subscription_id ON invoices (subscription_id);
CREATE UNIQUE INDEX idx_invoices_billing_id ON invoices (billing_id);
CREATE INDEX idx_invoices_deleted_at ON invoices (deleted_at);

-- payments
CREATE TABLE payments (
    id                   UUID        PRIMARY KEY,
    created_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ,
    deleted_at           TIMESTAMPTZ,
    account_id           BIGINT      NOT NULL,
    invoice_id           BIGINT,
    payment_intent_id    TEXT,
    charge_id            TEXT,
    amount               BIGINT,
    platform_fee         BIGINT,
    net_amount           BIGINT,
    currency             VARCHAR(3)  NOT NULL,
    status               VARCHAR(40) NOT NULL,
    payment_method_type  TEXT,
    failure_code         TEXT,
    failure_message      TEXT,
    provider_metadata    JSONB
);
CREATE INDEX idx_payments_deleted_at ON payments (deleted_at);
CREATE INDEX idx_payments_account_id ON payments (account_id);
CREATE INDEX idx_payments_invoice_id ON payments (invoice_id);
CREATE UNIQUE INDEX idx_payments_payment_intent_id ON payments (payment_intent_id);
CREATE UNIQUE INDEX idx_payments_charge_id ON payments (charge_id);
