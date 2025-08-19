# Makefile для проекта qweasley_go

BINARY_NAME=function
BUILD_DIR=./build
MAIN_FILE=cmd/function/main.go

.PHONY: help build clean test run dev deploy

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

dev: ## Запустить в режиме разработки
	@LOCAL_TEST=true go run $(MAIN_FILE)

check: ## Запустить в режиме разработки
	@LOCAL_TEST=true go run scripts/check_db.go

deploy: ## Развернуть в Яндекс.Облако
	@./deploy.sh