BEGIN;

ALTER TABLE exports_versions
DROP COLUMN IF EXISTS release_note;

COMMIT;
