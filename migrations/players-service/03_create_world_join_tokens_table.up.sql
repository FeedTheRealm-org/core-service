BEGIN;

CREATE TABLE IF NOT EXISTS world_join_tokens (
    token_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    world_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_world_join_tokens_character_info
        FOREIGN KEY (user_id) REFERENCES character_infos(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_world_join_tokens_user_id ON world_join_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_world_join_tokens_world_id ON world_join_tokens(world_id);
CREATE INDEX IF NOT EXISTS idx_world_join_tokens_expires_at ON world_join_tokens(expires_at);

COMMIT;
