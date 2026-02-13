// Индексы для коллекции messages
db.messages.createIndex({ "id": 1 }, { unique: true });
db.messages.createIndex({ "chat_id": 1, "created_at": -1 });
db.messages.createIndex({ "chat_id": 1, "read_by.user_id": 1 });
db.messages.createIndex({ "saved_by.user_id": 1, "created_at": -1 });
db.messages.createIndex({ "author_id": 1, "created_at": -1 });
db.messages.createIndex({ "deleted": 1 });