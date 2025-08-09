# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for dependency management
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bot

# PM2 Runtime stage
FROM node:20-alpine

# Install PM2 globally
RUN npm install -g pm2@latest

# Install additional tools for health checks and monitoring
RUN apk add --no-cache ca-certificates tzdata curl wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Create logs and PM2 directories
RUN mkdir -p ~/logs/gtn-bot /app/.pm2 && chown -R appuser:appuser /app ~/logs/gtn-bot

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/internal/config/config.yaml ./internal/config/
COPY --from=builder /app/ecosystem.config.js .

# Change ownership to appuser (must be done after copying files)
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose metrics port
EXPOSE 8080

# Health check using PM2
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD sh -c "pm2 list | grep -q 'online' || exit 1"

# Start PM2 in foreground mode
CMD ["pm2-runtime", "start", "ecosystem.config.js", "--env", "production"]
