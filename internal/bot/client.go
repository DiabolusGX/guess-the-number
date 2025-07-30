package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/paginator"
	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/mod"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/user"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/message"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/misc"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
)

func NewClient(cfg *config.Configuration) (bot.Client, error) {
	var gatewayOpts []gateway.ConfigOpt

	// Add intents to gateway options
	gatewayOpts = append(gatewayOpts, gateway.WithIntents(
		gateway.IntentDirectMessages,
		gateway.IntentGuilds,
		gateway.IntentGuildMembers,
		gateway.IntentGuildMessages,
		gateway.IntentMessageContent,
		gateway.IntentGuildMessageReactions,
	))

	client, err := disgo.New(
		cfg.Bot.Token,
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds|cache.FlagChannels|cache.FlagRoles|cache.FlagMembers),
			cache.WithMemberCachePolicy(func(entity discord.Member) bool {
				// cache all bots for now, ideally should only cache "self" member.
				// kept for easy testing across bots.
				return entity.User.Bot
			}),
		),
		bot.WithGatewayConfigOpts(gatewayOpts...),
		bot.WithLogger(slog.Default()),
		bot.WithEventListeners(paginator.New()),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

var Module = fx.Module(
	"bot",

	fx.Provide(
		NewClient,
		NewBotHandler,
	),

	fx.Invoke(func(client bot.Client, handler BotHandler, commandParams commands.CommandParams, eventParams events.EventListenerParams) {
		// Mod commands
		handler.AddCommand(mod.NewSetupCommand(commandParams))
		handler.AddCommand(mod.NewStartCommand(commandParams))
		handler.AddCommand(mod.NewHintCommand(commandParams))
		handler.AddCommand(mod.NewFinishCommand(commandParams))

		// User commands
		handler.AddCommand(user.NewGameInfoCommand(commandParams))
		handler.AddCommand(user.NewPingCommand(commandParams))
		handler.AddCommand(user.NewUserinfoCommand(commandParams))
		handler.AddCommand(user.NewInviteCommand(commandParams))

		// Event listeners
		handler.AddEventListener(message.NewMessageCreateListener(eventParams))
		handler.AddEventListener(misc.NewReadyEventListener(eventParams))
	}),
)

func Start(lc fx.Lifecycle, client bot.Client, logger *logger.Logger, cfg *config.Configuration, handler BotHandler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infow(
				"Guess The Number Bot starting up",
				"mode", cfg.Deployment.Mode,
				"log_level", cfg.Logging.Level,
				"sentry_enabled", cfg.Sentry.Enabled,
			)

			err := handler.SyncCommands()
			if err != nil {
				return err
			}

			client.AddEventListeners(handler)
			return client.OpenGateway(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Guess The Number Bot shutting down")
			client.Close(ctx)
			return nil
		},
	})
}
