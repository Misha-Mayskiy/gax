# 📚 GAX API Reference Documentation

Документация REST API, WebSocket событий и моделей данных для проекта GAX.

## 📋 Содержание

1.  [Общая информация](#общая-информация)
2.  [Auth Service (Авторизация)](#-auth-service)
3.  [User Service (Пользователи)](#-user-service)
4.  [Chat Service (Чаты)](#-chat-service)
5.  [Message Service (Сообщения)](#-message-service)
6.  [Media Service (Файлы)](#-media-service)
7.  [Search Service (Поиск)](#-search-service)
8.  [Room Service (Музыкальные комнаты)](#-room-service)
9.  [Call Service (Видеозвонки)](#-call-service)

---

## 🌐 Общая информация

- **Base URL:** `http://localhost:8080` (напрямую к Gateway) или `http://localhost` (через Caddy).
- **Формат данных:** `application/json` (если не указано иное).
- **Кодировка:** UTF-8.

### Аутентификация

Сервис использует JWT (JSON Web Token).

- После успешного входа (`/auth/login`) сервер устанавливает **HttpOnly Cookie** с именем `jwtToken`.
- Все защищенные эндпоинты автоматически проверяют наличие и валидность этой куки.
- При работе через WebSocket токен передается в параметрах запроса: `?token=...`.

### Коды ответов

- `200 OK` — Успешный синхронный запрос.
- `201 Created` — Ресурс успешно создан.
- `400 Bad Request` — Ошибка валидации данных или неверный формат JSON.
- `401 Unauthorized` — Токен отсутствует или невалиден.
- `403 Forbidden` — Нет прав на выполнение операции.
- `404 Not Found` — Ресурс не найден.
- `500 Internal Server Error` — Внутренняя ошибка сервера.

---

## 🔐 Auth Service

Сервис отвечает за регистрацию, вход и безопасность учетных записей.

### 1. Регистрация

Создает нового пользователя в системе.

- **Endpoint:** `POST /auth/register`
- **Body:**
  ```json
  {
    "email": "user@example.com",
    "username": "AlexDoe",
    "password": "strongPassword123"
  }
  ```
- **Response (200):**
  ```json
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Registration successful"
  }
  ```

### 2. Вход в систему

Аутентифицирует пользователя и устанавливает сессионную куку.

- **Endpoint:** `POST /auth/login`
- **Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "strongPassword123"
  }
  ```
- **Response (200):** + `Set-Cookie: jwtToken=...`
  ```json
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "success": true,
    "message": "Login successful"
  }
  ```

### 3. Смена пароля

- **Endpoint:** `POST /auth/password_change`
- **Body:**
  ```json
  {
    "uuid": "user-uuid",
    "old_password": "currentPassword",
    "new_password": "newStrongPassword"
  }
  ```
- **Response (200):**
  ```json
  {
    "success": true,
    "message": "Password changed successfully"
  }
  ```

---

## 👤 User Service

Управление профилем пользователя, друзьями и статусами.

### 1. Создание профиля

Вызывается автоматически после регистрации или вручную для заполнения данных.

- **Endpoint:** `PUT /user/create`
- **Body:**
  ```json
  {
    "uuid": "uuid-from-auth",
    "email": "user@example.com",
    "user_name": "AlexDoe",
    "avatar": "http://minio-host/bucket/avatar.jpg",
    "about_me": "Software Engineer",
    "friends": ["friend-uuid-1", "friend-uuid-2"]
  }
  ```
- **Response (201):** JSON объект с данными пользователя.

### 2. Получение информации

Получает публичный профиль пользователя.

- **Endpoint:** `GET /user/get`
- **Query Params:**
  - `uuid`: ID пользователя.
- **Response (200):**
  ```json
  {
    "uuid": "550e8400-...",
    "email": "user@example.com",
    "user_name": "AlexDoe",
    "avatar": "...",
    "about_me": "...",
    "status": "online",
    "friends": [],
    "created_at": "2023-12-01T12:00:00Z",
    "updated_at": "2023-12-01T12:00:00Z"
  }
  ```

### 3. Обновление профиля

- **Endpoint:** `PATCH /user/update`
- **Body:** (все поля опциональны, кроме uuid)
  ```json
  {
    "uuid": "user-uuid",
    "user_name": "New Name",
    "about_me": "New Bio",
    "avatar": "new-avatar-url"
  }
  ```

### 4. Удаление профиля

- **Endpoint:** `DELETE /user/delete`
- **Body:**
  ```json
  {
    "uuid": "user-uuid"
  }
  ```

### 5. Установка онлайн-статуса

- **Endpoint:** `POST /user/set_online`
- **Body:**
  ```json
  {
    "uuid": "user-uuid",
    "ttl_seconds": 300
  }
  ```

### 6. Проверка статуса

- **Endpoint:** `GET /user/is_online`
- **Query Params:** `uuid`
- **Response:** `{"uuid": "...", "online": true}`

### 7. Список всех онлайн пользователей

- **Endpoint:** `GET /user/get_online_users`
- **Response:** `{"uuids": ["id1", "id2", ...]}`

---

## 💬 Chat Service

Управление чатами (комнатами).

### 1. Создать личный чат (Direct)

- **Endpoint:** `POST /chat/create-direct`
- **Body:**
  ```json
  {
    "user_id": "my-uuid",
    "peer_id": "friend-uuid"
  }
  ```
- **Response (200):** Объект `Chat`.

### 2. Создать групповой чат

- **Endpoint:** `POST /chat/create-group`
- **Body:**
  ```json
  {
    "user_id": "creator-uuid",
    "title": "Project Team",
    "member_ids": ["uuid1", "uuid2", "uuid3"]
  }
  ```

### 3. Обновить группу

Добавление/удаление участников, смена названия.

- **Endpoint:** `PATCH /chat/update-group`
- **Body:**
  ```json
  {
    "chat_id": "chat-uuid",
    "title": "New Title",
    "add_member_ids": ["new-user-uuid"],
    "remove_member_ids": ["kicked-user-uuid"],
    "requester_id": "admin-uuid"
  }
  ```

### 4. Получить информацию о чате

- **Endpoint:** `GET /chat/get`
- **Query Params:** `chat_id`

### 5. Список чатов пользователя

- **Endpoint:** `GET /chat/list`
- **Query Params:**
  - `user_id`: ID пользователя.
  - `limit`: (int) Лимит (по умолчанию 50).
  - `cursor`: (string) Курсор пагинации.

---

## 💌 Message Service

Работа с сообщениями внутри чатов.

### 1. Отправить сообщение

- **Endpoint:** `POST /chat/message/send`
- **Body:**
  ```json
  {
    "chat_id": "chat-uuid",
    "author_id": "user-uuid",
    "text": "Hello world!",
    "media": [
      {
        "id": "file-id",
        "type": "image",
        "url": "http://..."
      }
    ]
  }
  ```
- **Response:** Объект `Message`.

### 2. Список сообщений (История)

- **Endpoint:** `GET /chat/messages/list`
- **Query Params:**
  - `chat_id`: ID чата.
  - `limit`: Количество сообщений.
  - `cursor`: ID последнего загруженного сообщения (для подгрузки старых).

### 3. Редактировать сообщение

- **Endpoint:** `PATCH /chat/message/update`
- **Body:**
  ```json
  {
    "message_id": "msg-uuid",
    "author_id": "user-uuid",
    "text": "Updated text"
  }
  ```

### 4. Удалить сообщение

- **Endpoint:** `DELETE /chat/message/delete`
- **Body:**
  ```json
  {
    "message_ids": ["msg-1", "msg-2"],
    "requester_id": "user-uuid",
    "hard_delete": false
  }
  ```
  - `hard_delete`: `true` (стереть из БД), `false` (пометить как удаленное).

### 5. Дополнительные функции

- **Отметить прочитанным:** `POST /chat/message/mark-read`
- **В избранное:** `POST /chat/message/toggle-saved`
- **Список избранных:** `GET /chat/messages/saved`
- **Список прочитанных:** `GET /chat/messages/read`

---

## 📁 Media Service

Хранение файлов (S3 MinIO).

### 1. Загрузка файла

- **Endpoint:** `POST /media/upload`
- **Headers:** `Content-Type: multipart/form-data`
- **Form Data:**
  - `file`: (Binary) Файл.
  - `user_id`: ID пользователя.
  - `chat_id`: (Опционально) ID чата.
- **Response (201):**
  ```json
  {
    "id": "file-uuid",
    "filename": "cat.jpg",
    "bucket": "files",
    "content_type": "image/jpeg",
    "size": 102450
  }
  ```

### 2. Скачивание файла

- **Endpoint:** `GET /media/download`
- **Query Params:** `id` (UUID файла).
- **Response:** Бинарный поток файла с заголовком `Content-Disposition`.

### 3. Метаданные и Управление

- **Инфо о файле:** `GET /media/meta?id=...`
- **Удалить файл:** `DELETE /media/delete?id=...&user_id=...`
- **Список файлов пользователя:** `GET /media/list?user_id=...&limit=50`

---

## 🔍 Search Service

Глобальный поиск по системе (Elasticsearch).

### 1. Поиск

- **Endpoint:** `GET /search`
- **Query Params:**
  - `q`: Текст запроса (например: "отчет").
  - `type`: Фильтр (`user`, `chat`, `message`, `file`). Если пусто — ищет везде.
  - `limit`: Количество результатов.
  - `offset`: Смещение.
  - `highlight`: `true` для подсветки найденных слов тегами `<b>`.
- **Response:**
  ```json
  {
    "items": [
      {
        "id": "obj-id",
        "type": "message",
        "title": "Финальный отчет",
        "snippet": "Сдаю финальный <b>отчет</b>..."
      }
    ],
    "limit": 20,
    "offset": 0
  }
  ```

---

## 🎵 Room Service

Сервис для синхронного просмотра/прослушивания контента.

### 1. REST API (Управление состоянием)

- **Создать комнату:** `POST /room/create`
  - Body: `{"host_id": "...", "track_url": "https://..."}`
- **Войти в комнату:** `POST /room/join`
  - Body: `{"room_id": "...", "user_id": "..."}`
- **Изменить состояние (Play/Pause):** `POST /room/playback`
  - Body: `{"room_id": "...", "action": "play", "timestamp": 12345}`
- **Получить состояние:** `GET /room/state?room_id=...`

### 2. WebSocket API (Real-time)

- **URL:** `ws://localhost/room/ws/`
- **Query Params:** `room_id=...`, `user_id=...`, `token=...`

**Сообщения от Клиента:**

- `control`: Управление плеером.
  ```json
  { "type": "control", "action": "seek", "position": 45.5 }
  ```

**Сообщения от Сервера:**

- `user_joined`: Кто-то вошел.
- `room_info`: Список участников.
- `play` / `pause` / `seek`: Команда плееру.

---

## 📞 Call Service

SFU сервер для WebRTC звонков.

### WebSocket Signaling

- **URL:** `ws://localhost:8086/ws`
- **Query Params:** `user_id=...` (или через хедер `X-User-ID` от Gateway).

**Протокол обмена (JSON):**

1.  **Вход в звонок:**
    ```json
    {
      "type": "join",
      "payload": { "room_id": "call-room-1" }
    }
    ```
2.  **WebRTC SDP Offer (от Сервера или Клиента):**
    ```json
    {
      "type": "offer",
      "payload": { "type": "offer", "sdp": "..." }
    }
    ```
3.  **WebRTC SDP Answer:**
    ```json
    {
      "type": "answer",
      "payload": { "type": "answer", "sdp": "..." }
    }
    ```
4.  **ICE Candidate:**
    ```json
    {
      "type": "candidate",
      "payload": { "candidate": "...", "sdpMid": "...", ... }
    }
    ```
