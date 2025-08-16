.PHONY: build build-function run test clean deps

# Build the application
build:
	go build -o bin/telegram-bot cmd/bot/main.go

# Build cloud function
build-function:
	go build -o bin/function cmd/function/main.go

# Run the application
run:
	go run cmd/bot/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install development tools
dev-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Развертывание в Yandex Cloud
deploy:
	@chmod +x deploy.sh
	@./deploy.sh

# Локальное тестирование
test-local:
	@echo "🧪 Локальное тестирование..."
	@chmod +x test-local.sh
	@./test-local.sh

# Запуск локального сервера
run-local:
	@echo "🚀 Запуск локального сервера..."
	LOCAL_TEST=true PORT=8080 go run cmd/function/main.go

# Docker команды для Go 1.21
docker-tidy:
	@echo "🐳 Обновление зависимостей в Go 1.21..."
	@docker build -t qweasley .
	@docker run --rm -v $(PWD):/app -w /app qweasley go mod tidy

docker-build-go121:
	@echo "🐳 Сборка в Go 1.21..."
	@docker build -t qweasley .
	@docker run --rm -v $(PWD):/app -w /app qweasley go build -o bin/bot cmd/bot/main.go

docker-test-go121:
	@echo "🐳 Тесты в Go 1.21..."
	@docker build -t qweasley .
	@docker run --rm -v $(PWD):/app -w /app qweasley go test ./...

# Docker
docker-build:
	docker build -t telegram-bot .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build        - Сборка бота"
	@echo "  build-function - Сборка облачной функции"
	@echo "  run          - Запуск бота локально"
	@echo "  test         - Запуск тестов"
	@echo "  clean        - Очистка временных файлов"
	@echo "  deps         - Загрузка зависимостей"
	@echo "  dev-tools    - Установка инструментов разработки"
	@echo "  lint         - Запуск линтера"
	@echo "  fmt          - Форматирование кода"
	@echo "  deploy       - Развертывание в Yandex Cloud"
	@echo "  test-local   - Локальное тестирование бота"
	@echo "  run-local    - Запуск локального сервера"
	@echo "  docker-tidy  - Обновление зависимостей в Go 1.21"
	@echo "  docker-build-go121 - Сборка в Go 1.21"
	@echo "  docker-test-go121  - Тесты в Go 1.21"
	@echo "  docker-build - Сборка Docker образа"
	@echo "  docker-run   - Запуск через Docker Compose"
	@echo "  docker-stop  - Остановка Docker контейнеров"