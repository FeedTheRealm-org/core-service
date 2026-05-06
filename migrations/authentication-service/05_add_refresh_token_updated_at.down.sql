BEGIN;

ALTER TABLE users
DROP COLUMN IF EXISTS refresh_token_updated_at;

COMMIT;
