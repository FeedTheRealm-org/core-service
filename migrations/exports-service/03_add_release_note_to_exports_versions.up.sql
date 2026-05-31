BEGIN;

ALTER TABLE exports_versions
ADD COLUMN IF NOT EXISTS release_note TEXT NOT NULL DEFAULT 'no release note provided.';

UPDATE exports_versions
SET release_note = 'no release note provided.'
WHERE release_note IS NULL OR release_note = '';

COMMIT;
