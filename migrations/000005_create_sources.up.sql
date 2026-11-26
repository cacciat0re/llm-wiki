CREATE TABLE sources (
    id      BIGSERIAL PRIMARY KEY,
    origin  VARCHAR(2048) NOT NULL,
    title   VARCHAR(512) NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    CONSTRAINT uk_sources_origin UNIQUE (origin)
);

CREATE TABLE wiki_page_sources (
    wiki_page_id BIGINT NOT NULL,
    source_id    BIGINT NOT NULL,
    CONSTRAINT pk_wiki_page_sources PRIMARY KEY (wiki_page_id, source_id),
    CONSTRAINT fk_wiki_page_sources_page
        FOREIGN KEY (wiki_page_id) REFERENCES wiki_pages (id) ON DELETE CASCADE,
    CONSTRAINT fk_wiki_page_sources_source
        FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE
);

CREATE INDEX idx_wiki_page_sources_source_id ON wiki_page_sources (source_id);