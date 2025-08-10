package commands

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
	disgoHandler "github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/service"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
)

type ApplicationCommandHandler struct {
	Config  *config.Configuration
	Client  *bot.Client
	Logger  *logger.Logger
	Metrics *metrics.Metrics

	GameService            service.GameService
	GuildManagementService service.GuildManagementService

	Commands map[string]Command
}

func NewApplicationCommandHandler(params events.EventListenerParams) *ApplicationCommandHandler {
	return &ApplicationCommandHandler{
		Config:  params.Config,
		Client:  params.Client,
		Logger:  params.Logger,
		Metrics: params.Metrics,

		GameService:            params.GameService,
		GuildManagementService: params.GuildManagementService,

		Commands: make(map[string]Command),
	}
}

func (h *ApplicationCommandHandler) EventName() events.EventListenerName {
	return events.ApplicationCommandInteraction
}

func (h *ApplicationCommandHandler) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *ApplicationCommandHandler) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.ApplicationCommandInteractionCreate)
	if !ok {
		return nil
	}

	// acknowledge the discord interaction
	// NOTE: need to decide if we want to keep defer ephemeral or not
	err := event.DeferCreateMessage(false)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("error acknowledging discord interaction", "err", err)
	}

	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())
	ctx = context.WithValue(ctx, lib.CtxChannelID, event.Channel().ID().String())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID().String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.User().ID.String())

	commandName := event.Data.CommandName()
	var commandSuccess = "true"

	err = h.handleCommand(ctx, event)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("error handling command", "err", err)
		commandSuccess = "false"
	}

	h.Metrics.Operations.WithLabelValues("command", commandName, commandSuccess).Inc()
	return nil
}

func (h *ApplicationCommandHandler) AddCommand(command Command) {
	h.Commands[command.Name()] = command
}

func (h *ApplicationCommandHandler) handleCommand(ctx context.Context, event *disgoEvents.ApplicationCommandInteractionCreate) error {
	commandName := event.Data.CommandName()

	appPermissions := event.AppPermissions()
	res := utils.CheckBotPermissions(appPermissions, discord.PermissionViewChannel, discord.PermissionSendMessages, discord.PermissionEmbedLinks)
	if !res.HasAllPermissions {
		h.Logger.FromContext(ctx).Infow("bot is missing permissions", "missing_permissions", strings.Join(res.MissingPermissions, ", "))
		return nil
	}

	guildConfig, err := h.GuildManagementService.GetGuildConfig(ctx, event.GuildID().String())
	if err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	// Check permissions for mod commands
	if isModCommand(commandName) {
		if !utils.CheckUserModPermissions(event, guildConfig.BotManager) {
			content := "❌ **Access Denied**\nYou need to be an Administrator or have the Bot Manager role to use this command."
			_, _ = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: &content,
			})
			h.Logger.FromContext(ctx).Infow("user is missing permissions for mod command", "command", commandName, "user_id", event.User().ID.String())
			return ierr.New(ierr.ErrCodeMissingPermissions, "user is missing permissions")
		}
	}

	commonData := &Data{
		GuildConfig: guildConfig,
	}

	command, ok := h.Commands[commandName]
	if !ok {
		return ierr.New(ierr.ErrCodeNotFound, "command not found")
	}

	if err := command.Handler(ctx, event, commonData); err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	return nil
}

func (h *ApplicationCommandHandler) SyncCommands() error {
	if !h.Config.Bot.SyncCommands {
		return nil
	}

	guildIDs := make([]snowflake.ID, 0, len(h.Config.Bot.DevGuilds))
	for _, guildID := range h.Config.Bot.DevGuilds {
		guildIDs = append(guildIDs, snowflake.MustParse(guildID))
	}

	commands := make([]discord.ApplicationCommandCreate, 0, len(h.Commands))
	for _, command := range h.Commands {
		commands = append(commands, command.Definition())
	}

	if err := disgoHandler.SyncCommands(h.Client, commands, guildIDs); err != nil {
		return fmt.Errorf("failed to sync commands: %w", err)
	}

	return nil
}

// isModCommand checks if a command name is a mod command
func isModCommand(commandName string) bool {
	modCommands := []string{"start", "setup", "finish-game", "end"}
	return slices.Contains(modCommands, commandName)
}
