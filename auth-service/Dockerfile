FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app
COPY go.sum go.mod ./

RUN go mod tidy
RUN go mod download

COPY ./. .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s"\
    -o /app/bin/server \
    ./cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s"\
    -o /app/bin/migrate \
    ./cmd/migrate/main.go

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate

COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

RUN chown -R appuser:appuser /app
USER appuser

FROM runtime AS migrate
ENTRYPOINT ["/app/migrate"]

FROM runtime AS server
EXPOSE 8081
CMD ["/app/server"]
