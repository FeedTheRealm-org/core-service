BEGIN;

-- Make sure item_categories exists
CREATE TABLE IF NOT EXISTS item_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- First, insert distinct category names found in items.category into item_categories
INSERT INTO item_categories (id, name, created_at, updated_at)
SELECT gen_random_uuid(), category, NOW(), NOW()
FROM items
WHERE category IS NOT NULL AND category <> ''
ON CONFLICT (name) DO NOTHING;

-- Insert distinct categories found in item_sprites.category into item_categories
INSERT INTO item_categories (id, name, created_at, updated_at)
SELECT gen_random_uuid(), category, NOW(), NOW()
FROM item_sprites
WHERE category IS NOT NULL AND category <> ''
ON CONFLICT (name) DO NOTHING;

-- Now set items.category_id to matching ids (only if category_id is null)
UPDATE items
SET category_id = ic.id
FROM item_categories ic
WHERE items.category = ic.name
  AND (items.category_id IS NULL OR items.category_id = '00000000-0000-0000-0000-000000000000'::UUID);

-- Set item_sprites.category_id to matching ids
UPDATE item_sprites
SET category_id = ic.id
FROM item_categories ic
WHERE item_sprites.category = ic.name
  AND (item_sprites.category_id IS NULL OR item_sprites.category_id = '00000000-0000-0000-0000-000000000000'::UUID);

COMMIT;
