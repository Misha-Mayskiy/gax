build:
	go build ./cmd/server/main.go
run:
	go run ./cmd/server/main.go
docker-start:
	docker-compose up -d

docker-build:
	docker-compose --env-file ./config/.env up --build -d
test:
	 go test ./... -cover