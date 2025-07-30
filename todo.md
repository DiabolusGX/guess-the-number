# Feature Implementation Roadmap

## Phase 1: Data Architecture Enhancement

### 1.1 Redis-based Game State Management

**Priority: High | Est: 3-4 days**

**Current State**: Games stored in memory map `runningGames` in GameService
**Target**: Migrate to Redis for persistence and scalability

**Implementation Plan**:

- Create Redis data structures for game state:

  ```bash
  game:channel:{channelID} -> Game JSON object
  game:guild:{guildID}:running -> Set of active channel IDs
  game:stats:{channelID}:guesses -> List of all guesses with metadata
  game:stats:{channelID}:closest -> Sorted set of closest guesses
  game:stats:{channelID}:repeat -> Hash of guess->count
  ```

- Update GameService to use Redis instead of in-memory map
- Add Redis TTL for automatic cleanup of finished games
- Ensure atomic operations for game state updates

### 1.2 MongoDB Sync Service

**Priority: High | Est: 2-3 days**
**Implementation Plan**:

- Create `internal/service/sync_service.go` with scheduled sync functionality
- Implement background goroutine that runs every 5-10 minutes (configurable)
- Sync process:
  1. Fetch active games from Redis
  2. Batch update game statistics in MongoDB
  3. Update user statistics and leaderboards
  4. Clean up completed game data from Redis
- Add configuration options:

  ```yaml
  sync:
    enabled: true
    interval: "5m"
    batch_size: 100
  ```

## Phase 2: Enhanced Statistics & Tracking

### 2.1 Extended Game Statistics

**Priority: Medium | Est: 2-3 days**
**Domain Changes**:
```go
type Game struct {
    // ... existing fields ...
    
    // New statistical fields
    AllGuesses      []GuessAttempt `bson:"allGuesses"`
    ClosestGuesses  []GuessAttempt `bson:"closestGuesses"`  // Top 5 closest
    RepeatGuesses   map[int64]int  `bson:"repeatGuesses"`   // number -> count
    GuessDistribution map[int64]int `bson:"guessDistribution"` // range buckets
}

type GuessAttempt struct {
    UserID    string    `bson:"userID"`
    Guess     int64     `bson:"guess"`
    Distance  int64     `bson:"distance"`  // abs(guess - answer)
    Timestamp time.Time `bson:"timestamp"`
}
```

**Implementation Plan**:

- Update game creation to initialize new stats fields
- Modify `HandleAttempt` to track detailed guess information
- Implement real-time closest guess tracking using Redis sorted sets
- Add repeat guess detection and counting

### 2.2 New API Endpoints for Stats

**Priority: Medium | Est: 1-2 days**

**New Commands**:

- `/game-stats` - Show detailed statistics for current/last game
- `/top-guesses` - Show closest guesses leaderboard
- `/guess-patterns` - Show repeat guess analysis

## Phase 3: Auto-Features Implementation

### 3.1 Auto-Hint System  

**Priority: High | Est: 2-3 days**

**Current State**: Manual hints via commands only
**Target**: Automatic emoji reactions on every guess

**Implementation Plan**:

- Extend `HandleAttemptResponse` to include hint direction
- Modify `internal/bot/events/message/handle_attempt.go`:

  ```go
  // After processing guess, add reaction
  if !response.Correct {
      if response.Game.Answer > request.Guess {
          event.Client().Rest().AddReaction(channelID, messageID, "⬆️")
      } else {
          event.Client().Rest().AddReaction(channelID, messageID, "⬇️")
      }
  }
  ```

- Add guild configuration option:

  ```go
  type GuildConfig struct {
      // ... existing fields ...
      AutoHints bool `bson:"autoHints"`
  }
  ```

- Create `/toggle-auto-hints` command for admins

### 3.2 Auto-Restart System

**Priority: Medium | Est: 3-4 days**

**Configuration Extension**:

```go
type GuildConfig struct {
    // ... existing fields ...
    AutoRestart      bool          `bson:"autoRestart"`
    RestartDelay     time.Duration `bson:"restartDelay"`     // e.g., 30s, 2m
    RestartRange     GameRange     `bson:"restartRange"`     // same range or custom
    MaxAutoRestarts  int           `bson:"maxAutoRestarts"`  // prevent infinite loops
}

type GameRange struct {
    LowerBound int64 `bson:"lowerBound"`
    UpperBound int64 `bson:"upperBound"`
}
```

**Implementation Plan**:

- Create auto-restart scheduler service
- Modify game completion flow in `handle_attempt.go`:

  ```go
  // After game completion processing
  if guildConfig.AutoRestart {
      go scheduleGameRestart(ctx, guildConfig, event.ChannelID)
  }
  ```

- Add configuration commands: `/auto-restart enable|disable [delay] [range]`
- Implement restart queue to handle multiple channels

## Phase 4: Administration Features

### 4.1 Score Redemption & Edit System

**Priority: Medium | Est: 3-4 days**

