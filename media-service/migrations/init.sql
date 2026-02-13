-- Создание таблицы files с новыми полями
CREATE TABLE IF NOT EXISTS files (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    bucket VARCHAR(255) NOT NULL,
    object_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    user_id VARCHAR(255),
    description TEXT,
    chat_id VARCHAR(255)
);

-- Создание индекса для быстрого поиска по user_id
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);

-- Создание индекса для быстрого поиска по chat_id
CREATE INDEX IF NOT EXISTS idx_files_chat_id ON files(chat_id);