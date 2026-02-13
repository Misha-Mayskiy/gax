build:
	go build ./cmd/server/main.go
run:
	go run ./cmd/server/main.go
docker-start:
	docker-compose up -d
stop:
	docker stop postgres_container
	docker stop server_app
	docker stop redis_container
docker-build:
	docker-compose --env-file ./config/.env up --build -d
test:
	 go test ./... -cover