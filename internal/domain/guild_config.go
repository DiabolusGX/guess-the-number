package domain

import (
	"context"
)

type GuildConfig struct {
	ID                string `bson:"id,omitempty"`
	Prefix            string `bson:"prefix,omitempty"`
	Premium           bool   `bson:"premium,omitempty"`
	BotManager        string `bson:"botManager,omitempty"`
	DM                bool   `bson:"dm,omitempty"`
	Msg               string `bson:"msg,omitempty"`
	WinRole           string `bson:"winRole,omitempty"`
	ReqRole           string `bson:"reqRole,omitempty"`
	LockRole          string `bson:"lockRole,omitempty"`
	LogChannel        string `bson:"logChannel,omitempty"`
	AutoReactionHints bool   `bson:"autoReactionHints,omitempty"`
}

type GuildConfigRepository interface {
	Get(ctx context.Context, id string) (*GuildConfig, error)
	Replace(ctx context.Context, cfg *GuildConfig) error
	Create(ctx context.Context, cfg *GuildConfig) error
	Delete(ctx context.Context, id string) error
}
