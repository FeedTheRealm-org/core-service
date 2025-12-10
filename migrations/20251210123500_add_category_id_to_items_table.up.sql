BEGIN;

ALTER TABLE items ADD COLUMN IF NOT EXISTS category_id UUID;

-- Add foreign key constraint only if not present, using a plpgsql DO block
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_items_category'
    ) THEN
        ALTER TABLE items
            ADD CONSTRAINT fk_items_category FOREIGN KEY (category_id) REFERENCES item_categories(id) ON DELETE RESTRICT;
    END IF;
END;
$$;

COMMIT;
