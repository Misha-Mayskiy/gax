FROM golang:alpine

WORKDIR /app

# Копируем go.mod и go.sum из текущей папки
COPY go.mod ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o api_gateway ./cmd/app

EXPOSE 8080
CMD ["./api_gateway"]
