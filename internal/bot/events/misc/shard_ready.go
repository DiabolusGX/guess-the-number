package misc

import (
	"context"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type GuildReadyEventListener struct {
	Logger *logger.Logger
}

func NewGuildReadyEventListener(params events.EventListenerParams) *GuildReadyEventListener {
	return &GuildReadyEventListener{
		Logger: params.Logger,
	}
}

func (h *GuildReadyEventListener) EventName() events.EventListenerName {
	return events.GuildReady
}

func (h *GuildReadyEventListener) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *GuildReadyEventListener) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.GuildReady)
	if !ok {
		return nil
	}

	h.Logger.FromContext(ctx).Debugw("Guild ready",
		"guild_id", event.GuildID.String(),
		"shard_id", event.ShardID(),
	)

	return nil
}

type GuildsReadyEventListener struct {
	Logger *logger.Logger
}

func NewGuildsReadyEventListener(params events.EventListenerParams) *GuildsReadyEventListener {
	return &GuildsReadyEventListener{
		Logger: params.Logger,
	}
}

func (h *GuildsReadyEventListener) EventName() events.EventListenerName {
	return events.GuildsReady
}

func (h *GuildsReadyEventListener) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *GuildsReadyEventListener) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.GuildsReady)
	if !ok {
		return nil
	}

	h.Logger.FromContext(ctx).Infow("All guilds ready for shard",
		"shard_id", event.ShardID(),
	)

	return nil
}
