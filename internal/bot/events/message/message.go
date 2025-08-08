package message

import (
	"context"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type MessageCreateListener struct {
	Logger          *logger.Logger
	Metrics         *metrics.Metrics
	GameService     service.GameService
	GuildManagement service.GuildManagementService
	SyncService     service.SyncService
}

func NewMessageCreateListener(params events.EventListenerParams) *MessageCreateListener {
	return &MessageCreateListener{
		Logger:          params.Logger,
		Metrics:         params.Metrics,
		GameService:     params.GameService,
		GuildManagement: params.GuildManagementService,
		SyncService:     params.SyncService,
	}
}

func (h *MessageCreateListener) EventName() events.EventListenerName {
	return events.MessageCreate
}

func (h *MessageCreateListener) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *MessageCreateListener) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.MessageCreate)
	if !ok || event.Message.Author.Bot || event.GuildID == nil {
		return nil
	}

	channel, ok := event.Channel()
	if !ok || channel.Type() != discord.ChannelTypeGuildText {
		return nil
	}

	ctx = context.WithValue(ctx, lib.CtxChannelID, event.ChannelID.String())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID.String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.Message.Author.ID.String())
	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())

	h.handleAttempt(ctx, event)

	return nil
}
