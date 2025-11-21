package commons

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/internal/types"
	"github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// GameStartRequest contains the parameters for starting a new game
type GameStartRequest struct {
	Client            *bot.Client
	TargetChannel     discord.GuildChannel
	CreatedBy         discord.User
	LowerBound        int64
	UpperBound        int64
	AutoReactionHints bool
	AutoRestarting    bool
	PreviousGameID    string
}

// GameStartResult contains the result of starting a game
type GameStartResult struct {
	Game              *types.CreateGameResponse
	GameStartMessage  *discord.Message
	UnlockError       error
	GameStartMsgError error
	PinError          error
	DMError           error
}

// StartGameCommon handles the common game starting logic used by both commands and modal interactions
func StartGameCommon(
	ctx context.Context,
	client rest.Rest,
	gameService service.GameService,
	request *GameStartRequest,
	guildConfig *domain.GuildConfig,
) (*GameStartResult, error) {
	guildID := request.TargetChannel.GuildID()

	// Create the game
	game, err := gameService.CreateGame(ctx, &types.CreateGameRequest{
		GuildID:           guildID.String(),
		ChannelID:         request.TargetChannel.ID().String(),
		CreatedBy:         request.CreatedBy.ID.String(),
		LowerBound:        request.LowerBound,
		UpperBound:        request.UpperBound,
		AutoReactionHints: request.AutoReactionHints,
		AutoRestarting:    request.AutoRestarting,
		PreviousGameID:    request.PreviousGameID,
	})
	if err != nil {
		return nil, err
	}
	if game == nil || game.Game == nil {
		return nil, errors.New(errors.ErrCodeInternalError, "Failed to create game. Please try again later.")
	}

	// unlock target channel from lock role if it's locked
	var lockRoleErr error
	if guildConfig != nil && guildConfig.LockRole != "" {
		lockRole, ok := request.Client.Caches.Role(guildID, snowflake.MustParse(guildConfig.LockRole))
		if ok {
			unlockReason := "New game started, unlocking channel"
			if request.AutoRestarting {
				unlockReason = "Game auto-restarted, unlocking channel"
			}
			lockRoleErr = utils.UnlockChannel(ctx, request.Client, request.TargetChannel, lockRole, unlockReason)
		}
	}

	// send game start message in target channel
	var gameStartMsg strings.Builder
	if request.AutoRestarting {
		gameStartMsg.WriteString("**Game Auto Restarted**\n")
	} else {
		gameStartMsg.WriteString("**Game Started**\n")
	}
	gameStartMsg.WriteString(fmt.Sprintf("Guess a number between `%d` and `%d`\n", request.LowerBound, request.UpperBound))
	gameStartMsg.WriteString(fmt.Sprintf("Correct guess will get you **%d points**", game.Game.Points))
	if guildConfig != nil && guildConfig.WinRole != "" {
		winRole, ok := request.Client.Caches.Role(guildID, snowflake.MustParse(guildConfig.WinRole))
		if ok {
			gameStartMsg.WriteString(fmt.Sprintf(" and %s role.", winRole.Mention()))
		}
	}
	gameStartMsg.WriteString(fmt.Sprintf("\n\n*Game ID: `%s`*\n", game.Game.ID))
	if request.PreviousGameID != "" {
		gameStartMsg.WriteString(fmt.Sprintf("*Previous Game ID: `%s`*\n", request.PreviousGameID))
	}
	msg, gameStartMsgErr := utils.SendMessage(request.Client.Rest, utils.MessageRequest{
		ChannelID: request.TargetChannel.ID(),
		Content:   gameStartMsg.String(),
		Emoji:     utils.EmojiSuccess,
	})

	// pin game start message in target channel
	var pinErr error
	if msg != nil {
		pinReason := "Pinning game start message"
		if request.AutoRestarting {
			pinReason = "Pinning auto-restarted game start message"
		}
		pinErr = request.Client.Rest.PinMessage(request.TargetChannel.ID(), msg.ID, rest.WithReason(pinReason))
	}

	// send game start message with answer to user's DM
	var dmContent strings.Builder
	if request.AutoRestarting {
		dmContent.WriteString("**Game Auto Restarted**\n")
	} else {
		dmContent.WriteString("Game started!\n")
	}
	dmContent.WriteString(fmt.Sprintf("Random answer ||%d|| has been set. Start guessing here: %s\n\nGame ID: `%s`", game.Game.Answer, msg.JumpURL(), game.Game.ID))
	if request.PreviousGameID != "" {
		dmContent.WriteString(fmt.Sprintf("\nPrevious game (`%s`) has ended, and a new one has automatically started.\n", request.PreviousGameID))
	}

	dmErr := utils.SendDM(request.Client.Rest, request.CreatedBy.ID, utils.MessageRequest{
		Content: dmContent.String(),
		Emoji:   utils.EmojiSuccess,
	})

	result := &GameStartResult{
		Game:              game,
		GameStartMessage:  msg,
		GameStartMsgError: gameStartMsgErr,
		UnlockError:       lockRoleErr,
		PinError:          pinErr,
		DMError:           dmErr,
	}

	// Log the game start to the log channel
	logGameStart(client, request, game, result, guildConfig)

	return result, nil
}

