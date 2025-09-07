# Administration Features

## Score Redemption & Edit System

**Priority:** Medium | Est: 3-4 days

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
- `/points-reset` - Reset guild points leaderboard (admin only)

**Implementation Plan**:

- Create `ScoreTransactionRepository` and service
- Add permission checking for admin commands
- Implement audit trail for all score modifications
- Add configuration for redemption rules per guild

## Dockerize

- one command setup with docker db or remote db options (for both mongo & redis)
- populate sample data in db for testing
- cluster mode for transaction support
- make commands for same
