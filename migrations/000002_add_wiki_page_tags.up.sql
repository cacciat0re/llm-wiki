ALTER TABLE wiki_pages
    ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}'::text[];

CREATE INDEX idx_wiki_pages_tags ON wiki_pages USING GIN (tags);