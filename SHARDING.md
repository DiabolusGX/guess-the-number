# Discord Bot Sharding Implementation

This document explains how sharding has been implemented in your Discord bot using the [disgoorg/disgo](https://github.com/disgoorg/disgo) library.

## What is Sharding?

Discord sharding is a way to scale Discord bots across multiple processes or servers when they grow beyond 2,500 guilds. Each shard handles a subset of guilds, allowing for:

- **Better Performance**: Distribute load across multiple processes
- **Improved Reliability**: If one shard fails, others continue working
- **Scalability**: Handle more guilds by adding more shards
- **Resource Management**: Better memory and CPU utilization

## Features Implemented

✅ **Automatic Shard Detection**: Bot can auto-detect optimal shard count based on guild count  
✅ **Manual Shard Configuration**: Set specific shard count and IDs  
✅ **Auto-Scaling**: Automatically adjust sharding based on growth  
✅ **Shard-Aware Logging**: Enhanced logging with shard information  
✅ **Graceful Fallback**: Works with or without sharding enabled  
✅ **Event Monitoring**: Special events for shard readiness tracking  

## Configuration

### Basic Configuration

The sharding configuration is added to your `config.yaml`:

```yaml
bot:
  token: "YOUR_BOT_TOKEN"
  sharding:
    enabled: false      # Set to true to enable sharding
    shard_count: 0      # 0 = auto-detect optimal shard count
    shard_ids: []       # empty = all shards, or specify like [0, 1]
    auto_scaling: true  # Enable automatic scaling
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable/disable sharding |
| `shard_count` | integer | `0` | Total number of shards (0 = auto-detect) |
| `shard_ids` | array | `[]` | Specific shard IDs to run (empty = all shards) |
| `auto_scaling` | boolean | `true` | Enable automatic scaling based on guild count |

## Usage Examples

### Example 1: Auto-Detection (Recommended)

Let Discord determine the optimal shard count:

```yaml
bot:
  sharding:
    enabled: true
    shard_count: 0      # Auto-detect
    shard_ids: []       # Run all shards
    auto_scaling: true
```

### Example 2: Manual Sharding

Set specific shard configuration:

```yaml
bot:
  sharding:
    enabled: true
    shard_count: 4      # Use 4 shards
    shard_ids: []       # Run all 4 shards
    auto_scaling: false
```

### Example 3: Specific Shard Instance

Run only specific shards (useful for distributed deployment):

```yaml
bot:
  sharding:
    enabled: true
    shard_count: 4      # Total of 4 shards
    shard_ids: [0, 1]   # This instance runs shards 0 and 1
    auto_scaling: true
```

## Deployment Strategies

### Single Server Deployment

Run all shards on one server:

```yaml
bot:
  sharding:
    enabled: true
    shard_count: 2
    shard_ids: []       # Run all shards
```

### Multi-Server Deployment

Distribute shards across multiple servers:

**Server 1:**
```yaml
bot:
  sharding:
    enabled: true
    shard_count: 4
    shard_ids: [0, 1]   # Run shards 0 and 1
```

**Server 2:**
```yaml
bot:
  sharding:
    enabled: true
    shard_count: 4
    shard_ids: [2, 3]   # Run shards 2 and 3
```

### Docker Deployment

Use environment variables for dynamic configuration:

```bash
# Run specific shards in containers
docker run -e GTN_BOT_SHARDING_ENABLED=true \
           -e GTN_BOT_SHARDING_SHARD_COUNT=4 \
           -e GTN_BOT_SHARDING_SHARD_IDS="[0,1]" \
           your-bot-image
```

## Monitoring and Logging

The implementation includes enhanced logging for shard monitoring:

### Shard Events

- **Ready Event**: Logs when individual shards are ready
- **Guild Ready**: Logs when guilds become ready on specific shards
- **Guilds Ready**: Logs when all guilds are ready for a shard

### Log Examples

```log
[INFO] Sharding enabled shard_count=2 shard_ids=[0,1] auto_scaling=true
[INFO] Starting bot with sharding enabled
[INFO] Bot shard is ready shard_id=0 user_id=123456789 username=YourBot
[INFO] Bot shard is ready shard_id=1 user_id=123456789 username=YourBot
[DEBUG] Guild ready guild_id=987654321 shard_id=0
[INFO] All guilds ready for shard shard_id=0
```

## Implementation Details

### Code Structure

The sharding implementation spans several files:

- `internal/config/config.go`: Configuration structure
- `internal/bot/client.go`: Bot client with sharding support
- `internal/bot/events/misc/ready.go`: Enhanced ready events
- `internal/bot/events/misc/shard_ready.go`: Shard-specific events

### Key Components

1. **Configuration**: `ShardingConfig` struct with all sharding options
2. **Client Setup**: Conditional setup based on sharding configuration
3. **Connection**: Uses `OpenShardManager()` vs `OpenGateway()`
4. **Events**: Shard-aware event logging and monitoring

### Disgo Integration

The implementation uses disgo's built-in sharding features:

```go
// Sharding configuration
shardOpts = append(shardOpts, sharding.WithShardCount(cfg.Bot.Sharding.ShardCount))
shardOpts = append(shardOpts, sharding.WithShardIDs(cfg.Bot.Sharding.ShardIDs...))
shardOpts = append(shardOpts, sharding.WithAutoScaling(true))

// Apply to client
clientOpts = append(clientOpts, bot.WithShardManagerConfigOpts(shardOpts...))
```

## When to Use Sharding

### Enable Sharding When:

- Bot is in 2,000+ guilds (approaching Discord's 2,500 limit)
- Experiencing performance issues with single instance
- Need better resource distribution
- Planning for growth beyond 2,500 guilds

### Sharding Guidelines:

- **Start Simple**: Use auto-detection first
- **Monitor Performance**: Watch memory and CPU usage
- **Scale Gradually**: Increase shards as needed
- **Test Thoroughly**: Test shard configuration before production

## Testing

### Local Testing

1. Copy the example configuration:
   ```bash
   cp internal/config/config-sharded.yaml internal/config/config.yaml
   ```

2. Enable sharding and set your token:
   ```yaml
   bot:
     token: "YOUR_BOT_TOKEN"
     sharding:
       enabled: true
       shard_count: 2
   ```

3. Run the bot:
   ```bash
   go run cmd/bot/main.go
   ```

4. Monitor the logs for shard-related messages

### Production Testing

1. Start with a small shard count (2-4)
2. Monitor performance and resource usage
3. Gradually increase as your bot grows
4. Test failover scenarios

## Troubleshooting

### Common Issues

**Build Errors**: Ensure all dependencies are up to date:
```bash
go mod tidy
```

**Connection Issues**: Check your bot token and permissions

**Shard Mismatch**: Ensure all instances use the same `shard_count`

**Resource Issues**: Monitor memory usage with multiple shards

### Debug Mode

Enable debug logging to see detailed shard information:

```yaml
logging:
  level: "debug"
```

## Best Practices

1. **Use Auto-Detection**: Let Discord determine optimal shard count
2. **Monitor Resources**: Watch memory and CPU usage
3. **Gradual Scaling**: Don't jump from 1 to 16 shards immediately
4. **Consistent Configuration**: All instances should use same `shard_count`
5. **Load Balancing**: Distribute shards evenly across servers
6. **Health Monitoring**: Implement health checks for each shard

## Future Enhancements

Potential improvements for the sharding implementation:

- **Cross-Shard Communication**: Redis pub/sub for shard coordination
- **Dynamic Resharding**: Automatic shard rebalancing
- **Health Monitoring**: Advanced shard health tracking
- **Metrics**: Prometheus metrics for shard performance
- **Auto-Recovery**: Automatic shard restart on failures

## References

- [Discord Sharding Documentation](https://discord.com/developers/docs/topics/gateway#sharding)
- [Disgo Sharding Example](https://github.com/disgoorg/disgo/blob/master/_examples/sharding/example.go)
- [Discord API Documentation](https://discord.com/developers/docs/)