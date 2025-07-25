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

## Sharding Configuration

The bot now supports Discord sharding for scaling across multiple guilds. Configure sharding in your `config.yaml`:

```yaml
bot:
  sharding:
    enabled: true        # Set to true to enable sharding
    shard_count: 0       # 0 = auto-detect, or specify number of shards
    auto_scale: true     # Whether to auto-scale shards
    gateway_url: ""      # Custom gateway URL if needed
```

### Sharding Options

- **enabled**: Enable or disable sharding (default: false)
- **shard_count**: Number of shards to use. Set to 0 for auto-detection based on guild count
- **auto_scale**: Allow automatic scaling of shards (default: true)
- **gateway_url**: Custom Discord gateway URL (leave empty for default)

### When to Use Sharding

- Your bot is in **2,500+ guilds** (Discord requires sharding at this point)
- You want to **distribute load** across multiple processes
- You need **better performance** for high-traffic bots
- You're planning for **future growth**

## References

- [disgoorg/disgo](https://github.com/disgoorg/disgo)
- [mongodb/mongo-go-driver](https://github.com/mongodb/mongo-go-driver)
- [flexprice/flexprice](https://github.com/flexprice/flexprice) (DDD inspiration)

---

## 🚧 Migration in Progress: Go Rewrite 🚧

This project is being rewritten in Go using [disgo](https://github.com/disgoorg/disgo) and the official [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver), following Domain Driven Design (DDD) principles. The new structure will use YAML config, structured logging, Sentry, and Prometheus metrics. See the plan in the migration branch or in this README.
