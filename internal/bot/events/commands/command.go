package commands

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
)

type CommandParams struct {
	fx.In

	Config                 *config.Configuration
	Logger                 *logger.Logger
	Metrics                *metrics.Metrics
	GameService            service.GameService
	GuildManagementService service.GuildManagementService
}

type Command interface {
	Name() string
	Definition() discord.ApplicationCommandCreate
	Handler(ctx context.Context, event *disgoEvents.ApplicationCommandInteractionCreate, data *Data) error
}

type Data struct {
	GuildConfig *domain.GuildConfig
}
