BEGIN;

ALTER TABLE zones_subscriptions
  ADD COLUMN IF NOT EXISTS is_admin_granted BOOLEAN NOT NULL DEFAULT false;

COMMIT;