// logGameStart logs the game start to the configured log channel
func logGameStart(
	client rest.Rest,
	request *GameStartRequest,
	game *types.CreateGameResponse,
	result *GameStartResult,
	guildConfig *domain.GuildConfig,
) {
	if request == nil || guildConfig == nil || guildConfig.LogChannel == "" {
		return
	}
	logChannelID := guildConfig.LogChannel

	var logContent strings.Builder
	logContent.WriteString(fmt.Sprintf("**Channel:** %s\n", request.TargetChannel.Mention()))
	logContent.WriteString(fmt.Sprintf("**Range:** %d - %d\n", request.LowerBound, request.UpperBound))
	logContent.WriteString(fmt.Sprintf("**Points:** %d\n", game.Game.Points))
	if request.AutoRestarting && request.PreviousGameID != "" {
		logContent.WriteString(fmt.Sprintf("**Previous Game ID:** `%s`\n", request.PreviousGameID))
		logContent.WriteString(fmt.Sprintf("**Auto Restarted by:** %s\n\n", request.CreatedBy.Mention()))
	} else {
		logContent.WriteString(fmt.Sprintf("**Started by:** %s\n\n", request.CreatedBy.Mention()))
	}

	// Add status information
	logContent.WriteString("**Status Report:**\n")

	// Channel unlock status
	if guildConfig.LockRole != "" {
		if result.UnlockError != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to unlock channel: %s\n", result.UnlockError.Error()))
		} else {
			logContent.WriteString("✅ Channel unlocked successfully\n")
		}
	} else {
		logContent.WriteString("⚪ No lock role configured\n")
	}

	// Message sending status
	if result.GameStartMsgError != nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to send start message: %s\n", result.GameStartMsgError.Error()))
	} else {
		logContent.WriteString("✅ Start message sent\n")
	}

	// Message pinning status
	if result.PinError != nil && result.GameStartMsgError == nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to pin start message: %s\n", result.PinError.Error()))
	} else if result.GameStartMsgError == nil {
		logContent.WriteString("✅ Start message pinned\n")
	}

	// DM status
	if result.DMError != nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to send DM with answer: %s", result.DMError.Error()))
	} else {
		logContent.WriteString("✅ DM with answer sent successfully")
	}

	// Add game id to log
	logContent.WriteString(fmt.Sprintf("\n\n**Game ID:** `%s`", result.Game.Game.ID))

	// Use different log title based on auto restart status
	logTitle := "🎮 Game Started"
	if request.AutoRestarting {
		logTitle = "🔄 Game Auto Restarted"
	}

	utils.LogToChannel(client, logChannelID, utils.LogTypeGameActivity, logTitle, logContent.String())
}

// FormatGameStartReply formats the reply message for the user who started the game
func FormatGameStartReply(
	result *GameStartResult,
	targetChannel discord.GuildChannel,
	guildConfig *domain.GuildConfig,
) string {
	var reply strings.Builder
	if result.Game.Game.ID != "" {
		// Check if this is an auto-restarted game
		if result.Game.Game.AutoRestarted {
			reply.WriteString(fmt.Sprintf("🔄 **Game Auto Restarted** (`%s`)\n", result.Game.Game.ID))
			if result.Game.Game.PreviousGameID != "" {
				reply.WriteString(fmt.Sprintf("*Previous game: `%s`*\n", result.Game.Game.PreviousGameID))
			}
		} else {
			reply.WriteString(fmt.Sprintf("**Game Started** (`%s`)\n", result.Game.Game.ID))
		}
	}
	if result.DMError == nil {
		reply.WriteString("Sent the game's answer to your DM.")
	}

	// Add unlock status if applicable
	if guildConfig != nil && guildConfig.LockRole != "" {
		if result.UnlockError != nil {
			reply.WriteString(fmt.Sprintf("\n> *Unable to unlock channel %s. Please unlock it manually so everyone can start guessing.*", targetChannel.Mention()))
		}
	}

	// Add game start message status
	if result.GameStartMsgError != nil {
		reply.WriteString(fmt.Sprintf("\n> Unable to send message in %s! Please check if bot has permissions to send messages in this channel.", targetChannel.Mention()))
	} else if result.GameStartMessage != nil {
		reply.WriteString(fmt.Sprintf("\nStart guessing: %s", result.GameStartMessage.JumpURL()))
	}

	// Add pin status if applicable
	if result.PinError != nil && result.GameStartMsgError == nil && result.GameStartMessage != nil {
		reply.WriteString(fmt.Sprintf("\n> Unable to pin game start message %s. See if limit of 50 pinned messages is reached. Please pin it manually so everyone would know Min and Max values.", result.GameStartMessage.JumpURL()))
	}

	// Add DM status
	if result.DMError != nil {
		reply.WriteString("\n> **Unable to send DM!** Please use the `/answer` command to receive the game's answer via DM.")
	}

	return reply.String()
}
