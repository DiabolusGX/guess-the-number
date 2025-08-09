# Docker Deployment Guide

This guide covers how to deploy the Guess The Number Discord Bot using Docker with support for both remote and local MongoDB/Redis instances.

## Quick Start

### 1. Environment Setup

Copy the environment template and configure your settings:

```bash
cp env.template .env
```

Edit `.env` file with your specific configuration:

```bash
# Required: Discord Bot Token
GTN_BOT_TOKEN=your_discord_bot_token_here

# Database Configuration
GTN_MONGO_HOST=your-mongo-connection-string
GTN_MONGO_USERNAME=your_username
GTN_MONGO_PASSWORD=your_password
GTN_MONGO_DATABASE=guess_the_number

# Redis Configuration  
GTN_REDIS_HOST=your-redis-host
GTN_REDIS_PORT=6379
GTN_REDIS_PASSWORD=your_redis_password
```

### 2. Deployment Options

#### Option A: Remote Databases (Production)

Use this when you have existing MongoDB and Redis instances (e.g., MongoDB Atlas, Redis Cloud):

```bash
# Build and run with remote databases
docker-compose up -d bot
```

#### Option B: Local Databases (Development/Self-hosted)

Use this to spin up MongoDB and Redis containers alongside your bot:

```bash
# Run with local database containers
docker-compose --profile local-db up -d
```

#### Option C: Development with Admin Tools

For development with database management tools:

```bash
# Use development compose with admin tools
docker-compose -f docker-compose.dev.yml --profile admin-tools up -d
```

