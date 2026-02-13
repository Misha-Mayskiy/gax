CREATE TABLE IF NOT EXISTS messages (
    id          TEXT PRIMARY KEY,
    chat_id     TEXT NOT NULL,
    author_id   TEXT NOT NULL,
    text        TEXT NOT NULL,
    created_at  BIGINT NOT NULL
);
