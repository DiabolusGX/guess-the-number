# Guess The Number Go Bot

A Discord bot rewritten in Go using [disgo](https://github.com/disgoorg/disgo), [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver), and Domain Driven Design (DDD) principles. Includes YAML config, structured logging, Sentry, and Prometheus metrics.

## Project Structure

- `cmd/bot/`: Main entry point
- `internal/`: Business logic, domain, repository, service, bot
- `pkg/`: External/generic dependencies (logging, sentry, metrics)
- `config.yaml`: Configuration file

## Setup

1. Copy `config.yaml` and fill in your settings.
2. Run `go mod tidy` to install dependencies.
3. Run the bot: `go run ./cmd/bot`

---

## Build and Run

`pm2-direct` make commands are recommended.
that's it for now
