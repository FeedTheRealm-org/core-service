BEGIN;

CREATE TABLE IF NOT EXISTS models (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  model_id UUID NOT NULL,
  world_id UUID NOT NULL,
  name TEXT NOT NULL,
  model_url TEXT NOT NULL,
  prefab_url TEXT,
  metadata_url TEXT,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
