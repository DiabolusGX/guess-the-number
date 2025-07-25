package domain

import "context"

type GuildData struct {
	ID           string                `bson:"id"`
	TotalGames   int                   `bson:"totalGames"`
	RunningGames map[string]*GameState `bson:"runningGames"`
	Users        map[string]*UserStats `bson:"users"`
}

type GameState struct {
	Answer int `bson:"answer"`
	Points int `bson:"points"`
}

type UserStats struct {
	Wins   int `bson:"wins"`
	Points int `bson:"points"`
}

type GuildDataRepository interface {
	Get(ctx context.Context, id string) (*GuildData, error)
	Replace(ctx context.Context, data *GuildData) (*GuildData, error)
	UpdateUserStats(ctx context.Context, guildID, userID string, points, wins int) error
	Create(ctx context.Context, data *GuildData) error
	Delete(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, guildID, userID string) error
}
