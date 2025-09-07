package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/internal/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
)

type GameCommand struct {
	name                   string
	gameService            service.GameService
	guildManagementService service.GuildManagementService
	statsService           service.StatsService
}

func NewGameCommand(params commands.CommandParams) *GameCommand {
	return &GameCommand{
		name:                   "game",
		gameService:            params.GameService,
		guildManagementService: params.GuildManagementService,
		statsService:           params.StatsService,
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
			discord.ApplicationCommandOptionSubCommand{
				Name:        "stats",
				Description: "View game statistics",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "type",
						Description: "Type of statistics to view",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{Name: "Closest Guesses", Value: "closest-guesses"},
							{Name: "Most frequently guessed numbers", Value: "top-numbers"},
							{Name: "Guessers with most unique numbers", Value: "top-guessers"},
							{Name: "Top Game Winners", Value: "top-game-winners"},
							{Name: "Top Points Earners", Value: "top-points"},
						},
					},
					discord.ApplicationCommandOptionString{
						Name:        "game-id",
						Description: "Specific game ID (for game-specific stats)",
						Required:    false,
					},
					discord.ApplicationCommandOptionString{
						Name:        "time-range",
						Description: "Time range for statistics",
						Required:    false,
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{Name: "Last Day", Value: "last-day"},
							{Name: "Last Week", Value: "last-week"},
							{Name: "Last Month", Value: "last-month"},
							{Name: "All Time", Value: "all-time"},
						},
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
				UseEmbed:         true,
				Emoji:            utils.EmojiError,
				EmbedTitle:       "Access Denied",
				EmbedDescription: "You need to be an Administrator or have the Bot Manager role to use this command.",
				EmbedColor:       utils.FailureEmbedColor,
				IsEphemeral:      true,
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
	case "stats":
		return c.handleStats(ctx, event, data)
	}
	return nil
}

func (c *GameCommand) handleInfo(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")

	game, err := c.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{ChannelID: channel.ID.String()})
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	var gameStatus strings.Builder
	var embedTitle string

	if game == nil || game.Game == nil {
		// No game running - show prompt to start game
		embedTitle = "🎮 Game Info"
		gameStatus.WriteString(fmt.Sprintf("No game currently running in <#%s>.\n\n", channel.ID.String()))
		gameStatus.WriteString(fmt.Sprintf("🎮 **Start a new game with %s command!**", utils.MentionApplicationCommand(event.Client().ID(), utils.CommandStart)))
	} else {
		// Game is running - show current game info
		gameInfo := game.Game
		embedTitle = "🎯 Running Game"
		gameStatus.WriteString(fmt.Sprintf("**Game ID:** `%s`\nRunning in <#%s>\n\n", gameInfo.ID, channel.ID.String()))
		gameStatus.WriteString(fmt.Sprintf("Answer lies between `%d` and `%d`\n", gameInfo.LowerBound, gameInfo.UpperBound))
		gameStatus.WriteString(fmt.Sprintf("Guesses so far: **%d**\n", gameInfo.Guesses))
		gameStatus.WriteString(fmt.Sprintf("Points for winner: **%d**\n", gameInfo.Points))
		gameStatus.WriteString(fmt.Sprintf("*Use %s with game ID for detailed statistics*", utils.MentionApplicationCommand(event.Client().ID(), utils.CommandGameStats)))
	}

	// Add game configuration details
	guildConfig := data.GuildConfig
	var configDetails strings.Builder

	// Required role to play
	if guildConfig.ReqRole != "" {
		configDetails.WriteString(fmt.Sprintf("• Required role: <@&%s>\n", guildConfig.ReqRole))
	} else {
		configDetails.WriteString("• Required role: None (everyone can play)\n")
	}

	// Winner role reward
	if guildConfig.WinRole != "" {
		configDetails.WriteString(fmt.Sprintf("• Winner role: <@&%s>\n", guildConfig.WinRole))
	} else {
		configDetails.WriteString("• Winner role: None\n")
	}

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiInfo,
		EmbedTitle:       embedTitle,
		EmbedDescription: gameStatus.String(),
		EmbedColor:       utils.InfoEmbedColor,
		Fields: []discord.EmbedField{
			{Name: "⚙️ Configuration", Value: configDetails.String(), Inline: omit.NewPtr(false).Value},
		},
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
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "No Hint Available",
			EmbedDescription: "No hint found for the current game in this channel.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
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
	hintMsg, messageErr := utils.SendMessage(event.Client().Rest, utils.MessageRequest{
		ChannelID: channel.ID,
		Content:   hintContent,
		Emoji:     utils.EmojiSuccess,
	})
	if messageErr != nil {
		utils.HandleError(ctx, event, messageErr)
		return nil
	}

	// pin hint message
	pinErr := event.Client().Rest.PinMessage(channel.ID, hintMsg.ID)

	// prepare user response
	var embedTitle, embedDescription string
	var embedColor int
	if pinErr != nil {
		embedTitle = "Hint Sent (Pin Failed)"
		embedDescription = fmt.Sprintf("Successfully sent hint at %s\n\n*Failed to pin the hint message. Make sure bot has required permissions and limit of 50 pinned messages has not been reached.*", hintMsg.JumpURL())
		embedColor = utils.WarningEmbedColor
	} else {
		embedTitle = "Hint Sent Successfully"
		embedDescription = fmt.Sprintf("Successfully sent and pinned hint at %s", hintMsg.JumpURL())
		embedColor = utils.SuccessEmbedColor
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

		// Message sending status (always successful at this point)
		logContent.WriteString("✅ Hint message sent\n")
		// Message pinning status
		if pinErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to pin hint message: %s", pinErr.Error()))
		} else {
			logContent.WriteString("✅ Hint message pinned")
		}

		// Add game id to log
		logContent.WriteString(fmt.Sprintf("\n\n**Game ID:** `%s`", hint.GameID))

		utils.LogToChannel(event.Client().Rest, data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "💡 Hint Given", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       embedTitle,
		EmbedDescription: embedDescription,
		EmbedColor:       embedColor,
		IsEphemeral:      true,
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

	dmErr := utils.SendDM(event.Client().Rest, event.User().ID, utils.MessageRequest{
		Content:  dmContent,
		WithVote: true,
	})

	// Prepare user response
	var embedTitle, embedDescription string
	var embedColor int
	var emoji utils.Emoji
	if dmErr != nil {
		embedTitle = "DM Failed"
		embedDescription = fmt.Sprintf("Failed to send answer via DM: %s\n\n*Make sure your DMs are open.*", dmErr.Error())
		embedColor = utils.FailureEmbedColor
		emoji = utils.EmojiError
	} else {
		embedTitle = "Answer Sent"
		embedDescription = "Game answer has been sent to your DMs!"
		embedColor = utils.SuccessEmbedColor
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

		// Add game id to log
		logContent.WriteString(fmt.Sprintf("\n\n**Game ID:** `%s`", gameInfo.Game.ID))

		utils.LogToChannel(event.Client().Rest, data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "🔍 Game Answer Requested", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            emoji,
		EmbedTitle:       embedTitle,
		EmbedDescription: embedDescription,
		EmbedColor:       embedColor,
		IsEphemeral:      true,
	})
}

