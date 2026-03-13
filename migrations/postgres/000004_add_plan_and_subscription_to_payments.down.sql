DROP INDEX IF EXISTS idx_payments_subscription_id;
DROP INDEX IF EXISTS idx_payments_plan_id;

ALTER TABLE payments DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE payments DROP COLUMN IF EXISTS plan_id;
