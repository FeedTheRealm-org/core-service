BEGIN;

-- Restore NOT NULL to category column on item_sprites
ALTER TABLE item_sprites
    ALTER COLUMN category SET NOT NULL;

COMMIT;
