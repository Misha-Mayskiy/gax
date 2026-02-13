# Makefile для Search Service

.PHONY: help build run test clean docker-build docker-run docker-stop lint migrate-up migrate-down install-tools

# Переменные окружения
BINARY_NAME=search-service
DOCKER_IMAGE=search-service
DOCKER_TAG=latest
GO_FILES=$(shell find . -name "*.go" -type f)

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  install-tools - Установить необходимые инструменты"
	@echo "  build        - Собрать бинарник"
	@echo "  run          - Запустить приложение локально"
	@echo "  test         - Запустить тесты"
	@echo "  clean        - Очистить временные файлы"
	@echo "  lint         - Проверка кода (форматирование и vet)"
	@echo "  docker-build - Собрать Docker образ"
	@echo "  docker-run   - Запустить в Docker"
	@echo "  docker-stop  - Остановить Docker контейнер"
	@echo "  migrate-up   - Применить миграции"
	@echo "  migrate-down - Откатить миграций"

# Установка инструментов
install-tools:
	@echo "Установка инструментов разработки..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install mvdan.cc/gofumpt@latest
	@echo "Инструменты установлены!"

# Сборка бинарника
build:
	@echo "Сборка бинарника..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/search-service/main.go

# Запуск локально
run:
	@echo "Запуск приложения..."
	POSTGRES_DSN="postgres://postgres:postgres@localhost:5435/users?sslmode=disable" \
	KAFKA_BROKER="localhost:29092" \
	ELASTIC_URL="http://localhost:9200" \
	HTTP_PORT=":8085" \
	LOG_LEVEL="debug" \
	LOG_PRETTY="true" \
	go run ./cmd/search-service/main.go

# Тесты
test:
	@echo "Запуск тестов..."
	go test -v ./... -cover

# Очистка
clean:
	@echo "Очистка временных файлов..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	go clean

# Проверка кода (упрощенная версия)
lint:
	@echo "Проверка форматирования кода..."
	@if command -v gofmt >/dev/null; then \
		echo "Проверка gofmt..."; \
		gofmt_out=$$(gofmt -l $(GO_FILES)); \
		if [ -n "$$gofmt_out" ]; then \
			echo "Файлы требуют форматирования:"; \
			echo "$$gofmt_out"; \
			echo "Для форматирования выполните: make format"; \
			exit 1; \
		else \
			echo "✓ Форматирование в порядке"; \
		fi; \
	else \
		echo "⚠ gofmt не установлен"; \
	fi
	

# Форматирование кода
format:
	@echo "Форматирование кода..."
	@gofmt -w $(GO_FILES)
	@if command -v goimports >/dev/null; then \
		goimports -w $(GO_FILES); \
		echo "✓ Импорты отформатированы"; \
	else \
		echo "⚠ goimports не установлен, установите: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@echo "✓ Код отформатирован"

# Docker сборка
docker-build:
	@echo "Сборка Docker образа..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker запуск
docker-run:
	@echo "Запуск контейнера..."
	docker run -d \
		--name search-service \
		--network host \
		-e POSTGRES_DSN="postgres://postgres:postgres@localhost:5435/users?sslmode=disable" \
		-e KAFKA_BROKER="localhost:29092" \
		-e ELASTIC_URL="http://localhost:9200" \
		-e HTTP_PORT=":8085" \
		-e LOG_LEVEL="info" \
		-p 8085:8085 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker остановка
docker-stop:
	@echo "Остановка контейнера..."
	docker stop search-service || true
	docker rm search-service || true

# Запуск миграций
migrate-up:
	@echo "Применение миграций..."
	# Добавьте команды для миграций, если используете мигратор

# Откат миграций
migrate-down:
	@echo "Откат миграций..."
	# Добавьте команды для отката миграций

# Запуск всей инфраструктуры (docker-compose)
compose-up:
	@echo "Запуск всей инфраструктуры..."
	docker-compose up -d

compose-down:
	@echo "Остановка инфраструктуры..."
	docker-compose down

compose-logs:
	@echo "Логи инфраструктуры..."
	docker-compose logs -f

# Дефолтная цель
default: help