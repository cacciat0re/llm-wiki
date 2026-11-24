CREATE TABLE wiki_pages (
    id         BIGSERIAL PRIMARY KEY,
    slug       VARCHAR(512) NOT NULL,
    title      VARCHAR(512) NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID,
    CONSTRAINT uk_page_slug UNIQUE (slug)
);

CREATE INDEX idx_wiki_pages_updated_at ON wiki_pages (updated_at DESC);