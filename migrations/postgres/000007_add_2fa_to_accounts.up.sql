ALTER TABLE accounts
    ADD COLUMN totp_secret    TEXT,
    ADD COLUMN is_2fa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN backup_codes   TEXT[];
