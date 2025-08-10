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
	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands/mod"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands/user"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/guild"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/interactions"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/interactions/component"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/interactions/modal"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/message"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/misc"
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
)

func NewClient(cfg *config.Configuration) (*bot.Client, error) {
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
		gateway.WithPresenceOpts(gateway.WithListeningActivity("your guesses || /help", gateway.WithActivityState("Guess what?")), gateway.WithOnlineStatus(discord.OnlineStatusOnline)),
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

	// Bot client
	fx.Provide(
		NewClient,
	),

	// Application command handler
	fx.Provide(
		commands.NewApplicationCommandHandler,
	),
	fx.Invoke(func(applicationCommandHandler *commands.ApplicationCommandHandler, commandParams commands.CommandParams) {
		// Mod commands
		applicationCommandHandler.AddCommand(mod.NewSetupCommand(commandParams))
		applicationCommandHandler.AddCommand(mod.NewStartCommand(commandParams))
		applicationCommandHandler.AddCommand(mod.NewFinishCommand(commandParams))
		applicationCommandHandler.AddCommand(mod.NewEndCommand(commandParams))

		// Mod + User commands
		applicationCommandHandler.AddCommand(user.NewGameCommand(commandParams))

		// User commands
		applicationCommandHandler.AddCommand(user.NewPingCommand(commandParams))
		applicationCommandHandler.AddCommand(user.NewUserinfoCommand(commandParams))
		applicationCommandHandler.AddCommand(user.NewInviteCommand(commandParams))
		applicationCommandHandler.AddCommand(user.NewHelpCommand(commandParams))
	}),

	// Component interactions
	fx.Provide(
		interactions.NewInteractionHandler,
	),
	fx.Invoke(func(interactionHandler *interactions.InteractionHandler, interactionParams interactions.InteractionHandlerParams) {
		interactionHandler.AddComponentInteraction(component.NewStartGameInteraction(interactionParams))
		interactionHandler.AddModalInteraction(modal.NewStartGameConfigModal(interactionParams))
	}),

	// Bot handler
	fx.Provide(
		NewBotHandler,
	),
	fx.Invoke(func(
		client *bot.Client,
		handler BotHandler,
		eventParams events.EventListenerParams,
		applicationCommandHandler *commands.ApplicationCommandHandler,
		interactionHandler *interactions.InteractionHandler,
	) {
		handler.AddEventListener(applicationCommandHandler)
		handler.AddEventListener(interactionHandler)
		handler.AddEventListener(message.NewMessageCreateListener(eventParams))
		handler.AddEventListener(misc.NewReadyEventListener(eventParams))
		handler.AddEventListener(misc.NewGuildReadyEventListener(eventParams))
		handler.AddEventListener(misc.NewGuildsReadyEventListener(eventParams))
		handler.AddEventListener(guild.NewGuildJoinListener(eventParams))
		handler.AddEventListener(guild.NewGuildLeaveListener(eventParams))
	}),
)

func Start(lc fx.Lifecycle, client *bot.Client, logger *logger.Logger, cfg *config.Configuration, handler BotHandler, syncService service.SyncService) {
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

			// Sync all active games stats from redis to mongo periodically
			go func() {
				ctx = context.WithValue(ctx, lib.CtxFlowID, "sync_all_active_games")

				defer func() {
					if r := recover(); r != nil {
						logger.FromContext(ctx).Error("Recovered from panic in sync all active games", "error", r)
					}
				}()

				err := syncService.SyncAllActiveGames(ctx)
				if err != nil {
					logger.FromContext(ctx).Error("Failed to sync all active games", "error", err)
				}
			}()

			// Use appropriate connection method based on sharding configuration
			if cfg.Bot.Sharding.Enabled {
				logger.Info("Starting bot with sharding enabled")
				// return client.OpenShardManager(ctx)

				// Start shards asynchronously to avoid Discord rate limits and Fx timeout
				// Discord allows only X shards per minute, so we need to start them in background
				go func() {
					// Use context.Background() to avoid Fx's 15-second timeout constraint
					shardCtx := context.Background()

					logger.Info("Opening shard manager asynchronously (no timeout)")
					err := client.OpenShardManager(shardCtx)
					if err != nil {
						logger.Errorw("Failed to open shard manager", "error", err)
						// Note: In a production environment, you might want to implement
						// a retry mechanism or alerting system here
					} else {
						logger.Info("Shard manager started successfully")
					}
				}()

				// Return immediately to avoid Fx timeout
				// The shards will start connecting in the background
				logger.Info("Shard startup initiated asynchronously")
				return nil
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
