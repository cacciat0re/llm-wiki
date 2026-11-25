DROP INDEX IF EXISTS idx_wiki_pages_tags;

ALTER TABLE wiki_pages DROP COLUMN IF EXISTS tags;