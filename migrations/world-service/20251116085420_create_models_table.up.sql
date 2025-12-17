BEGIN;

CREATE TABLE IF NOT EXISTS models (
  model_id UUID NOT NULL,
  world_id UUID NOT NULL,
  name TEXT NOT NULL,
  model_url TEXT NOT NULL,
  material_url TEXT,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_models_world_id ON models(world_id);

COMMIT;
