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

## 🚧 Migration in Progress: Go Rewrite 🚧

This project is being rewritten in Go using [disgo](https://github.com/disgoorg/disgo) and the official [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver), following Domain Driven Design (DDD) principles. The new structure will use YAML config, structured logging, Sentry, and Prometheus metrics. See the plan in the migration branch or in this README.

---

Issue while making `CreateGuild` in guild management service transactional -

- We know for a fact that duplicate keys will exist cuz we don't delete guildData but delete guildConfig when bot leavs guild.
- The core issue: When MongoDB encounters a duplicate key error during a transaction, the transaction is automatically aborted by MongoDB. However, your Go code catches this error and returns nil, making the transaction manager think the operation succeeded.
- Result: The transaction manager tries to commit an already-aborted transaction, causing the "transaction aborted" message without returning a proper error.
