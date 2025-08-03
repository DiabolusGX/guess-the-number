package commons

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/internal/types"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// GameStartRequest contains the parameters for starting a new game
type GameStartRequest struct {
	Client            bot.Client
	TargetChannel     discord.GuildChannel
	CreatedBy         discord.User
	LowerBound        int64
	UpperBound        int64
	AutoReactionHints bool
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
	})
	if err != nil {
		return nil, err
	}

	// unlock target channel from lock role if it's locked
	var lockRoleErr error
	if guildConfig != nil && guildConfig.LockRole != "" {
		lockRole, ok := request.Client.Caches().Role(guildID, snowflake.MustParse(guildConfig.LockRole))
		if ok {
			lockRoleErr = utils.UnlockChannel(ctx, request.Client, request.TargetChannel, lockRole, "New game started, unlocking channel")
		}
	}

	// send game start message in target channel
	var gameStartMsg strings.Builder
	gameStartMsg.WriteString("**Game Started**\n")
	gameStartMsg.WriteString(fmt.Sprintf("Guess a number between `%d` and `%d`\n", request.LowerBound, request.UpperBound))
	gameStartMsg.WriteString(fmt.Sprintf("Correct guess will get you **%d points**", game.Game.Points))
	if guildConfig != nil && guildConfig.WinRole != "" {
		winRole, ok := request.Client.Caches().Role(guildID, snowflake.MustParse(guildConfig.WinRole))
		if ok {
			gameStartMsg.WriteString(fmt.Sprintf(" and %s role.", winRole.Mention()))
		}
	}
	msg, gameStartMsgErr := utils.SendMessage(request.Client.Rest(), utils.MessageRequest{
		ChannelID: request.TargetChannel.ID(),
		Content:   gameStartMsg.String(),
		Emoji:     utils.EmojiSuccess,
	})

	// pin game start message in target channel
	pinErr := request.Client.Rest().PinMessage(request.TargetChannel.ID(), msg.ID, rest.WithReason("Pinning game start message"))

	// send game start message with answer to user's DM
	dmErr := utils.SendDM(request.Client.Rest(), request.CreatedBy.ID, utils.MessageRequest{
		Content: fmt.Sprintf("Game started!\nRandom answer ||%d|| has been set. Start guessing here: %s", game.Game.Answer, msg.JumpURL()),
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
	logContent.WriteString(fmt.Sprintf("**Started by:** %s\n\n", request.CreatedBy.Mention()))

	// Add status information
	logContent.WriteString("**Status Report:**\n")

	// Channel unlock status
	if guildConfig != nil && guildConfig.LockRole != "" {
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

	utils.LogToChannel(client, logChannelID, utils.LogTypeGameActivity, "🎮 Game Started", logContent.String())
}

// FormatGameStartReply formats the reply message for the user who started the game
func FormatGameStartReply(
	result *GameStartResult,
	targetChannel discord.GuildChannel,
	guildConfig *domain.GuildConfig,
) string {
	var reply strings.Builder
	reply.WriteString("**Game Started**\n")
	reply.WriteString("Sending the game's answer to your DM.")

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
