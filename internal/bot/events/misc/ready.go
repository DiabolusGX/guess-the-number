package misc

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type ReadyEventListener struct {
	Client bot.Client
	Logger *logger.Logger
}

func NewReadyEventListener(params events.EventListenerParams) *ReadyEventListener {
	return &ReadyEventListener{
		Client: params.Client,
		Logger: params.Logger,
	}
}

func (h *ReadyEventListener) EventName() events.EventListenerName {
	return events.Ready
}

func (h *ReadyEventListener) OnEvent(ctx context.Context, e bot.Event) {
	event, ok := e.(*disgoEvents.Ready)
	if !ok {
		return
	}

	ctx = context.WithValue(ctx, lib.CtxUserID, event.User.ID.String())

	// Log with shard information if available
	shardID := event.ShardID()
	if shardID >= 0 {
		h.Logger.FromContext(ctx).Infow("Bot shard is ready",
			"shard_id", shardID,
			"user_id", event.User.ID.String(),
			"username", event.User.Username,
		)
	} else {
		h.Logger.FromContext(ctx).Infow("Bot is ready",
			"user_id", event.User.ID.String(),
			"username", event.User.Username,
		)
	}
}
