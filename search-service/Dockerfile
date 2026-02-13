# Используем официальный образ Go
FROM golang:1.25.4-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -p 1 -v -ldflags="-s -w" -o search-service ./cmd/search-service/main.go

# Финальный образ
FROM alpine:3.18

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Настраиваем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder --chown=appuser:appgroup /app/search-service .

# Переключаемся на непривилегированного пользователя
USER appuser

# Экспортируем порт
EXPOSE 8085

# Команда запуска
CMD ["./search-service"]