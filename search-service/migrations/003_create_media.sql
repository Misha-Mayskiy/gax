CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    bucket VARCHAR(50) NOT NULL,
    object_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100),
    size BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);