**New Domain Objects**:

```go
type ScoreTransaction struct {
    ID          string    `bson:"_id"`
    GuildID     string    `bson:"guildID"`
    UserID      string    `bson:"userID"`
    AdminID     string    `bson:"adminID"`
    Type        TxnType   `bson:"type"`        // "redeem", "edit", "bonus"
    Amount      int64     `bson:"amount"`
    Reason      string    `bson:"reason"`
    Timestamp   time.Time `bson:"timestamp"`
}

type TxnType string
const (
    TxnTypeRedeem TxnType = "redeem"
    TxnTypeEdit   TxnType = "edit"
    TxnTypeBonus  TxnType = "bonus"
)
```

**New Commands**:

- `/redeem-points <amount> [reason]` - Allow users to redeem points
- `/edit-score @user <amount> [reason]` - Admin command to modify scores
- `/score-history [@user]` - Show transaction history
- `/leaderboard-reset` - Reset guild leaderboard (admin only)

**Implementation Plan**:

- Create `ScoreTransactionRepository` and service
- Add permission checking for admin commands
- Implement audit trail for all score modifications
- Add configuration for redemption rules per guild

### 4.2 Enhanced Leaderboard System

**Priority: Low | Est: 2 days**

**Features**:

- Monthly/weekly leaderboards
- Category-based leaderboards (fastest win, most games, etc.)
- Leaderboard snapshots and archives
- Export functionality for admins

## Phase 5: Sharding Implementation

### 5.1 Discord Bot Sharding

**Priority: High | Est: 5-7 days**

**Current Challenge**: Single bot instance handling all guilds
**Target**: Multiple shards for better performance and reliability

**Implementation Plan**:

1. **Configuration Updates**:

```yaml
bot:
  sharding:
    enabled: true
    total_shards: 4
    shard_id: 0  # Set per instance
    auto_shard: true  # Auto-detect optimal shard count
```

2. **Shard-Aware Services**:

```go
type ShardConfig struct {
    ShardID     int `json:"shard_id"`
    TotalShards int `json:"total_shards"`
    Enabled     bool `json:"enabled"`
}

// Update bot client initialization
func NewShardedClient(cfg *config.Configuration) bot.Client {
    var sessionConfig *bot.Config
    if cfg.Bot.Sharding.Enabled {
        sessionConfig = &bot.Config{
            ShardCount: cfg.Bot.Sharding.TotalShards,
            ShardIDs:   []int{cfg.Bot.Sharding.ShardID},
        }
    }
    // ... rest of client setup
}
```

3. **Cross-Shard Communication**:

- Implement Redis pub/sub for cross-shard events
- Design shard-aware data partitioning
- Add shard coordination for global leaderboards

4. **Deployment Strategy**:

- Docker containers for each shard
- Load balancer configuration  
- Health checks and auto-scaling

### 5.2 Data Sharding Strategy

**Priority: Medium | Est: 3-4 days**

**Approach**: Shard data by Guild ID for optimal distribution

**Implementation**:

- Guild data stays on guild's primary shard
- Cross-shard queries use Redis aggregation
- Global statistics computed via scheduled jobs
- Implement shard discovery service

## Implementation Timeline

### Week 1-2: Foundation (Phase 1)

- Redis migration for game state
- MongoDB sync service implementation  
- Testing and validation

### Week 3-4: Statistics & Auto-Features (Phases 2-3)

- Enhanced statistics tracking
- Auto-hint system
- Auto-restart functionality

### Week 5-6: Administration (Phase 4)

- Score redemption system
- Admin commands and audit trails
- Enhanced leaderboards

### Week 7-9: Sharding (Phase 5)

- Discord bot sharding
- Cross-shard communication
- Production deployment and testing

## Configuration Migration

### New Config Structure

```yaml
# Add to existing config.yaml
features:
  auto_hints:
    enabled: true
    default_enabled: false  # Per-guild default
  auto_restart:
    enabled: true
    default_delay: "30s"
    max_delay: "10m"
  score_system:
    redemption_enabled: true
    min_redeem_amount: 10
    
sync:
  enabled: true
  interval: "5m"
  batch_size: 100

sharding:
  enabled: false
  total_shards: 1
  shard_id: 0
  auto_shard: true
```

## Risk Mitigation

1. **Data Consistency**: Implement proper Redis transactions and MongoDB write concerns
2. **Performance**: Use Redis pipelining for batch operations
3. **Scalability**: Design with horizontal scaling in mind from start
4. **Monitoring**: Add comprehensive metrics for all new features
5. **Rollback**: Maintain compatibility with current data structures during migration

## Testing Strategy

1. **Unit Tests**: All new services and domain logic
2. **Integration Tests**: Redis-MongoDB sync, cross-shard communication  
3. **Load Testing**: Simulate high-volume game scenarios
4. **Backward Compatibility**: Ensure existing functionality remains intact
5. **Gradual Rollout**: Feature flags for safe production deployment
