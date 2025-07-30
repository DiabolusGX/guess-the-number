package domain

import "context"

type GuildData struct {
	ID           string                `bson:"id"`
	TotalGames   int64                 `bson:"totalGames"`
	RunningGames map[string]*GameState `bson:"runningGames"`
	Users        map[string]*UserStats `bson:"users"`
}

type GameState struct {
	Answer int64 `bson:"answer"`
	Points int64 `bson:"points"`
}

type UserStats struct {
	Wins   int64 `bson:"wins"`
	Points int64 `bson:"points"`
}

type GuildDataRepository interface {
	Get(ctx context.Context, id string) (*GuildData, error)
	Replace(ctx context.Context, data *GuildData) (*GuildData, error)
	UpdateGameAndUserStats(ctx context.Context, guildID, channelID, userID string, points, wins int64) error
	// NOTE: merged these two into UpdateGameAndUserStats however, we can keep them here & in repository implementation for now
	// UpdateGameStats(ctx context.Context, guildID, channelID string) error
	// UpdateUserStats(ctx context.Context, guildID, userID string, points, wins int64) error
	Create(ctx context.Context, data *GuildData) error
	Delete(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, guildID, userID string) error
}
