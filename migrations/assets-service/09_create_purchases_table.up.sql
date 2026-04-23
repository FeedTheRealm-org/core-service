BEGIN

CREATE TABLE IF NOT EXISTS purchases (
  player_id UUID NOT NULL,
  cosmetic_id UUID NOT NULL,
  purchase_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (player_id, cosmetic_id),
  FOREIGN KEY (cosmetic_id) REFERENCES cosmetics(id) ON DELETE CASCADE
);

COMMIT;
