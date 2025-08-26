# Makefile для проекта qweasley_go

BINARY_NAME=function
BUILD_DIR=./build
MAIN_FILE=cmd/function/main.go

.PHONY: help build build-docker clean test run dev check deploy up

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

up: ## Запустить контейнеры проекта
	@docker-compose up -d --remove-orphans

dev: ## Запустить в режиме разработки в контейнере
	@docker-compose exec -it go-dev sh -c "LOCAL_TEST=true go run $(MAIN_FILE)"

check: ## Запустить проверку БД в контейнере
	@docker-compose exec -it go-dev sh -c "LOCAL_TEST=true go run scripts/check_db.go"

build: ## Собрать проект
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

build-docker: ## Собрать проект в контейнере
	@docker-compose exec -it go-dev sh -c "go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)"

clean: ## Очистить сборку
	@rm -rf $(BUILD_DIR)

deploy: build ## Развернуть в Яндекс.Облако
	@./deploy.sh