ALTER TABLE plans
    ADD COLUMN stripe_price_id   TEXT NOT NULL DEFAULT '',
    ADD COLUMN stripe_product_id TEXT NOT NULL DEFAULT '';
