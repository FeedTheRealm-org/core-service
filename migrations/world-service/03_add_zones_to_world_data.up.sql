BEGIN;

ALTER TABLE IF EXISTS world_data
  ADD COLUMN IF NOT EXISTS createable_data jsonb NOT NULL DEFAULT '{}'::jsonb;

CREATE TABLE IF NOT EXISTS world_zones (
  id integer NOT NULL,
  world_id UUID NOT NULL,
  zone_data jsonb NOT NULL,
  PRIMARY KEY (world_id, id),
  CONSTRAINT fk_world_zones_world_data FOREIGN KEY (world_id)
    REFERENCES world_data(id)
    ON DELETE CASCADE
);

COMMIT;
