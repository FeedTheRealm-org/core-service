BEGIN;

ALTER TABLE users
ADD COLUMN IF NOT EXISTS refresh_token_updated_at TIMESTAMPTZ DEFAULT NOW();

UPDATE users
SET refresh_token_updated_at = NOW()
WHERE refresh_token_updated_at IS NULL;

COMMIT;
