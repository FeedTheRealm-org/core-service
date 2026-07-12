BEGIN;

ALTER TABLE models DROP CONSTRAINT IF EXISTS models_pkey;

DROP INDEX IF EXISTS idx_models_world_id;

ALTER TABLE models ADD PRIMARY KEY (world_id, id);

CREATE INDEX IF NOT EXISTS idx_models_id_world_id ON models(id, world_id);

COMMIT;
