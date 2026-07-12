BEGIN;

DROP INDEX IF EXISTS idx_models_id_world_id;

ALTER TABLE models DROP CONSTRAINT IF EXISTS models_pkey;

ALTER TABLE models ADD PRIMARY KEY (id);

CREATE INDEX IF NOT EXISTS idx_models_world_id ON models(world_id);

COMMIT;
