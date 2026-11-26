UPDATE wiki_pages SET tags = '{}' WHERE tags IS NULL;

ALTER TABLE wiki_pages ALTER COLUMN tags SET NOT NULL;