package domain

import "context"

type GuildConfig struct {
	ID         string `bson:"id"`
	Prefix     string `bson:"prefix"`
	Premium    bool   `bson:"premium"`
	BotManager string `bson:"botManager"`
	DM         bool   `bson:"dm"`
	Msg        string `bson:"msg"`
	WinRole    string `bson:"winRole"`
	ReqRole    string `bson:"reqRole"`
	LockRole   string `bson:"lockRole"`
}

type GuildConfigRepository interface {
	Get(ctx context.Context, id string) (*GuildConfig, error)
	Replace(ctx context.Context, cfg *GuildConfig) (*GuildConfig, error)
	Create(ctx context.Context, cfg *GuildConfig) error
	Delete(ctx context.Context, id string) error
}
