package message

import (
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type MessageCreateListener struct {
	Client bot.Client
	Logger *logger.Logger
}

func NewMessageCreateListener(params events.EventListenerParams) *MessageCreateListener {
	return &MessageCreateListener{
		Client: params.Client,
		Logger: params.Logger,
	}
}

func (h *MessageCreateListener) EventName() events.EventListenerName {
	return events.MessageCreate
}

func (h *MessageCreateListener) OnEvent(e bot.Event) {
	event, ok := e.(*disgoEvents.MessageCreate)
	if !ok {
		return
	}

	if event.Message.Author.Bot {
		return
	}

	if event.Message.Content == "!ping" {
		utils.SendEmbed(h.Client.Rest(), event.Message.ChannelID, "Ping", "Pong", 0)
	}
}
