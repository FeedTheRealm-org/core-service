BEGIN;

ALTER TABLE accounts
DROP COLUMN IF EXISTS expiration_verify_code;

COMMIT;
