CREATE TABLE questions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);