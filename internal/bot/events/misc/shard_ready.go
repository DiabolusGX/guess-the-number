package misc

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type GuildReadyEventListener struct {
	Client bot.Client
	Logger *logger.Logger
}

func NewGuildReadyEventListener(params events.EventListenerParams) *GuildReadyEventListener {
	return &GuildReadyEventListener{
		Client: params.Client,
		Logger: params.Logger,
	}
}

func (h *GuildReadyEventListener) EventName() events.EventListenerName {
	return events.GuildReady
}

func (h *GuildReadyEventListener) OnEvent(ctx context.Context, e bot.Event) {
	event, ok := e.(*disgoEvents.GuildReady)
	if !ok {
		return
	}

	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID.String())

	h.Logger.FromContext(ctx).Debugw("Guild ready",
		"guild_id", event.GuildID.String(),
		"shard_id", event.ShardID,
	)
}

type GuildsReadyEventListener struct {
	Client bot.Client
	Logger *logger.Logger
}

func NewGuildsReadyEventListener(params events.EventListenerParams) *GuildsReadyEventListener {
	return &GuildsReadyEventListener{
		Client: params.Client,
		Logger: params.Logger,
	}
}

func (h *GuildsReadyEventListener) EventName() events.EventListenerName {
	return events.GuildsReady
}

func (h *GuildsReadyEventListener) OnEvent(ctx context.Context, e bot.Event) {
	event, ok := e.(*disgoEvents.GuildsReady)
	if !ok {
		return
	}

	h.Logger.FromContext(ctx).Infow("All guilds ready for shard",
		"shard_id", event.ShardID,
	)
}
