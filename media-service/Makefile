BINARY_NAME=media-service
COVERAGE_FILE=coverage.out

run:
	go run cmd/app/main.go

build:
	go build -o ${BINARY_NAME} cmd/app/main.go

test:
	go test -v ./... -cover

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${COVERAGE_FILE}

docker-build:
	docker build -t ${BINARY_NAME} .

docker-up-deps:
	docker-compose up -d minio media-psql kafka

deps:
	go mod download
	go mod tidy

.PHONY: run build test cover clean docker-build