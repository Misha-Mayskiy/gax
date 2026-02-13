CREATE TABLE IF NOT EXISTS chats (
    id          TEXT PRIMARY KEY,
    kind        TEXT NOT NULL,
    title       TEXT,
    created_by  TEXT NOT NULL,
    created_at  BIGINT NOT NULL
);