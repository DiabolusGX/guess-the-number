package bot

import (
	"context"
	"log/slog"
	"os"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/paginator"
	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/mod"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands/user"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/message"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/misc"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
)

func NewClient(cfg *config.Configuration) (bot.Client, error) {
	var gatewayOpts []gateway.ConfigOpt

	// Add intents to gateway options and compression for better performance
	gatewayOpts = append(gatewayOpts,
		gateway.WithIntents(
			gateway.IntentDirectMessages,
			gateway.IntentGuilds,
			gateway.IntentGuildMembers,
			gateway.IntentGuildMessages,
			gateway.IntentMessageContent,
			gateway.IntentGuildMessageReactions,
		),
		gateway.WithCompress(true),
	)

	var clientOpts []bot.ConfigOpt

	logger := slog.Default()
	switch cfg.Logging.Level {
	case lib.LogLevelDebug:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case lib.LogLevelInfo:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case lib.LogLevelWarn:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case lib.LogLevelError:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	// Add cache configuration
	clientOpts = append(clientOpts,
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds|cache.FlagChannels|cache.FlagRoles|cache.FlagMembers),
			cache.WithMemberCachePolicy(func(entity discord.Member) bool {
				// cache all bots for now, ideally should only cache "self" member.
				// kept for easy testing across bots.
				return entity.User.Bot
			}),
		),
		bot.WithLogger(logger),
		bot.WithEventListeners(paginator.New()),
	)

	// Configure sharding if enabled
	if cfg.Bot.Sharding.Enabled {
		var shardOpts []sharding.ConfigOpt

		// Set shard count (0 means auto-detect) & set specific shard IDs
		if cfg.Bot.Sharding.ShardCount > 0 && len(cfg.Bot.Sharding.ShardIDs) > 0 {
			shardOpts = append(shardOpts, sharding.WithShardCount(cfg.Bot.Sharding.ShardCount))
			shardOpts = append(shardOpts, sharding.WithShardIDs(cfg.Bot.Sharding.ShardIDs...))
		}

		// Enable auto-scaling if configured
		if cfg.Bot.Sharding.AutoScaling {
			shardOpts = append(shardOpts, sharding.WithAutoScaling(true))
		}

		// Add gateway options to sharding config
		shardOpts = append(shardOpts, sharding.WithGatewayConfigOpts(gatewayOpts...))

		// Add shard manager configuration
		clientOpts = append(clientOpts, bot.WithShardManagerConfigOpts(shardOpts...))

		slog.Info("Sharding enabled",
			slog.Int("shard_count", cfg.Bot.Sharding.ShardCount),
			slog.Any("shard_ids", cfg.Bot.Sharding.ShardIDs),
			slog.Bool("auto_scaling", cfg.Bot.Sharding.AutoScaling),
		)
	} else {
		// Use regular gateway configuration for non-sharded setup
		clientOpts = append(clientOpts, bot.WithGatewayConfigOpts(gatewayOpts...))
		slog.Info("Sharding disabled, using single gateway connection")
	}

	client, err := disgo.New(cfg.Bot.Token, clientOpts...)
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
		handler.AddEventListener(misc.NewGuildReadyEventListener(eventParams))
		handler.AddEventListener(misc.NewGuildsReadyEventListener(eventParams))
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
				"sharding_enabled", cfg.Bot.Sharding.Enabled,
			)

			err := handler.SyncCommands()
			if err != nil {
				return err
			}

			client.AddEventListeners(handler)

			// Use appropriate connection method based on sharding configuration
			if cfg.Bot.Sharding.Enabled {
				logger.Info("Starting bot with sharding enabled")
				return client.OpenShardManager(ctx)
			} else {
				logger.Info("Starting bot with single gateway connection")
				return client.OpenGateway(ctx)
			}
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Guess The Number Bot shutting down")
			client.Close(ctx)
			return nil
		},
	})
}
