BEGIN;

ALTER TABLE zones_subscriptions
  DROP COLUMN IF EXISTS is_admin_granted;

COMMIT;
