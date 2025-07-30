package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/errors"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
	disgoHandler "github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type BotHandlerParams struct {
	fx.In

	Client  bot.Client
	Config  *config.Configuration
	Logger  *logger.Logger
	Metrics *metrics.Metrics

	GameService            service.GameService
	GuildManagementService service.GuildManagementService
}

type BotHandler interface {
	bot.EventListener

	AddCommand(command commands.Command)
	AddEventListener(listener events.Listener)
	SyncCommands() error
}

type handler struct {
	client  bot.Client
	logger  *logger.Logger
	config  *config.Configuration
	metrics *metrics.Metrics

	commands  map[string]commands.Command
	listeners map[events.EventListenerName]events.Listener

	ReadyTime       time.Time
	Game            service.GameService
	GuildManagement service.GuildManagementService
}

func NewBotHandler(params BotHandlerParams) BotHandler {
	return &handler{
		client:  params.Client,
		logger:  params.Logger,
		config:  params.Config,
		metrics: params.Metrics,

		commands:  make(map[string]commands.Command),
		listeners: make(map[events.EventListenerName]events.Listener),

		Game:            params.GameService,
		GuildManagement: params.GuildManagementService,
	}
}

func (h *handler) AddCommand(command commands.Command) {
	h.commands[command.Name()] = command
}

func (h *handler) AddEventListener(listener events.Listener) {
	h.listeners[listener.EventName()] = listener
}

// isModCommand checks if a command name is a mod command
func isModCommand(commandName string) bool {
	modCommands := []string{"start", "setup", "hint", "end"}
	for _, cmd := range modCommands {
		if cmd == commandName {
			return true
		}
	}
	return false
}

func (h *handler) OnEvent(event bot.Event) {
	if event == nil {
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, lib.CtxRequestID, lib.NewRequestID())

	// Track event metrics
	eventName := string(events.GetEventListenerName(event))
	h.metrics.Operations.WithLabelValues("event", eventName, "N/A").Inc()

	// panic recovery
	defer func() {
		if r := recover(); r != nil {
			h.logger.FromContext(ctx).Errorw("panic in event listener", "err", r)
		}
	}()

	if listener, ok := h.listeners[events.GetEventListenerName(event)]; ok {
		listener.OnEvent(ctx, event)
	}

	switch e := event.(type) {
	case *disgoEvents.ApplicationCommandInteractionCreate:
		// acknowledge the discord interaction
		err := e.DeferCreateMessage(true)
		if err != nil {
			h.logger.FromContext(ctx).Errorw("error acknowledging discord interaction", "err", err)
		}

		h.OnApplicationCommandInteraction(ctx, e)
	}
}

func (h *handler) OnApplicationCommandInteraction(ctx context.Context, event *disgoEvents.ApplicationCommandInteractionCreate) {
	ctx = context.WithValue(ctx, lib.CtxChannelID, event.Channel().ID().String())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID().String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.User().ID.String())

	commandName := event.Data.CommandName()
	var commandSuccess = "true"

	// Track command metrics (will be updated based on success/failure)
	defer func() {
		h.metrics.Operations.WithLabelValues("command", commandName, commandSuccess).Inc()
	}()

	appPermissions := event.AppPermissions()
	res := utils.CheckBotPermissions(appPermissions, discord.PermissionViewChannel, discord.PermissionSendMessages)
	if !res.HasAllPermissions {
		h.logger.FromContext(ctx).Infow("missing permissions", "missing_permissions", strings.Join(res.MissingPermissions, ", "))
		commandSuccess = "false"
		return
	}

	guildConfig, err := h.GuildManagement.GetGuildConfig(ctx, event.GuildID().String())
	if err != nil {
		h.ErrorHandler(ctx, event, err)
		commandSuccess = "false"
		return
	}

	// Check permissions for mod commands
	if isModCommand(commandName) {
		if !utils.CheckUserModPermissions(event, guildConfig.BotManager) {
			content := "❌ **Access Denied**\nYou need to be an Administrator or have the Bot Manager role to use this command."
			_, _ = event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: &content,
			})
			commandSuccess = "false"
			h.logger.FromContext(ctx).Infow("user lacks permissions for mod command", "command", commandName, "user_id", event.User().ID.String())
			return
		}
	}

	commonData := &commands.Data{
		GuildConfig: guildConfig,
	}

	if command, ok := h.commands[commandName]; ok {
		if err := command.Handler(ctx, event, commonData); err != nil {
			h.ErrorHandler(ctx, event, err)
			commandSuccess = "false"
		}
	} else {
		commandSuccess = "false"
	}
}

func (h *handler) SyncCommands() error {
	if !h.config.Bot.SyncCommands {
		return nil
	}

	guildIDs := make([]snowflake.ID, 0, len(h.config.Bot.DevGuilds))
	for _, guildID := range h.config.Bot.DevGuilds {
		guildIDs = append(guildIDs, snowflake.MustParse(guildID))
	}

	commands := make([]discord.ApplicationCommandCreate, 0, len(h.commands))
	for _, command := range h.commands {
		commands = append(commands, command.Definition())
	}

	if err := disgoHandler.SyncCommands(h.client, commands, guildIDs); err != nil {
		return fmt.Errorf("failed to sync commands: %w", err)
	}

	return nil
}

func (h *handler) ErrorHandler(ctx context.Context, event *disgoEvents.ApplicationCommandInteractionCreate, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.WithError(err).Mark(errors.ErrCodeInternalError)
	}

	h.metrics.Operations.WithLabelValues("command", event.Data.CommandName(), "false").Inc()

	h.logger.FromContext(ctx).Errorw("error handling command", "err", appErr.Error())

	message := "An unexpected error occurred. Please try again later."
	if appErr.Code == errors.ErrCodeValidation {
		message = appErr.Message
	}

	_, _ = event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Content: &message,
	})
}
