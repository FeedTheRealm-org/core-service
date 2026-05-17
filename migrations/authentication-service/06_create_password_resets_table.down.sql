BEGIN;

DROP INDEX IF EXISTS idx_password_resets_reset_token_hash;
DROP INDEX IF EXISTS idx_password_resets_user_id;
DROP TABLE IF EXISTS password_resets;

COMMIT;
