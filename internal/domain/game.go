package domain

import (
	"context"
	"time"
)

type Game struct {
	ID                string    `bson:"_id,omitempty"`
	GuildID           string    `bson:"guildID"`
	ChannelID         string    `bson:"channelID"`
	WinMessageID      string    `bson:"winMessageID"`
	WonBy             string    `bson:"wonBy"`
	Answer            int64     `bson:"answer"`
	Points            int64     `bson:"points"`
	Guesses           int64     `bson:"guesses"`
	FinishedAt        time.Time `bson:"finishedAt"`
	Finished          bool      `bson:"finished"`
	Disabled          bool      `bson:"disabled"`
	CreatedBy         string    `bson:"createdBy"`
	CreatedAt         time.Time `bson:"createdAt"`
	UpdatedAt         time.Time `bson:"updatedAt"`
	AutoReactionHints bool      `bson:"autoReactionHints"`
}

type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	Finish(ctx context.Context, gameID, messageID, wonBy string, guesses int64) error
	GetRunningInChannel(ctx context.Context, channelID string) (*Game, error)
	IncrementGuesses(ctx context.Context, gameID, channelID string) (int64, error)
	GetDailyLeaderboard(ctx context.Context, guildID string) ([]Game, error)
	GetAllTimeLeaderboard(ctx context.Context, guildID string) ([]Game, error)
	GetGamesBetweenDates(ctx context.Context, guildID string, startDate, endDate time.Time) ([]Game, error)
}
