# GAX Messenger

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)
![Kafka](https://img.shields.io/badge/Kafka-Event%20Driven-231F20?style=flat&logo=apachekafka)
![WebRTC](https://img.shields.io/badge/WebRTC-Pion-333333?style=flat&logo=webrtc)

GAX — это микросервисный мессенджер с возможностью обмена сообщениями, голосовыми/видео звонками (SFU) и синхронным просмотром контента (Room).

---

## Архитектура

Проект построен на **микросервисной архитектуре**. Все входящие запросы проходят через **Caddy** (Reverse Proxy) и **API Gateway**. Взаимодействие между сервисами осуществляется через gRPC (синхронно) и Kafka (асинхронно).

### Список сервисов

| Сервис             | Описание                                                              | Технологии                        |
| ------------------ | --------------------------------------------------------------------- | --------------------------------- |
| **API Gateway**    | Единая точка входа. Маршрутизация, Auth Middleware, агрегация данных. | REST, gRPC Client                 |
| **Auth Service**   | Регистрация, авторизация, выдача JWT токенов.                         | PostgreSQL, gRPC                  |
| **User Service**   | Профили пользователей, друзья, онлайн-статус.                         | PostgreSQL, Redis, Kafka Producer |
| **Chat Service**   | Личные и групповые чаты, история сообщений.                           | MongoDB, Kafka Producer           |
| **Media Service**  | Загрузка и стриминг файлов (фото, аудио, видео).                      | MinIO (S3), PostgreSQL            |
| **Search Service** | Полнотекстовый поиск по пользователям, чатам и файлам.                | Elasticsearch, Kafka Consumer     |
| **Call Service**   | Сервер видеозвонков (SFU). Управление WebRTC потоками.                | Pion WebRTC, WebSocket, UDP       |
| **Room Service**   | Комнаты для синхронного просмотра/прослушивания контента.             | WebSocket, Redis                  |

### Инфраструктура

- **Базы данных:** PostgreSQL (x4), MongoDB, Redis (x3).
- **Брокер сообщений:** Apache Kafka + Zookeeper.
- **Поиск:** Elasticsearch + Kibana.
- **Файловое хранилище:** MinIO (S3 compatible).
- **Прокси:** Caddy.

---

## Запуск проекта

### Требования

- Docker & Docker Compose
- Минимум 4 GB RAM (рекомендуется 6-8 GB, можно со Swap файлом)
- Лично мы запустили на 1 ядро CPU, 2 GB RAM + 4 GB Swap

### Установка и запуск

1.  **Клонирование репозитория:**

    ```bash
    git clone https://gitlab.crja72.ru/golang/2025/autumn/projects/go22/gax.git
    cd gax
    ```

2.  **Сборка и запуск (Linux/Mac):**
    В проекте предусмотрен скрипт для последовательной сборки (чтобы избежать Out Of Memory):

    ```bash
    chmod +x deploy.sh
    ./deploy.sh
    ```

    _Или классический вариант (если много RAM):_

    ```bash
    docker-compose up --build -d
    ```

3.  **Доступ к сервисам:**
    - **Frontend / API:** http://localhost (80 порт)
    - **MinIO Console:** http://localhost:9001
    - **Kafka UI:** http://localhost:9293

---

## Технические решения

1.  **Асинхронное взаимодействие:**
    События создания сообщений, пользователей и файлов отправляются в Kafka. Search Service вычитывает их и индексирует в Elasticsearch. Это обеспечивает eventual consistency и быструю работу API.

2.  **WebRTC SFU (Call Service):**
    Мы не используем готовые решения. Реализован кастомный SFU на базе `pion/webrtc`. Сервис принимает RTP потоки от участников и маршрутизирует их остальным, не микшируя видео (экономия CPU).

3.  **Синхронизация состояния (Room Service):**
    Используется гибридный подход: WebSocket для мгновенных событий (play/pause) и Redis для хранения "истинного" состояния комнаты. Это позволяет подключаться к комнате в любой момент и получать актуальный таймкод.

4.  **Оптимизация Docker:**
    Используются `alpine` образы для уменьшения размера. Внедрен `.dockerignore`. Настроены `healthcheck` для баз данных, чтобы сервисы не падали при старте.

---

## Тестирование

Проект подключен к GitLab CI/CD.

- При каждом пуше запускаются Unit-тесты для всех сервисов.
- Происходит проверка покрытия кода (Coverage).
- Собираются бинарные файлы.
- Производится деплой (только при изменениях в main)

---

## Команда

Михаил, Егор, Кирилл

Проект сделан благодаря программе Яндекса "Веб-разработка на GO"
