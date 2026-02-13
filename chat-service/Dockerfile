# Используем официальный образ Go для сборки
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum из текущей папки
COPY go.mod ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o chat-service ./cmd/app

# Финальный образ
FROM alpine:3.23

WORKDIR /app
COPY --from=builder /app/chat-service .

EXPOSE 8081
CMD ["./chat-service"]