This includes:
- MongoDB Express (http://localhost:8081) - MongoDB admin interface
- Redis Insight (http://localhost:8082) - Redis admin interface

#### Option D: Production with PM2 Process Manager

For production deployment with advanced process management:

```bash
# Deploy with PM2 process manager
make pm2

# Or with local databases
make pm2-local

# Or with full monitoring stack
make pm2-full
```

PM2 provides:
- Automatic restarts and crash recovery
- Process monitoring and logging
- Memory management and limits
- Graceful reloads with zero downtime
- Built-in load balancing (for future scaling)
- Web-based monitoring dashboard

## Configuration

### Environment Variables

All configuration can be set via environment variables with the `GTN_` prefix:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GTN_BOT_TOKEN` | Discord bot token | - | ✅ |
| `GTN_DEPLOYMENT_MODE` | Deployment mode | `production` | ❌ |
| `GTN_MONGO_HOST` | MongoDB connection string | - | ✅ |
| `GTN_MONGO_USERNAME` | MongoDB username | - | ✅ |
| `GTN_MONGO_PASSWORD` | MongoDB password | - | ✅ |
| `GTN_MONGO_DATABASE` | MongoDB database name | `guess_the_number` | ❌ |
| `GTN_REDIS_HOST` | Redis hostname | `localhost` | ❌ |
| `GTN_REDIS_PORT` | Redis port | `6379` | ❌ |
| `GTN_REDIS_PASSWORD` | Redis password | - | ❌ |
| `GTN_LOGGING_LEVEL` | Log level | `info` | ❌ |
| `GTN_METRICS_ENABLED` | Enable metrics endpoint | `true` | ❌ |

### Docker Compose Profiles

- **default**: Only the bot service
- **local-db**: Bot + MongoDB + Redis containers
- **admin-tools**: Database management interfaces
- **monitoring**: PM2 web dashboard and monitoring tools

## Service Access

| Service | URL | Description |
|---------|-----|-------------|
| Bot Metrics | http://localhost:8080/metrics | Prometheus metrics |
| MongoDB Express | http://localhost:8081 | MongoDB admin (dev only) |
| Redis Insight | http://localhost:8082 | Redis admin (dev only) |
| PM2 Web UI | http://localhost:4200 | PM2 monitoring dashboard |

## Production Deployment

### 1. Using Remote Databases

```bash
# 1. Configure environment
cp env.template .env
# Edit .env with your remote database credentials

# 2. Build and deploy
docker-compose up -d bot

# 3. Check logs
docker-compose logs -f bot
```

### 2. Using Local Databases

```bash
# 1. Configure environment
cp env.template .env
# Set GTN_MONGO_HOST=mongodb://mongo:27017 and GTN_REDIS_HOST=redis

# 2. Deploy with local databases
docker-compose --profile local-db up -d

# 3. Check all services
docker-compose ps
```

### 3. SSL/TLS Configuration

For production MongoDB with SSL:

```bash
# In .env
GTN_MONGO_HOST=mongodb+srv://your-cluster.mongodb.net
GTN_MONGO_PARAMS=retryWrites=true&w=majority&ssl=true
```

## Development Setup

### Local Development with Hot Reload

For development with code changes:

```bash
# 1. Use development compose
docker-compose -f docker-compose.dev.yml up -d

# 2. Access admin tools
# MongoDB Express: http://localhost:8081
# Redis Insight: http://localhost:8082

# 3. View logs with debug level
docker-compose -f docker-compose.dev.yml logs -f bot
```

### Building Custom Images

```bash
# Build specific version
docker build -t guess-the-number:v1.0.0 .

# Build with custom tag
docker build -t your-registry/guess-the-number:latest .

# Push to registry
docker push your-registry/guess-the-number:latest
```

## Monitoring and Health Checks

### Health Check Endpoints

- Bot health: `http://localhost:8080/metrics`
- MongoDB: Available through compose health checks
- Redis: Available through compose health checks

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f bot

# Last 100 lines
docker-compose logs --tail=100 bot
```

### Resource Monitoring

```bash
# Container stats
docker stats

# Specific container
docker stats guess-the-number-bot
```

## Backup and Data Persistence

### MongoDB Backup

```bash
# Create backup
docker exec guess-the-number-mongo mongodump --uri="mongodb://admin:password@localhost:27017/guess_the_number" --out=/backup

# Copy backup from container
docker cp guess-the-number-mongo:/backup ./mongo-backup
```

### Redis Backup

```bash
# Create Redis snapshot
docker exec guess-the-number-redis redis-cli BGSAVE

# Copy RDB file
docker cp guess-the-number-redis:/data/dump.rdb ./redis-backup/
```

### Volume Management

```bash
# List volumes
docker volume ls

# Backup volume
docker run --rm -v guess-the-number_mongo_data:/data -v $(pwd):/backup alpine tar czf /backup/mongo-data.tar.gz -C /data .

# Restore volume
docker run --rm -v guess-the-number_mongo_data:/data -v $(pwd):/backup alpine tar xzf /backup/mongo-data.tar.gz -C /data
```

## Troubleshooting

### Common Issues

1. **Bot can't connect to MongoDB**
   ```bash
   # Check MongoDB container status
   docker-compose ps mongo
   
   # Check MongoDB logs
   docker-compose logs mongo
   
   # Test connection
   docker exec -it guess-the-number-mongo mongosh
   ```

2. **Bot can't connect to Redis**
   ```bash
   # Check Redis container status
   docker-compose ps redis
   
   # Test Redis connection
   docker exec -it guess-the-number-redis redis-cli ping
   ```

3. **Permission issues**
   ```bash
   # Check file permissions
   ls -la logs/
   
   # Fix permissions
   sudo chown -R $USER:$USER logs/
   ```

### Debug Mode

Enable debug logging:

```bash
# In .env
GTN_LOGGING_LEVEL=debug
GTN_DEPLOYMENT_MODE=local

# Restart services
docker-compose restart bot
```

### Container Shell Access

```bash
# Access bot container
docker exec -it guess-the-number-bot sh

# Access MongoDB container
docker exec -it guess-the-number-mongo mongosh

# Access Redis container
docker exec -it guess-the-number-redis redis-cli
```

## Security Considerations

1. **Environment Variables**: Never commit `.env` files with secrets
2. **Network Security**: Use custom networks for service isolation
3. **User Permissions**: Bot runs as non-root user in container
4. **Database Security**: Use strong passwords and enable authentication
5. **SSL/TLS**: Use encrypted connections for remote databases

## PM2 Process Management

### PM2 Deployment

PM2 provides advanced process management features ideal for production environments:

#### Quick Start with PM2

```bash
# Deploy with PM2 process manager
make pm2

# Deploy with local databases and PM2
make pm2-local

# Deploy with monitoring dashboard
make pm2-monitoring

# Full deployment with everything
make pm2-full
```

#### PM2 Features

- **Auto-restart**: Automatically restarts crashed processes
- **Graceful reload**: Zero-downtime deployments
- **Memory monitoring**: Automatic restart on memory limit
- **Log management**: Centralized logging with rotation
- **Process monitoring**: Real-time process statistics
- **Cluster mode**: Built-in load balancing (ready for scaling)

#### PM2 Management Commands

```bash
# Check PM2 process status
make pm2-status

# View PM2 monitoring interface
make pm2-monit

# Restart PM2 processes
make pm2-restart

# Graceful reload (zero downtime)
make pm2-reload

# View PM2 logs
make pm2-logs

# Stop PM2 processes
make pm2-stop

# Flush PM2 logs
make pm2-flush
```

#### PM2 Configuration

The PM2 configuration is defined in `ecosystem.config.js`:

```javascript
module.exports = {
  apps: [{
    name: "guess-the-number-bot",
    script: "./main",
    instances: 1,
    exec_mode: "fork",
    autorestart: true,
    max_memory_restart: "512M",
    // ... more options
  }]
};
```

#### PM2 Web Dashboard

Enable the PM2 web dashboard for visual monitoring:

```bash
# Start with monitoring profile
make pm2-monitoring

# Access dashboard at http://localhost:4200
```

#### Production PM2 Setup

For production with PM2:

```bash
# 1. Configure environment
cp env.template .env
# Edit .env with production settings

# 2. Deploy with PM2
make pm2-build

# 3. Monitor processes
make pm2-status
make pm2-monit
```

### Sharding with PM2

For Discord bot sharding, modify `ecosystem.config.js`:

```javascript
module.exports = {
  apps: [
    {
      name: "bot-shard-0",
      script: "./main",
      env: {
        GTN_BOT_SHARDING_SHARD_IDS: "0,1",
      }
    },
    {
      name: "bot-shard-1", 
      script: "./main",
      env: {
        GTN_BOT_SHARDING_SHARD_IDS: "2,3",
      }
    }
  ]
};
```

## Scaling

### Horizontal Scaling

```bash
# Scale bot instances (with sharding)
docker-compose up -d --scale bot=3

# PM2 cluster mode (load balancing)
# Modify ecosystem.config.js:
# instances: "max" or specific number
```

### Resource Limits

Add to `docker-compose.yml`:

```yaml
services:
  bot:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## Support

For issues and questions:
- Check logs: `docker-compose logs -f`
- Monitor metrics: http://localhost:8080/metrics
- Review configuration in `internal/config/config.yaml`
