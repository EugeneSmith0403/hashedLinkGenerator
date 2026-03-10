-- roles
CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT
);
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);

INSERT INTO roles (name, description, created_at, updated_at)
VALUES
    ('admin',   'Full access to all resources and management',           NOW(), NOW()),
    ('creator', 'Can create and manage subscriptions and links',         NOW(), NOW()),
    ('viewer',  'Read-only access to resources',                         NOW(), NOW());

-- add role_id to accounts
ALTER TABLE accounts ADD COLUMN role_id BIGINT REFERENCES roles (id);
CREATE INDEX idx_accounts_role_id ON accounts (role_id);

-- features
CREATE TABLE features (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    key         VARCHAR(100) NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_features_deleted_at ON features (deleted_at);
CREATE INDEX idx_features_key ON features (key);

-- auth_sessions
CREATE TABLE auth_sessions (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    account_id  BIGINT      NOT NULL REFERENCES accounts (id),
    token       TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    ip_address  TEXT,
    user_agent  TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_auth_sessions_account_id ON auth_sessions (account_id);
CREATE INDEX idx_auth_sessions_token ON auth_sessions (token);
CREATE INDEX idx_auth_sessions_deleted_at ON auth_sessions (deleted_at);
