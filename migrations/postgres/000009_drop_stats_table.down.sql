CREATE TABLE stats (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    link_id    BIGINT REFERENCES links (id),
    clicks     INTEGER DEFAULT 0,
    date       DATE
);
CREATE UNIQUE INDEX idx_stats_link_date ON stats (link_id, date);
CREATE INDEX idx_stats_deleted_at ON stats (deleted_at);
