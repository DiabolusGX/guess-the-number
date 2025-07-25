package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/errors"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
	disgoHandler "github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type BotHandlerParams struct {
	fx.In

	Client                 bot.Client
	Config                 *config.Configuration
	Logger                 *logger.Logger
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
	client bot.Client
	logger *logger.Logger
	config *config.Configuration

	commands  map[string]commands.Command
	listeners map[events.EventListenerName]events.Listener

	ReadyTime       time.Time
	Game            service.GameService
	GuildManagement service.GuildManagementService
}

func NewBotHandler(params BotHandlerParams) BotHandler {
	return &handler{
		client: params.Client,
		logger: params.Logger,
		config: params.Config,

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

func (h *handler) OnEvent(event bot.Event) {
	if event == nil {
		return
	}

	if listener, ok := h.listeners[events.GetEventListenerName(event)]; ok {
		listener.OnEvent(event)
	}

	switch e := event.(type) {
	case *disgoEvents.ApplicationCommandInteractionCreate:
		h.OnApplicationCommandInteraction(e)
	}
}

func (h *handler) OnApplicationCommandInteraction(event *disgoEvents.ApplicationCommandInteractionCreate) {
	if command, ok := h.commands[event.Data.CommandName()]; ok {
		if err := command.Handler(event); err != nil {
			h.ErrorHandler(event, err)
		}
	}
}

func (h *handler) OnMessageCreate(event *disgoEvents.MessageCreate) {
	logger := h.logger.With(
		slog.String("guild_id", event.GuildID.String()),
		slog.String("channel_id", event.ChannelID.String()),
		slog.String("user_id", event.Message.Author.ID.String()),
	)

	if event.Message.Author.Bot {
		return
	}

	if len(event.Message.Mentions) > 0 {
		if event.Message.Mentions[0].ID == h.client.ApplicationID() {
			cfg, err := h.GuildManagement.GetGuildConfig(context.Background(), event.GuildID.String())
			if err != nil {
				logger.Error("failed to get guild config", "err", err)
				return
			}
			_ = utils.SendEmbed(h.client.Rest(), event.ChannelID, "Prefix", fmt.Sprintf("My prefix is `%s`", cfg.Prefix), 0x00FF00)
			return
		}
	}

	cfg, err := h.GuildManagement.GetGuildConfig(context.Background(), event.GuildID.String())
	if err != nil {
		logger.Error("failed to get guild config", "err", err)
		return
	}

	if !strings.HasPrefix(event.Message.Content, cfg.Prefix) {
		return
	}

	guess, err := strconv.ParseInt(strings.TrimPrefix(event.Message.Content, cfg.Prefix), 10, 64)
	if err != nil {
		return
	}

	correct, game, err := h.Game.CheckAnswer(context.Background(), event.ChannelID.String(), event.Message.Author.ID.String(), guess)
	if err != nil {
		logger.Error("failed to check answer", "err", err)
		return
	}

	if correct {
		if err := h.GuildManagement.UpdateUserStats(context.Background(), event.GuildID.String(), event.Message.Author.ID.String(), int(game.Points), 1); err != nil {
			logger.Error("failed to update user stats", "err", err)
		}

		cfg, err := h.GuildManagement.GetGuildConfig(context.Background(), event.GuildID.String())
		if err != nil {
			logger.Error("failed to get guild config", "err", err)
		}

		if cfg.WinRole != "" {
			roleID, err := snowflake.Parse(cfg.WinRole)
			if err != nil {
				logger.Error("failed to parse win role", "err", err)
			}
			if err := h.client.Rest().AddMemberRole(*event.GuildID, event.Message.Author.ID, roleID); err != nil {
				logger.Error("failed to add win role", "err", err)
			}
		}

		err = utils.SendEmbed(h.client.Rest(), event.ChannelID, "Congratulations!", fmt.Sprintf("Congratulations <@%s>! You guessed the number %d and won %d points!", event.Message.Author.ID, game.Answer, game.Points), 0x00FF00)
		if err != nil {
			logger.Error("failed to send message", "err", err)
		}
	}
}

func (h *handler) OnGuildJoin(event *disgoEvents.GuildJoin) {
	logger := h.logger.With(slog.String("guild_id", event.Guild.ID.String()))
	if err := h.GuildManagement.CreateGuild(context.Background(), event.Guild.ID.String()); err != nil {
		logger.Error("failed to create guild", "err", err)
	}
}

func (h *handler) OnGuildLeave(event *disgoEvents.GuildLeave) {
	logger := h.logger.With(slog.String("guild_id", event.GuildID.String()))
	if err := h.GuildManagement.DeleteGuild(context.Background(), event.GuildID.String()); err != nil {
		logger.Error("failed to delete guild", "err", err)
	}
}

func (h *handler) OnGuildMemberLeave(event *disgoEvents.GuildMemberLeave) {
	logger := h.logger.With(
		slog.String("guild_id", event.GuildID.String()),
		slog.String("user_id", event.User.ID.String()),
	)
	if err := h.GuildManagement.DeleteUser(context.Background(), event.GuildID.String(), event.User.ID.String()); err != nil {
		logger.Error("failed to delete user", "err", err)
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

func (h *handler) ErrorHandler(event *disgoEvents.ApplicationCommandInteractionCreate, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.WithError(err).Mark(errors.ErrCodeSystemError)
	}

	h.logger.Error("error handling command", "err", appErr)

	message := "An unexpected error occurred. Please try again later."
	if appErr.Code == errors.ErrCodeValidation {
		message = appErr.Message
	}

	_ = event.CreateMessage(discord.MessageCreate{
		Content: message,
		Flags:   discord.MessageFlagEphemeral,
	})
}
