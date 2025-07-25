# Discord Bot Sharding Guide

This guide explains how to use the newly implemented sharding feature in your Discord bot built with [DisGo](https://github.com/disgoorg/disgo).

## What is Sharding?

Discord sharding is a technique used to distribute your bot's workload across multiple processes or connections when your bot grows large. Discord **requires** sharding when your bot is in 2,500+ guilds.

## Features Implemented

✅ **Configuration-based Sharding**: Enable/disable sharding via YAML config  
✅ **Auto-detection**: Automatically determine optimal shard count  
✅ **Manual Shard Count**: Specify exact number of shards if needed  
✅ **Logging Integration**: Track sharding status in bot logs  
✅ **Seamless Integration**: Works with existing bot architecture  

## Configuration

### Enable Sharding

Edit your `config.yaml` file:

```yaml
bot:
  token: "YOUR_BOT_TOKEN"
  sharding:
    enabled: true        # Enable sharding
    shard_count: 0       # 0 = auto-detect, or specify number
    auto_scale: true     # Allow automatic scaling
    gateway_url: ""      # Custom gateway URL (optional)
```

### Configuration Options

| Option       | Description                                    | Default | Required |
|-------------|------------------------------------------------|---------|----------|
| `enabled`   | Enable or disable sharding                    | `false` | No       |
| `shard_count` | Number of shards (0 = auto-detect)          | `0`     | No       |
| `auto_scale` | Allow automatic scaling                       | `true`  | No       |
| `gateway_url` | Custom Discord gateway URL                   | `""`    | No       |

## When to Use Sharding

### **Required Sharding (2,500+ Guilds)**
- Discord **forces** sharding at 2,500+ guilds
- Bot will not work without sharding at this scale

### **Recommended Sharding (1,000+ Guilds)**
- Improved performance and reliability
- Better distribution of events and load
- Preparation for future growth

### **Optional Sharding (<1,000 Guilds)**
- Not necessary but can help with performance
- Good for testing sharding setup

## How It Works

1. **Configuration Loading**: Bot reads sharding config from YAML
2. **Gateway Setup**: DisGo configures gateway with shard parameters
3. **Event Distribution**: Events are distributed across shards
4. **Automatic Management**: DisGo handles shard lifecycle automatically

## Implementation Details

### Code Changes Made

1. **Configuration Structure** (`internal/config/config.go`):
   ```go
   type ShardingConfig struct {
       Enabled      bool   `mapstructure:"enabled" default:"false"`
       ShardCount   int    `mapstructure:"shard_count" default:"0"`
       AutoScale    bool   `mapstructure:"auto_scale" default:"true"`
       GatewayUrl   string `mapstructure:"gateway_url"`
   }
   ```

2. **Client Factory** (`internal/bot/client.go`):
   ```go
   if cfg.Bot.Sharding.Enabled && cfg.Bot.Sharding.ShardCount > 0 {
       gatewayOpts = append(gatewayOpts, gateway.WithShardCount(cfg.Bot.Sharding.ShardCount))
   }
   ```

3. **Logging Integration**: Bot startup logs now include sharding status
   ```
   INFO Guess The Number Bot starting up
      sharding_enabled=true shard_count=4
   ```

## Example Configurations

### Basic Sharding (Auto-detect)
```yaml
bot:
  sharding:
    enabled: true
    shard_count: 0  # Auto-detect based on guild count
```

### Fixed Shard Count
```yaml
bot:
  sharding:
    enabled: true
    shard_count: 4  # Use exactly 4 shards
```

### Disabled Sharding
```yaml
bot:
  sharding:
    enabled: false  # Single connection (default)
```

## Monitoring & Debugging

### Log Messages

When sharding is enabled, you'll see logs like:
```
INFO Guess The Number Bot starting up mode=production sharding_enabled=true shard_count=4
```

### Performance Monitoring

The existing Prometheus metrics will continue to work with sharding:
- `gtn_guilds_total`: Total guilds across all shards
- `gtn_games_running`: Games running across all shards
- Bot latency and performance metrics

## Troubleshooting

### Common Issues

1. **Bot not connecting with sharding enabled**
   - Check if your bot token is valid
   - Verify shard count is appropriate for your guild count
   - Check Discord API status

2. **Events missing or duplicated**
   - Ensure `shard_count` matches across all bot instances
   - Check if multiple bot instances are running

3. **Performance issues**
   - Monitor memory usage per shard
   - Consider adjusting shard count
   - Check network latency to Discord

### Best Practices

1. **Start Small**: Begin with auto-detection (`shard_count: 0`)
2. **Monitor Performance**: Track memory and CPU usage
3. **Gradual Scaling**: Increase shards gradually as you grow
4. **Consistent Configuration**: Use same shard count across deployments

## Migration from Single Client

To migrate from a single client to sharding:

1. **Update Configuration**:
   ```yaml
   bot:
     sharding:
       enabled: true
       shard_count: 0  # Start with auto-detection
   ```

2. **Test in Development**: Verify bot behavior with sharding enabled

3. **Deploy Gradually**: Enable sharding in staging first

4. **Monitor**: Watch logs and metrics after deployment

## Advanced Usage

### Custom Gateway URLs

For enterprise Discord or special setups:
```yaml
bot:
  sharding:
    enabled: true
    gateway_url: "wss://gateway.discord.gg"
```

### Multiple Bot Instances

When running multiple instances of your bot:
- Ensure consistent `shard_count` across all instances
- Use different shard ranges per instance if manual sharding
- Monitor for duplicate events

## Resources

- [Discord Sharding Documentation](https://discord.com/developers/docs/topics/gateway#sharding)
- [DisGo Documentation](https://github.com/disgoorg/disgo)
- [Discord API Reference](https://discord.com/developers/docs/reference)

---

## Summary

Your Discord bot now supports sharding with:
- ✅ Easy configuration via YAML
- ✅ Automatic shard count detection
- ✅ Seamless integration with existing architecture
- ✅ Comprehensive logging and monitoring
- ✅ Production-ready implementation

The implementation is designed to be simple to use while providing the flexibility needed for large-scale Discord bots. 