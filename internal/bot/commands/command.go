package commands

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
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
	Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *Data) error
}

type Data struct {
	GuildConfig *domain.GuildConfig
}
