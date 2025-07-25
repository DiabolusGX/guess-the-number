package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/paginator"
	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/mod"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/user"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/message"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
)

func NewClient(cfg *config.Configuration) (bot.Client, error) {
	var gatewayOpts []gateway.ConfigOpt

	// Configure sharding if enabled
	if cfg.Bot.Sharding.Enabled && cfg.Bot.Sharding.ShardCount > 0 {
		// Use specified shard count for sharding
		gatewayOpts = append(gatewayOpts, gateway.WithShardCount(cfg.Bot.Sharding.ShardCount))
	}

	// Add intents to gateway options
	gatewayOpts = append(gatewayOpts, gateway.WithIntents(
		gateway.IntentGuilds,
		gateway.IntentGuildMembers,
		gateway.IntentGuildMessages,
		gateway.IntentMessageContent,
		gateway.IntentGuildMessageReactions,
	))

	client, err := disgo.New(
		cfg.Bot.Token,
		// TODO: revisit cache config
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds),
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
		handler.AddCommand(mod.NewEndCommand(commandParams))

		// User commands
		handler.AddCommand(user.NewGameInfoCommand(commandParams))
		handler.AddCommand(user.NewPingCommand(commandParams))
		handler.AddCommand(user.NewUserinfoCommand(commandParams))
		handler.AddCommand(user.NewInviteCommand(commandParams))

		// Event listeners
		handler.AddEventListener(message.NewMessageCreateListener(eventParams))
	}),
)

func NewBotProviderOptions() []fx.Option {
	return []fx.Option{
		fx.Provide(
			NewClient,
			NewBotHandler,
		),

		fx.Invoke(
			func(client bot.Client, handler BotHandler, commandParams commands.CommandParams, eventParams events.EventListenerParams) {
				// Mod commands
				handler.AddCommand(mod.NewSetupCommand(commandParams))
				handler.AddCommand(mod.NewStartCommand(commandParams))
				handler.AddCommand(mod.NewHintCommand(commandParams))
				handler.AddCommand(mod.NewEndCommand(commandParams))

				// User commands
				handler.AddCommand(user.NewGameInfoCommand(commandParams))
				handler.AddCommand(user.NewPingCommand(commandParams))
				handler.AddCommand(user.NewUserinfoCommand(commandParams))
				handler.AddCommand(user.NewInviteCommand(commandParams))

				// Event listeners
				handler.AddEventListener(message.NewMessageCreateListener(eventParams))
			},
		),
	}
}

func Start(lc fx.Lifecycle, client bot.Client, logger *logger.Logger, cfg *config.Configuration, handler BotHandler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infow(
				"Guess The Number Bot starting up",
				"mode", cfg.Deployment.Mode,
				"log_level", cfg.Logging.Level,
				"sentry_enabled", cfg.Sentry.Enabled,
				"sharding_enabled", cfg.Bot.Sharding.Enabled,
				"shard_count", cfg.Bot.Sharding.ShardCount,
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
