BEGIN;

DROP INDEX IF EXISTS idx_world_zones_world_online;
DROP INDEX IF EXISTS idx_world_zones_online;

ALTER TABLE world_data
DROP COLUMN IF EXISTS max_active_players,
DROP COLUMN IF EXISTS max_average_player_time;

DROP TABLE IF EXISTS global_player_metrics;

COMMIT;
