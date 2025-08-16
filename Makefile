.PHONY: build dev test clean deps deploy docker help

.DEFAULT_GOAL := help

build: ## 🔨 Build cloud function
	go build -o bin/function cmd/function/main.go

dev: ## 🚀 Run local server for testing
	LOCAL_TEST=true PORT=8080 go run cmd/function/main.go

test: ## 🧪 Run tests
	go test ./...

clean: ## 🧹 Clean build artifacts
	rm -rf bin/ build/

deps: ## 📦 Download dependencies
	go mod download
	go mod tidy

fmt: ## ✨ Format code
	go fmt ./...

deploy: ## ☁️ Deploy to Yandex Cloud
	@chmod +x deploy.sh
	@./deploy.sh

docker: ## 🐳 Update dependencies with Go 1.21
	@docker build -t qweasley .
	@docker run --rm -v $(PWD):/app -w /app qweasley go mod tidy

help: ## 💡 Show this help message
	@echo "🤖 Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'