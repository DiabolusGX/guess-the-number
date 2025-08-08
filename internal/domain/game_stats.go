package domain

import (
	"context"
	"time"
)

// GuessAttempt represents an individual guess made by a user
type GuessAttempt struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	GameID    string    `bson:"gameId" json:"gameId"`
	GuildID   string    `bson:"guildId" json:"guildId"`
	ChannelID string    `bson:"channelId" json:"channelId"`
	UserID    string    `bson:"userId" json:"userId"`
	Guess     int64     `bson:"guess" json:"guess"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	IsCorrect bool      `bson:"isCorrect,omitempty" json:"isCorrect,omitempty"`
	Distance  int64     `bson:"distance" json:"distance"` // abs(guess - answer)
}

// GameStatsRepository defines the interface for managing guess attempts
type GameStatsRepository interface {
	BulkInsert(ctx context.Context, attempts []*GuessAttempt) error
	GetByGame(ctx context.Context, gameID string) ([]*GuessAttempt, error)
	GetByGuildAndTimeRange(ctx context.Context, guildID string, timeRange TimeRange) ([]*GuessAttempt, error)
	GetTopGuessedNumbers(ctx context.Context, guildID, gameID string, timeRange TimeRange, limit int) ([]*NumberFrequencyStats, error)
	GetTopGuessers(ctx context.Context, guildID, gameID string, timeRange TimeRange, limit int) ([]*UserGuessStats, error)
	GetClosestGuesses(ctx context.Context, gameID, guildID string, limit int) ([]*GuessAttempt, error)
}

// RedisStatsRepository defines the interface for Redis-based stats operations
type RedisStatsRepository interface {
	RecordGuessAsync(ctx context.Context, game *Game, userID string, guess, distance, timestamp int64) error
	GetGameGuesses(ctx context.Context, game *Game) ([]*GuessAttempt, error)
	GetGameTopNumbers(ctx context.Context, game *Game, limit int) ([]*NumberFrequencyStats, error)
	GetGameTopGuessers(ctx context.Context, game *Game, limit int) ([]*UserGuessStats, error)
	GetGameClosestGuesses(ctx context.Context, game *Game, limit int) ([]*GuessAttempt, error)
	CleanupGameData(ctx context.Context, game *Game) error
}
