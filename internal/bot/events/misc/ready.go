package misc

import (
	"context"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type ReadyEventListener struct {
	Logger *logger.Logger
}

func NewReadyEventListener(params events.EventListenerParams) *ReadyEventListener {
	return &ReadyEventListener{
		Logger: params.Logger,
	}
}

func (h *ReadyEventListener) EventName() events.EventListenerName {
	return events.Ready
}

func (h *ReadyEventListener) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *ReadyEventListener) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.Ready)
	if !ok {
		return nil
	}

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

	return nil
}
