package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/internal/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type GameCommand struct {
	name                   string
	gameService            service.GameService
	guildManagementService service.GuildManagementService
}

func NewGameCommand(params commands.CommandParams) *GameCommand {
	return &GameCommand{
		name:                   "game",
		gameService:            params.GameService,
		guildManagementService: params.GuildManagementService,
	}
}

func (c *GameCommand) Name() string {
	return c.name
}

func (c *GameCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Game management commands",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "info",
				Description: "Shows information about the current game in a channel",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel to get game info for",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "hint",
				Description: "Get a hint for the current game (Moderator only)",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "type",
						Description: "The type of hint to get",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{
								Name:  types.HintTypeNumber.String(),
								Value: types.HintTypeNumber.String(),
							},
							{
								Name:  types.HintTypeFirstDigit.String(),
								Value: types.HintTypeFirstDigit.String(),
							},
							{
								Name:  types.HintTypeLastDigit.String(),
								Value: types.HintTypeLastDigit.String(),
							},
						},
					},
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel to get the hint for",
						Required:    true,
					},
					discord.ApplicationCommandOptionInt{
						Name:        "number",
						Description: "The number to compare to the answer against",
						Required:    false,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "answer",
				Description: "Get the answer for the current game via DM (Moderator only)",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel where the game is running",
						Required:    true,
					},
				},
			},
		},
	}
}

func (c *GameCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName

	// Check permissions for mod subcommands
	if subcommand == "hint" || subcommand == "answer" {
		if !utils.CheckUserModPermissions(event, data.GuildConfig.BotManager) {
			return utils.EventReply(event, utils.MessageRequest{
				Content:     "**Access Denied**\nYou need to be an Administrator or have the Bot Manager role to use this command.",
				Emoji:       utils.EmojiError,
				IsEphemeral: true,
			})
		}
	}

	switch subcommand {
	case "info":
		return c.handleInfo(ctx, event, data)
	case "hint":
		return c.handleHint(ctx, event, data)
	case "answer":
		return c.handleAnswer(ctx, event, data)
	}
	return nil
}

func (c *GameCommand) handleInfo(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")
	guildID := event.GuildID().String()

	game, err := c.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{ChannelID: channel.ID.String()})
	if err != nil {
		return err
	}

	// Get guild data for statistics
	guildData, err := c.guildManagementService.GetGuildData(ctx, guildID)
	if err != nil {
		return err
	}

	var content strings.Builder

	if game == nil || game.Game == nil {
		// No game running - show prompt to start game and statistics
		content.WriteString(fmt.Sprintf("No game found in <#%s>\n\n", channel.ID.String()))
		content.WriteString("🎮 **Start a new game with `/start` command!**\n\n")
	} else {
		// Game is running - show current game info
		gameInfo := game.Game
		content.WriteString(fmt.Sprintf("🎯 **Game in <#%s>**\n", channel.ID.String()))
		content.WriteString(fmt.Sprintf("• Guesses so far: %d\n", gameInfo.Guesses))
		content.WriteString(fmt.Sprintf("• Points for winner: %d\n\n", gameInfo.Points))
	}

	// Add game configuration details
	guildConfig := data.GuildConfig
	content.WriteString("⚙️ **Game Configuration:**\n")

	// Required role to play
	if guildConfig.ReqRole != "" {
		content.WriteString(fmt.Sprintf("• Required role to play: <@&%s>\n", guildConfig.ReqRole))
	} else {
		content.WriteString("• Required role to play: None (everyone can play)\n")
	}

	// Winner role reward
	if guildConfig.WinRole != "" {
		content.WriteString(fmt.Sprintf("• Winner receives role: <@&%s>\n", guildConfig.WinRole))
	} else {
		content.WriteString("• Winner receives role: None\n")
	}

	// Total games statistics
	content.WriteString("\n📊 **Statistics:**\n")
	content.WriteString(fmt.Sprintf("• Total games played: %d", guildData.TotalGames))

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}

