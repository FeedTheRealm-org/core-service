ALTER TABLE world_zones
DROP COLUMN IF EXISTS player_count_updated_at,
DROP COLUMN IF EXISTS active_players;
