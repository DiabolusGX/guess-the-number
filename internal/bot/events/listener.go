package events

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/disgoorg/disgo/bot"
	disgoEvents "github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
)

type EventListenerParams struct {
	fx.In

	Client  bot.Client
	Config  *config.Configuration
	Logger  *logger.Logger
	Metrics *metrics.Metrics

	GameService            service.GameService
	GuildManagementService service.GuildManagementService
}

type Listener interface {
	EventName() EventListenerName
	Aliases() []EventListenerName
	OnEvent(ctx context.Context, event bot.Event) error
}

type EventListenerName string

const (
	EventUnspecified EventListenerName = "Unspecified"

	MessageCreate    EventListenerName = "MessageCreate"
	Ready            EventListenerName = "Ready"
	GuildJoin        EventListenerName = "GuildJoin"
	GuildLeave       EventListenerName = "GuildLeave"
	GuildMemberLeave EventListenerName = "GuildMemberLeave"
	GuildReady       EventListenerName = "GuildReady"
	GuildsReady      EventListenerName = "GuildsReady"

	// interactions
	ApplicationCommandInteraction EventListenerName = "ApplicationCommandInteraction"
	ComponentInteraction          EventListenerName = "ComponentInteraction"
	ModalSubmit                   EventListenerName = "ModalSubmit"
)

func GetEventListenerName(event bot.Event) EventListenerName {
	switch event.(type) {
	case *disgoEvents.ApplicationCommandInteractionCreate:
		return ApplicationCommandInteraction
	case *disgoEvents.ComponentInteractionCreate:
		return ComponentInteraction
	case *disgoEvents.ModalSubmitInteractionCreate:
		return ModalSubmit
	case *disgoEvents.MessageCreate:
		return MessageCreate
	case *disgoEvents.Ready:
		return Ready
	case *disgoEvents.GuildLeave:
		return GuildLeave
	case *disgoEvents.GuildMemberLeave:
		return GuildMemberLeave
	case *disgoEvents.GuildJoin:
		return GuildJoin
	case *disgoEvents.GuildReady:
		return GuildReady
	case *disgoEvents.GuildsReady:
		return GuildsReady
	}

	return ""
}