func (c *GameCommand) handleStats(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	statsType := event.SlashCommandInteractionData().String("type")
	gameID, hasGameID := event.SlashCommandInteractionData().OptString("game-id")
	timeRangeStr, hasTimeRange := event.SlashCommandInteractionData().OptString("time-range")
	if !hasTimeRange {
		timeRangeStr = "all-time"
	}

	// Parse time range
	timeRange := domain.TimeRange{Type: timeRangeStr}

	guildID := event.GuildID().String()
	userID := event.User().ID.String()

	var content strings.Builder
	var noContent bool
	if hasGameID {
		content.WriteString(fmt.Sprintf("**Game ID:** `%s`\n", gameID))
	}
	content.WriteString(fmt.Sprintf("**Time Range:** `%s`\n\n", timeRangeStr))

	switch statsType {
	case "closest-guesses":
		if !hasGameID {
			return utils.EventReply(event, utils.MessageRequest{
				UseEmbed:         true,
				Emoji:            utils.EmojiError,
				EmbedTitle:       "Game ID Required",
				EmbedDescription: "Closest guesses stats require a specific game ID.",
				EmbedColor:       utils.FailureEmbedColor,
				IsEphemeral:      true,
			})
		}

		// For closest guesses, we need to check permissions at the command level
		isMod := utils.CheckUserModPermissions(event, data.GuildConfig.BotManager)
		stats, err := c.statsService.GetClosestGuesses(ctx, &types.GetClosestGuessesRequest{
			RequesterUserID: userID,
			IsUserMod:       isMod,
			GuildID:         guildID,
			GameID:          gameID,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}

		if stats.Game != nil && stats.Game.Finished {
			content.WriteString(fmt.Sprintf("Game winner: <@%s> (at %s)\n", stats.Game.WonBy, utils.FormMessageLink(guildID, stats.Game.ChannelID, stats.Game.WinMessageID)))
			content.WriteString(fmt.Sprintf("Game answer: **%d** 🎉 (guessed after **%d attempts**)\n\n", stats.Game.Answer, stats.Game.Guesses))
		}

		if len(stats.Guesses) == 0 {
			noContent = true
		} else {
			for i, guess := range stats.Guesses {
				content.WriteString(fmt.Sprintf("`%d.` <@%s> guessed **%d**\n", i+1, guess.UserID, guess.Guess))
			}
		}

	case "top-numbers":
		result, err := c.statsService.GetTopGuessedNumbers(ctx, &types.GetTopGuessedNumbersRequest{
			RequesterUserID: userID,
			GuildID:         guildID,
			GameID:          gameID,
			TimeRange:       timeRange,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}

		if result.Game != nil && result.Game.Finished {
			content.WriteString(fmt.Sprintf("Game winner: <@%s> (at %s)\n", result.Game.WonBy, utils.FormMessageLink(guildID, result.Game.ChannelID, result.Game.WinMessageID)))
			content.WriteString(fmt.Sprintf("Game answer: **%d** 🎉 (guessed after **%d attempts**)\n\n", result.Game.Answer, result.Game.Guesses))
		}

		if len(result.Numbers) == 0 {
			noContent = true
		} else {
			for i, stat := range result.Numbers {
				content.WriteString(fmt.Sprintf("`%d.` **%d** - guessed **%d** times\n", i+1, stat.Number, stat.Count))
			}
		}

	case "top-guessers":
		result, err := c.statsService.GetTopGuessers(ctx, &types.GetTopGuessersRequest{
			RequesterUserID: userID,
			GuildID:         guildID,
			GameID:          gameID,
			TimeRange:       timeRange,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}

		if result.Game != nil && result.Game.Finished {
			content.WriteString(fmt.Sprintf("Game winner: <@%s> (at %s)\n", result.Game.WonBy, utils.FormMessageLink(guildID, result.Game.ChannelID, result.Game.WinMessageID)))
			content.WriteString(fmt.Sprintf("Game answer: **%d** 🎉 (guessed after **%d attempts**)\n\n", result.Game.Answer, result.Game.Guesses))
		}

		if len(result.TopGuessers) == 0 {
			noContent = true
		} else {
			for i, stat := range result.TopGuessers {
				content.WriteString(fmt.Sprintf("`%d.` <@%s> - **%d** unique guesses\n", i+1, stat.UserID, stat.TotalGuesses))
			}
		}

	case "top-game-winners":
		if hasTimeRange {
			return utils.EventReply(event, utils.MessageRequest{
				UseEmbed:         true,
				Emoji:            utils.EmojiError,
				EmbedTitle:       "Time Range not required",
				EmbedDescription: "Top game winners stats do not require a time range. Winners are ranked by all-time total wins or points.",
				EmbedColor:       utils.FailureEmbedColor,
				IsEphemeral:      true,
			})
		}

		result, err := c.statsService.GetTopWinners(ctx, &types.GetTopWinnersRequest{
			RequesterUserID: userID,
			GuildID:         guildID,
			ByGame:          true,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}

		if len(result.Winners) == 0 {
			noContent = true
		} else {
			for i, stat := range result.Winners {
				content.WriteString(fmt.Sprintf("`%d.` <@%s> - **%d wins**\n", i+1, stat.UserID, stat.Wins))
			}
		}

	case "top-points":
		if hasTimeRange {
			return utils.EventReply(event, utils.MessageRequest{
				UseEmbed:         true,
				Emoji:            utils.EmojiError,
				EmbedTitle:       "Time Range not required",
				EmbedDescription: "Top game winners stats do not require a time range. Winners are ranked by all-time total wins or points.",
				EmbedColor:       utils.FailureEmbedColor,
				IsEphemeral:      true,
			})
		}

		result, err := c.statsService.GetTopWinners(ctx, &types.GetTopWinnersRequest{
			RequesterUserID: userID,
			GuildID:         guildID,
			ByPoints:        true,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}

		if len(result.Winners) == 0 {
			noContent = true
		} else {
			for i, stat := range result.Winners {
				content.WriteString(fmt.Sprintf("`%d.` <@%s> - **%d points**\n", i+1, stat.UserID, stat.Points))
			}
		}

	default:
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Stats Type",
			EmbedDescription: "Please select a valid statistics type.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	if noContent {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiInfo,
			EmbedTitle:       "Not enough data",
			EmbedDescription: "Not enough data found for this game.",
			EmbedColor:       utils.InfoEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Determine embed title based on stats type
	var embedTitle string
	switch statsType {
	case "closest-guesses":
		embedTitle = "🎯 Closest Guesses"
	case "top-numbers":
		embedTitle = "🔢 Most Guessed Numbers"
	case "top-guessers":
		embedTitle = "👥 Most Active Guessers"
	case "top-game-winners":
		embedTitle = "🏆 Top Winners"
	case "top-points":
		embedTitle = "💰 Top Point Earners"
	}

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiInfo,
		EmbedTitle:       embedTitle,
		EmbedDescription: content.String(),
		EmbedColor:       utils.InfoEmbedColor,
		WithVote:         true,
	})
}
