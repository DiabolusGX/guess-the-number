package message

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type MessageCreateListener struct {
	Client          bot.Client
	Logger          *logger.Logger
	Metrics         *metrics.Metrics
	GameService     service.GameService
	GuildManagement service.GuildManagementService
}

func NewMessageCreateListener(params events.EventListenerParams) *MessageCreateListener {
	return &MessageCreateListener{
		Client:          params.Client,
		Logger:          params.Logger,
		Metrics:         params.Metrics,
		GameService:     params.GameService,
		GuildManagement: params.GuildManagementService,
	}
}

func (h *MessageCreateListener) EventName() events.EventListenerName {
	return events.MessageCreate
}

func (h *MessageCreateListener) OnEvent(ctx context.Context, e bot.Event) {
	event, ok := e.(*disgoEvents.MessageCreate)
	if !ok || event.Message.Author.Bot || event.GuildID == nil {
		return
	}

	channel, ok := event.Channel()
	if !ok || channel.Type() != discord.ChannelTypeGuildText {
		return
	}

	ctx = context.WithValue(ctx, lib.CtxChannelID, event.ChannelID.String())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID.String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.Message.Author.ID.String())

	go h.handleAttempt(ctx, event)
}
