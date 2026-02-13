# Этап 1: сборка
FROM golang:1.25.4 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd/app/main.go

# Этап 2: финальный образ
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/user-service .
EXPOSE 50051
CMD ["/app/user-service"]
