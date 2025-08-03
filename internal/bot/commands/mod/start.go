package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/internal/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
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
	guildChannel, ok := event.Client().Caches().Channel(channel.ID)
	if !ok {
		return fmt.Errorf("unknown channel")
	}

	// check bot's permissions in target channel
	res := utils.CheckBotPermissionsInChannel(
		event, guildChannel,
		discord.PermissionViewChannel,
		discord.PermissionSendMessages,
		discord.PermissionManageChannels,
		discord.PermissionManageRoles,
		discord.PermissionAddReactions,
	)
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

	game, err := c.gameService.CreateGame(ctx, &types.CreateGameRequest{
		GuildID:           event.GuildID().String(),
		ChannelID:         channel.ID.String(),
		CreatedBy:         event.User().ID.String(),
		LowerBound:        int64(min),
		UpperBound:        int64(max),
		AutoReactionHints: autoReactionHints,
	})
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	// start forming ephemeral reply to user about state of everything
	var eventReply strings.Builder
	eventReply.WriteString("**Game Started**\n")
	eventReply.WriteString("Sending the game's answer to your DM.")

	// unlock target channel from lock role if it's locked
	var lockRoleErr error
	if data.GuildConfig != nil && data.GuildConfig.LockRole != "" {
		lockRole, ok := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.LockRole))
		if ok {
			lockRoleErr = utils.UnlockChannel(ctx, event, guildChannel, lockRole, "New game started, unlocking channel")
			if lockRoleErr != nil {
				eventReply.WriteString(fmt.Sprintf("\n> *Unable to unlock channel %s. Please unlock it manually so everyone can start guessing.*", guildChannel.Mention()))
			}
		}
	}

	// send game start message in target channel
	var gameStartMsg strings.Builder
	gameStartMsg.WriteString("**Game Started**\n")
	gameStartMsg.WriteString(fmt.Sprintf("Guess a number between `%d` and `%d`\n", min, max))
	gameStartMsg.WriteString(fmt.Sprintf("Correct guess will get you **%d points**", game.Game.Points))
	if data.GuildConfig != nil && data.GuildConfig.WinRole != "" {
		winRole, ok := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.WinRole))
		if ok {
			gameStartMsg.WriteString(fmt.Sprintf(" and %s role.", winRole.Mention()))
		}
	}
	msg, gameStartMsgErr := utils.SendMessage(event.Client().Rest(), utils.MessageRequest{
		ChannelID: guildChannel.ID(),
		Content:   gameStartMsg.String(),
		Emoji:     utils.EmojiSuccess,
	})
	if gameStartMsgErr != nil {
		eventReply.WriteString(fmt.Sprintf("\n> Unable to send message in %s! Please check if bot has permissions to send messages in this channel.", guildChannel.Mention()))
	} else {
		eventReply.WriteString(fmt.Sprintf("\nStart guessing: %s", msg.JumpURL()))
	}

	// pin game start message in target channel
	pinErr := event.Client().Rest().PinMessage(guildChannel.ID(), msg.ID, rest.WithReason("Pinning game start message"))
	if pinErr != nil {
		eventReply.WriteString(fmt.Sprintf("\n> Unable to pin game start message %s. See if limit of 50 pinned messages is reached. Please pin it manually so everyone would know Min and Max values.", msg.JumpURL()))
	}

	// send game start message with answer to user's DM
	dmErr := utils.SendDM(event.Client().Rest(), event.User().ID, utils.MessageRequest{
		Content: fmt.Sprintf("Game started!\nRandom answer ||%d|| has been set. Start guessing here: %s", game.Game.Answer, msg.JumpURL()),
		Emoji:   utils.EmojiSuccess,
	})
	if dmErr != nil {
		eventReply.WriteString("\n> **Unable to send DM!** Please use the `/answer` command to receive the game's answer via DM.")
	}

	// Log the game start to the log channel
	if data.GuildConfig != nil && data.GuildConfig.LogChannel != "" {
		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("**Channel:** <#%s>\n", channel.ID.String()))
		logContent.WriteString(fmt.Sprintf("**Range:** %d - %d\n", min, max))
		logContent.WriteString(fmt.Sprintf("**Points:** %d\n", game.Game.Points))
		logContent.WriteString(fmt.Sprintf("**Started by:** <@%s>\n\n", event.User().ID.String()))

		// Add status information
		logContent.WriteString("**Status Report:**\n")

		// Channel unlock status
		if data.GuildConfig != nil && data.GuildConfig.LockRole != "" {
			if lockRoleErr != nil {
				logContent.WriteString(fmt.Sprintf("❌ Failed to unlock channel: %s\n", lockRoleErr.Error()))
			} else {
				logContent.WriteString("✅ Channel unlocked successfully\n")
			}
		} else {
			logContent.WriteString("⚪ No lock role configured\n")
		}

		// Message sending status
		if gameStartMsgErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to send start message: %s\n", gameStartMsgErr.Error()))
		} else {
			logContent.WriteString("✅ Start message sent\n")
		}

		// Message pinning status
		if pinErr != nil && gameStartMsgErr == nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to pin start message: %s\n", pinErr.Error()))
		} else if gameStartMsgErr == nil {
			logContent.WriteString("✅ Start message pinned\n")
		}

		// DM status
		if dmErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to send DM with answer: %s", dmErr.Error()))
		} else {
			logContent.WriteString("✅ DM with answer sent successfully")
		}

		utils.LogToChannel(event.Client().Rest(), data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "🎮 Game Started", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		ChannelID:   channel.ID,
		Content:     eventReply.String(),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}
