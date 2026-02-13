# Используем официальный образ Go
FROM golang:1.25.5-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник (указываем правильный путь к main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o room-service ./cmd/room-service/main.go

# Финальный минимальный образ
FROM alpine:3.18

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder --chown=appuser:appgroup /app/room-service .

# Переключаемся на непривилегированного пользователя
USER appuser

EXPOSE 8083

CMD ["./room-service"]