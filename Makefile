.PHONY: help build up down logs clean prod local-db pm2 pm2-build pm2-local pm2-monitoring pm2-full

# Docker configuration
COMPOSE_FILE := docker-compose.yml
PM2_COMPOSE_FILE := docker-compose.pm2.yml

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the Docker image
	docker compose build

build-no-cache: ## Build the Docker image without cache
	docker compose build --no-cache

# Environment setup
env: ## Copy environment template
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Created .env file from template. Please edit it with your configuration."; \
	else \
		echo ".env file already exists. Skipping copy."; \
	fi

# Production deployment
prod: env ## Deploy with remote databases (production)
	docker compose up -d bot

prod-build: env ## Build and deploy with remote databases
	docker compose up -d --build bot

# Local development with local databases
local-db: env ## Deploy with local MongoDB and Redis
	docker compose --profile local-db up -d

local-db-build: env ## Build and deploy with local databases
	docker compose --profile local-db up -d --build

# PM2 deployment targets
pm2: env ## Deploy with PM2 process manager
	docker compose -f $(PM2_COMPOSE_FILE) up -d bot

pm2-build: env ## Build and deploy with PM2
	docker compose -f $(PM2_COMPOSE_FILE) up -d --build bot

pm2-local: env ## Deploy PM2 with local databases
	docker compose -f $(PM2_COMPOSE_FILE) --profile local-db up -d

pm2-monitoring: env ## Deploy PM2 with monitoring dashboard
	docker compose -f $(PM2_COMPOSE_FILE) --profile monitoring up -d

pm2-full: env ## Deploy PM2 with local databases and monitoring
	docker compose -f $(PM2_COMPOSE_FILE) --profile local-db --profile monitoring up -d

# Service management
up: ## Start all services
	docker compose up -d

down: ## Stop all services
	docker compose down

stop: ## Stop services without removing containers
	docker compose stop

restart: ## Restart all services
	docker compose restart

restart-bot: ## Restart only the bot service
	docker compose restart bot

# Logs and monitoring
logs: ## Show logs for all services
	docker compose logs -f

logs-bot: ## Show logs for bot service only
	docker compose logs -f bot

logs-mongo: ## Show logs for MongoDB service
	docker compose logs -f mongo

logs-redis: ## Show logs for Redis service
	docker compose logs -f redis

# PM2 specific log commands
logs-pm2: ## Show PM2 logs
	docker compose -f $(PM2_COMPOSE_FILE) logs -f bot

pm2-status: ## Show PM2 process status
	docker exec -it guess-the-number-bot-pm2 pm2 status

pm2-monit: ## Open PM2 monitoring interface
	docker exec -it guess-the-number-bot-pm2 pm2 monit

pm2-restart: ## Restart PM2 processes
	docker exec -it guess-the-number-bot-pm2 pm2 restart all

pm2-reload: ## Graceful reload PM2 processes
	docker exec -it guess-the-number-bot-pm2 pm2 reload all

pm2-stop: ## Stop PM2 processes
	docker exec -it guess-the-number-bot-pm2 pm2 stop all

pm2-logs: ## Show PM2 application logs
	docker exec -it guess-the-number-bot-pm2 pm2 logs

pm2-flush: ## Flush PM2 logs
	docker exec -it guess-the-number-bot-pm2 pm2 flush

# Database operations
mongo-shell: ## Connect to MongoDB shell
	docker exec -it guess-the-number-mongo mongosh

redis-cli: ## Connect to Redis CLI
	docker exec -it guess-the-number-redis redis-cli

# Backup operations
backup-mongo: ## Backup MongoDB data
	@mkdir -p ./backups
	docker exec guess-the-number-mongo mongodump --uri="mongodb://admin:password@localhost:27017/guess_the_number" --out=/backup
	docker cp guess-the-number-mongo:/backup ./backups/mongo-$(shell date +%Y%m%d_%H%M%S)

backup-redis: ## Backup Redis data
	@mkdir -p ./backups
	docker exec guess-the-number-redis redis-cli BGSAVE
	docker cp guess-the-number-redis:/data/dump.rdb ./backups/redis-$(shell date +%Y%m%d_%H%M%S).rdb

# Maintenance
clean: ## Remove containers, networks, and volumes
	docker compose down -v --remove-orphans
	docker system prune -f

clean-all: ## Remove everything including images
	docker compose down -v --remove-orphans --rmi all
	docker system prune -af

update: ## Pull latest images and restart
	docker compose pull
	docker compose up -d

# Health checks
status: ## Show status of all services
	docker compose ps

health: ## Check health of running containers
	@echo "=== Container Status ==="
	docker compose ps
	@echo "\n=== Container Health ==="
	@docker ps --format "table {{.Names}}\t{{.Status}}" | grep guess-the-number || echo "No running containers"
	@echo "\n=== Bot Metrics (if available) ==="
	@curl -s http://localhost:8080/metrics > /dev/null && echo "✅ Bot metrics endpoint is accessible" || echo "❌ Bot metrics endpoint is not accessible"

# Development helpers
shell: ## Access bot container shell
	docker exec -it guess-the-number-bot sh

shell-dev: ## Access development bot container shell
	docker exec -it guess-the-number-bot-dev sh

# Admin tools URLs
urls: ## Show admin interface URLs
	@echo "=== Service URLs ==="
	@echo "Bot Metrics:      http://localhost:8080/metrics"
	@echo "MongoDB Express:  http://localhost:8081 (dev mode only)"
	@echo "Redis Insight:    http://localhost:8082 (dev mode only)"
	@echo "PM2 Web UI:       http://localhost:4200 (PM2 monitoring profile)"

# Testing
test-connections: ## Test database connections
	@echo "Testing MongoDB connection..."
	@docker exec guess-the-number-mongo mongosh --eval "db.adminCommand('ping')" > /dev/null && echo "✅ MongoDB is accessible" || echo "❌ MongoDB connection failed"
	@echo "Testing Redis connection..."
	@docker exec guess-the-number-redis redis-cli ping > /dev/null && echo "✅ Redis is accessible" || echo "❌ Redis connection failed"
