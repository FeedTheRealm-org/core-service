BEGIN;

ALTER TABLE account_verifications DROP CONSTRAINT IF EXISTS fk_user;
DROP TABLE IF EXISTS account_verifications;

COMMIT;
