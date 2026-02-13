#!/bin/bash

API_URL="http://localhost:8080"
ROOM_ID=""

echo "=== Тестирование Room Service через API Gateway ==="
echo

# 1. Создание комнаты
echo "1. Создание комнаты..."
CREATE_RESPONSE=$(curl -s -X POST "$API_URL/room/create" \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "test_user_'$(date +%s)'",
    "track_url": "https://music.yandex.ru/track/'$(date +%s)'"
  }')

echo "Ответ создания:"
echo "$CREATE_RESPONSE" | jq .

# Извлекаем ID комнаты
ROOM_ID=$(echo "$CREATE_RESPONSE" | jq -r '.room_id')
if [ -z "$ROOM_ID" ] || [ "$ROOM_ID" = "null" ]; then
    echo "❌ Не удалось создать комнату"
    exit 1
fi

echo "✅ Комната создана: $ROOM_ID"
echo

# 2. Получение состояния
echo "2. Получение состояния комнаты..."
curl -s -X GET "$API_URL/room/state?room_id=$ROOM_ID" | jq .
echo

# 3. Присоединение пользователя
echo "3. Присоединение пользователя..."
curl -s -X POST "$API_URL/room/join" \
  -H "Content-Type: application/json" \
  -d "{
    \"room_id\": \"$ROOM_ID\",
    \"user_id\": \"guest_user_$(date +%s)\"
  }" | jq .
echo

# 4. Управление воспроизведением
echo "4. Запуск воспроизведения..."
curl -s -X POST "$API_URL/room/playback" \
  -H "Content-Type: application/json" \
  -d "{
    \"room_id\": \"$ROOM_ID\",
    \"action\": \"play\",
    \"timestamp\": $(date +%s)
  }" | jq .
echo

# 5. Пауза
echo "5. Пауза воспроизведения..."
curl -s -X POST "$API_URL/room/playback" \
  -H "Content-Type: application/json" \
  -d "{
    \"room_id\": \"$ROOM_ID\",
    \"action\": \"pause\",
    \"timestamp\": $(date +%s)
  }" | jq .
echo

echo "=== Тестирование завершено ==="

