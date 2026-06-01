BEGIN;

ALTER TABLE world_data
ADD COLUMN max_active_players INTEGER NOT NULL DEFAULT 0,
ADD COLUMN max_average_player_time INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS global_player_metrics (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    max_active_players INTEGER NOT NULL DEFAULT 0,
    max_average_player_time INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO global_player_metrics (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_world_zones_world_online
    ON world_zones (world_id, is_online);

CREATE INDEX IF NOT EXISTS idx_world_zones_online
    ON world_zones (is_online);

COMMIT;
