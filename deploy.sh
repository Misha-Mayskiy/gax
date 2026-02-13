#!/bin/bash
set -e # Остановить скрипт, если какая-то команда упадет

echo "🚀 Начинаем последовательную сборку..."

# Собираем легкие сервисы
docker compose build auth-service
docker compose build user-service
docker compose build chat-service
docker compose build media-service
docker compose build room-service
docker compose build call-service
docker compose build api-gateway
docker compose build search-service

echo "✅ Сборка завершена. Запускаем контейнеры..."
docker compose up -d