func (c *GameCommand) handleHint(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	hintType := event.SlashCommandInteractionData().String("type")
	channel := event.SlashCommandInteractionData().Channel("channel")

	hint, err := c.gameService.GetHint(ctx, &types.GetHintRequest{
		ChannelID: channel.ID.String(),
		HintType:  types.HintType(hintType),
		CompareTo: int64(event.SlashCommandInteractionData().Int("number")),
	})
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	if hint.Hint == "" {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "No hint found",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	var hintContent string

	switch hintType {
	case types.HintTypeNumber.String():
		if hint.Hint == "lower" {
			hintContent = fmt.Sprintf("The game answer is **lower than %d**.", event.SlashCommandInteractionData().Int("number"))
		} else {
			hintContent = fmt.Sprintf("The game answer is **higher than %d**.", event.SlashCommandInteractionData().Int("number"))
		}
	case types.HintTypeFirstDigit.String():
		hintContent = fmt.Sprintf("The first digit of the game answer is %s.", hint.Hint)
	case types.HintTypeLastDigit.String():
		hintContent = fmt.Sprintf("The last digit of the game answer is %s.", hint.Hint)
	}

	// send hint message
	hintMsg, messageErr := utils.SendMessage(event.Client().Rest(), utils.MessageRequest{
		ChannelID: channel.ID,
		Content:   hintContent,
		Emoji:     utils.EmojiSuccess,
	})
	if messageErr != nil {
		utils.HandleError(ctx, event, messageErr)
		return nil
	}

	// pin hint message
	pinErr := event.Client().Rest().PinMessage(channel.ID, hintMsg.ID)

	// prepare user response
	var userResponse string
	if pinErr != nil {
		userResponse = fmt.Sprintf("Successfully sent hint at %s\n> *Failed to pin the hint message. Make sure bot has required permissions and limit of 50 pinned message has not reached.* ", hintMsg.JumpURL())
	} else {
		userResponse = fmt.Sprintf("Successfully sent hint at %s", hintMsg.JumpURL())
	}

	// Log the hint to the log channel
	if data.GuildConfig != nil && data.GuildConfig.LogChannel != "" {
		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("**Channel:** <#%s>\n", channel.ID.String()))
		logContent.WriteString(fmt.Sprintf("**Hint Type:** %s\n", hintType))
		logContent.WriteString(fmt.Sprintf("**Hint:** %s\n", hintContent))
		logContent.WriteString(fmt.Sprintf("**Given by:** <@%s>\n\n", event.User().ID.String()))

		// Add status information
		logContent.WriteString("**Status Report:**\n")

		// Message sending status
		if messageErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to send hint message: %s\n", messageErr.Error()))
			logContent.WriteString("⚪ Cannot pin - message send failed")
		} else {
			logContent.WriteString("✅ Hint message sent\n")
			// Message pinning status
			if pinErr != nil {
				logContent.WriteString(fmt.Sprintf("❌ Failed to pin hint message: %s", pinErr.Error()))
			} else {
				logContent.WriteString("✅ Hint message pinned")
			}
		}

		utils.LogToChannel(event.Client().Rest(), data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "💡 Hint Given", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		Content:     userResponse,
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}

func (c *GameCommand) handleAnswer(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")

	// Get game information
	gameInfo, err := c.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{
		ChannelID: channel.ID.String(),
	})
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	// Send answer via DM
	dmContent := fmt.Sprintf("🎯 **Game Answer**\n\nThe answer for the game in <#%s> is: **%d**\n\n*Game Details:*\n• Guesses so far: %d\n• Started by: <@%s>",
		channel.ID.String(),
		gameInfo.Game.Answer,
		gameInfo.Game.Guesses,
		gameInfo.Game.CreatedBy,
	)
	dmContent += "\n\n*Please use `/game info` for more information about the game.*"

	dmErr := utils.SendDM(event.Client().Rest(), event.User().ID, utils.MessageRequest{
		Content:  dmContent,
		Emoji:    utils.EmojiSuccess,
		WithVote: true,
	})

	// Prepare user response
	var userResponse string
	var emoji utils.Emoji
	if dmErr != nil {
		userResponse = fmt.Sprintf("Failed to send answer via DM: %s\n> *Make sure your DMs are open.*", dmErr.Error())
		emoji = utils.EmojiError
	} else {
		userResponse = "Game answer sent to your DMs!"
		emoji = utils.EmojiSuccess
	}

	// Log the answer request to the log channel
	if data.GuildConfig != nil && data.GuildConfig.LogChannel != "" {
		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("**Channel:** <#%s>\n", channel.ID.String()))
		logContent.WriteString(fmt.Sprintf("**Current Guesses:** %d\n", gameInfo.Game.Guesses))
		logContent.WriteString(fmt.Sprintf("**Requested by:** <@%s>\n\n", event.User().ID.String()))

		// Add status information
		logContent.WriteString("**Status Report:**\n")

		// DM sending status
		if dmErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to send DM: %s", dmErr.Error()))
		} else {
			logContent.WriteString("✅ Answer sent via DM")
		}

		utils.LogToChannel(event.Client().Rest(), data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "🔍 Game Answer Requested", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		Content:     userResponse,
		Emoji:       emoji,
		IsEphemeral: true,
	})
}
