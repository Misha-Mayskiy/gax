# 1: Сборка (Builder)
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o call-service ./cmd/main.go

# 2: Запуск образа из бинарника
FROM alpine:latest

WORKDIR /root/

# Устанавливаем сертификаты
RUN apk --no-cache add ca-certificates

# Копируем только бинарник
COPY --from=builder /app/call-service .

EXPOSE 8086
EXPOSE 50000-60000/udp

CMD ["./call-service"]