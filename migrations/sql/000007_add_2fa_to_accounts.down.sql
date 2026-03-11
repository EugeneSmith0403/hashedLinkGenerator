ALTER TABLE accounts
    DROP COLUMN IF EXISTS totp_secret,
    DROP COLUMN IF EXISTS is_2fa_enabled,
    DROP COLUMN IF EXISTS backup_codes;
