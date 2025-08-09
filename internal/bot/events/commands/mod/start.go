package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/commons"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type StartCommand struct {
	name        string
	gameService service.GameService
}

func NewStartCommand(params commands.CommandParams) *StartCommand {
	return &StartCommand{
		name:        "start",
		gameService: params.GameService,
	}
}

func (c *StartCommand) Name() string {
	return c.name
}

func (c *StartCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Starts a new game of guess the number in current channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "min",
				Description: "The minimum number",
				Required:    true,
			},
			discord.ApplicationCommandOptionInt{
				Name:        "max",
				Description: "The maximum number",
				Required:    true,
			},
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to start the game in",
				Required:    true,
			},
			discord.ApplicationCommandOptionBool{
				Name:        "auto-reaction-hints",
				Description: "Enable automatic reaction hints (⬆️/⬇️) for wrong guesses",
				Required:    false,
			},
		},
	}
}

func (c *StartCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	min := event.SlashCommandInteractionData().Int("min")
	max := event.SlashCommandInteractionData().Int("max")
	channel := event.SlashCommandInteractionData().Channel("channel")
	targetChannel, ok := event.Client().Caches().Channel(channel.ID)
	if !ok {
		return fmt.Errorf("unknown channel")
	}

	requiredPermissions := []discord.Permissions{
		discord.PermissionViewChannel,
		discord.PermissionSendMessages,
		discord.PermissionEmbedLinks,
	}
	if data.GuildConfig != nil && data.GuildConfig.AutoReactionHints {
		requiredPermissions = append(requiredPermissions, discord.PermissionAddReactions)
	}
	if data.GuildConfig != nil && data.GuildConfig.LockRole != "" {
		requiredPermissions = append(requiredPermissions, discord.PermissionManageChannels)
	}
	if data.GuildConfig != nil && data.GuildConfig.WinRole != "" {
		requiredPermissions = append(requiredPermissions, discord.PermissionManageRoles)
	}

	// check bot's permissions in target channel
	res := utils.CheckBotPermissionsInChannel(event, targetChannel, requiredPermissions...)
	if !res.HasAllPermissions {
		utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("**Missing permissions**\nPlease grant the following permissions to the bot and try again: %s", strings.Join(res.MissingPermissions, ", ")),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
		return nil
	}

	// Handle auto-reaction-hints parameter
	var autoReactionHints bool
	if value, ok := event.SlashCommandInteractionData().OptBool("auto-reaction-hints"); ok {
		autoReactionHints = value
	}
	if !autoReactionHints && data.GuildConfig != nil && data.GuildConfig.AutoReactionHints {
		autoReactionHints = data.GuildConfig.AutoReactionHints
	}

	// Start the game using common logic
	gameRequest := &commons.GameStartRequest{
		Client:            event.Client(),
		TargetChannel:     targetChannel,
		CreatedBy:         event.User(),
		LowerBound:        int64(min),
		UpperBound:        int64(max),
		AutoReactionHints: autoReactionHints,
	}

	result, err := commons.StartGameCommon(
		ctx,
		event.Client().Rest(),
		c.gameService,
		gameRequest,
		data.GuildConfig,
	)
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	// Format the reply message
	eventReply := commons.FormatGameStartReply(result, targetChannel, data.GuildConfig)
	return utils.EventReply(event, utils.MessageRequest{
		ChannelID:   channel.ID,
		Emoji:       utils.EmojiSuccess,
		Content:     eventReply,
		IsEphemeral: true,
	})
}
