-- Добавление новых полей к существующей таблице
ALTER TABLE files 
ADD COLUMN IF NOT EXISTS user_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS chat_id VARCHAR(255);

-- Создание индексов (если их нет)
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_chat_id ON files(chat_id);