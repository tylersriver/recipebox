ALTER TABLE recipes ADD COLUMN share_token TEXT DEFAULT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_recipes_share_token ON recipes(share_token) WHERE share_token IS NOT NULL;
