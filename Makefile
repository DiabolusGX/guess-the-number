.PHONY: help pm2-direct pm2-direct-stop pm2-direct-restart pm2-direct-status pm2-direct-logs pm2-direct-rebuild

# Go build configuration
GO_BINARY := bot
GO_PACKAGE := ./cmd/bot

# Environment setup
env: ## Copy environment template
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Created .env file from template. Please edit it with your configuration."; \
	else \
		echo ".env file already exists. Skipping copy."; \
	fi

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Direct PM2 management (no Docker)
pm2-direct: env ## Build and run Go bot directly with PM2 (no Docker)
	@echo "Building Go binary for Linux..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(GO_BINARY) $(GO_PACKAGE)
	chmod +x $(GO_BINARY)
	@echo "Starting with PM2..."
	pm2 start ./$(GO_BINARY) --name "gtn-go-bot" --restart-delay=10000 --max-memory-restart=6G

pm2-direct-stop: ## Stop direct PM2 process
	pm2 stop gtn-go-bot || true

pm2-direct-restart: ## Restart direct PM2 process
	pm2 restart gtn-go-bot || make pm2-direct

pm2-direct-status: ## Show direct PM2 status
	pm2 status gtn-go-bot

pm2-direct-logs: ## Show direct PM2 logs
	pm2 logs gtn-go-bot

## Stop, rebuild and restart direct PM2
pm2-direct-rebuild: env
	@echo "Building Go binary for Linux..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(GO_BINARY) $(GO_PACKAGE)
	chmod +x $(GO_BINARY)
	@echo "Restarting with PM2..."
	pm2 restart gtn-go-bot || true

health: ## Check health of running containers
	@echo "=== Container Status ==="
	docker compose ps
	@echo "\n=== Container Health ==="
	@docker ps --format "table {{.Names}}\t{{.Status}}" | grep guess-the-number || echo "No running containers"
	@echo "\n=== Bot Metrics (if available) ==="
	@curl -s http://localhost:8080/metrics > /dev/null && echo "✅ Bot metrics endpoint is accessible" || echo "❌ Bot metrics endpoint is not accessible"
