IMAGE_NAME=call-service
PORT=8086
REDIS_ADDR=127.0.0.1:6380
PUBLIC_IP=127.0.0.1 # Изменить при хостинге

redis-up:
	docker run -d --name call-RDS -p 6380:6379 redis:alpine

redis-down:
	docker rm -f call-RDS

run:
	PORT=$(PORT) REDIS_ADDR=$(REDIS_ADDR) go run cmd/main.go

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm \
		--network host \
		--env PORT=$(PORT) \
		--env REDIS_ADDR=$(REDIS_ADDR) \
		--env PUBLIC_IP=$(PUBLIC_IP) \
		$(IMAGE_NAME)

clean:
	go clean
	rm -f call-service