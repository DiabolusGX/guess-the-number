package commands

import (
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
)

type CommandParams struct {
	fx.In

	Logger                 *logger.Logger
	GameService            service.GameService
	GuildManagementService service.GuildManagementService
}

type Command interface {
	Name() string
	Definition() discord.ApplicationCommandCreate
	Handler(event *events.ApplicationCommandInteractionCreate) error
}
