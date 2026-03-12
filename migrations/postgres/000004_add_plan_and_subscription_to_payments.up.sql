ALTER TABLE payments ADD COLUMN IF NOT EXISTS plan_id BIGINT REFERENCES plans (id);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS subscription_id BIGINT REFERENCES subscriptions (id);

CREATE INDEX IF NOT EXISTS idx_payments_plan_id ON payments (plan_id);
CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments (subscription_id);
