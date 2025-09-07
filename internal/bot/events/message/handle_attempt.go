package message

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commons"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/types"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

func (h *MessageCreateListener) handleAttempt(ctx context.Context, event *disgoEvents.MessageCreate) {
	channel, ok := event.Channel()
	if !ok || channel.Type() != discord.ChannelTypeGuildText {
		return
	}

	content := event.Message.Content
	content = strings.TrimSuffix(strings.TrimPrefix(content, " "), " ")
	if content == "" {
		return
	}

	number, err := strconv.ParseInt(content, 10, 64)
	if err != nil || number <= 0 {
		return
	}

	// Check if user has required role to make guesses
	guildConfig, err := h.GuildManagement.GetGuildConfig(ctx, event.GuildID.String())
	if err != nil {
		h.Logger.FromContext(ctx).Error("failed to get guild config for role check", "error", err.Error())
		return
	}

	// If a required role is set, check if user has it
	if guildConfig != nil && guildConfig.ReqRole != "" {
		requiredRoleID, err := snowflake.Parse(guildConfig.ReqRole)
		if err != nil {
			h.Logger.FromContext(ctx).Error("invalid required role ID in guild config", "roleID", guildConfig.ReqRole, "error", err.Error())
			return
		}

		hasRequiredRole := slices.Contains(event.Message.Member.RoleIDs, requiredRoleID)
		if !hasRequiredRole {
			h.Logger.Debugf("user %s does not have required role %s", event.Message.Author.ID.String(), guildConfig.ReqRole)
			return
		}
	}

	correctGuessLabel := metrics.MetricLabelValueFalse

	defer func() {
		if err != nil {
			h.Logger.FromContext(ctx).Error("failed to handle attempt", "error", err.Error())
			h.Metrics.Guesses.WithLabelValues(correctGuessLabel, metrics.MetricLabelValueTrue).Inc()
			return
		}
		h.Metrics.Guesses.WithLabelValues(correctGuessLabel, metrics.MetricLabelValueFalse).Inc()
	}()

	response, err := h.GameService.HandleAttempt(ctx, &types.HandleAttemptRequest{
		ChannelID: event.ChannelID.String(),
		MessageID: event.Message.ID.String(),
		UserID:    event.Message.Author.ID.String(),
		Guess:     number,
		Timestamp: event.Message.CreatedAt.Unix(),
	})
	if err != nil {
		if appErr, ok := err.(*ierr.AppError); ok && appErr.Code == ierr.ErrCodeNotFound {
			err = nil
			return
		}
		h.Logger.FromContext(ctx).Error("failed to handle attempt in game service", "error", err.Error())
		return
	}

	if !response.Correct {
		// Add auto reaction hints if enabled for this game
		if response.Game.AutoReactionHints {
			var emoji string
			if number < response.Game.Answer {
				emoji = "⬆️" // Guess is too low, answer is higher
			} else {
				emoji = "⬇️" // Guess is too high, answer is lower
			}

			reactionErr := event.Client().Rest.AddReaction(event.ChannelID, event.Message.ID, emoji)
			if reactionErr != nil {
				h.Logger.FromContext(ctx).Error("failed to add reaction hint", "error", reactionErr.Error(), "emoji", emoji)
			}
		}
		return
	}

	// TODO: remove this and make it event based from the game service
	go func() {
		// Use clean context for sync operations to avoid user/guild specific logging
		ctx := lib.CopyServiceContextKeys(ctx)
		ctx = context.WithValue(ctx, lib.CtxGameID, response.Game.ID)
		h.SyncService.SyncGameOnFinish(ctx, response.Game.ID)
	}()

	// handle correct guess
	var winDM, winChannelMsg strings.Builder
	winDM.WriteString(fmt.Sprintf("**Congratulations %s 🎉**\n", event.Message.Author.Mention()))
	winDM.WriteString(fmt.Sprintf("You guessed the correct number **%d** after **%d** guesses at %s\n", response.Game.Answer, response.Game.Guesses, event.Message.JumpURL()))
	winDM.WriteString(fmt.Sprintf("You have also won **%d points!**", response.Game.Points))

	winChannelMsg.WriteString(fmt.Sprintf("**Congratulations %s 🎉**\n", event.Message.Author.Mention()))
	winChannelMsg.WriteString(fmt.Sprintf("You guessed the correct number **%d** after **%d** guesses at %s\n", response.Game.Answer, response.Game.Guesses, event.Message.JumpURL()))
	winChannelMsg.WriteString(fmt.Sprintf("You have also won **%d points!**", response.Game.Points))

	// lock channel
	if guildConfig != nil && guildConfig.LockRole != "" {
		lockRole, ok := event.Client().Caches.Role(*event.GuildID, snowflake.MustParse(guildConfig.LockRole))
		if ok {
			err = utils.LockChannel(ctx, event.Client(), channel, lockRole, "Game completed, locking channel")
			if err != nil {
				return
			}
		}
	}

	// allocate win role
	var winRoleErr error
	var winRole discord.Role
	var winRoleExists bool
	if guildConfig != nil && guildConfig.WinRole != "" {
		winRole, winRoleExists = event.Client().Caches.Role(*event.GuildID, snowflake.MustParse(guildConfig.WinRole))
		if winRoleExists {
			winRoleErr = event.Client().Rest.AddMemberRole(*event.GuildID, event.Message.Author.ID, winRole.ID, rest.WithReason("Game winner, awarding win role"))
			if winRoleErr != nil {
				h.Logger.FromContext(ctx).Error("failed to add win role to winner", "error", winRoleErr.Error())
				winDM.WriteString("\n\n> *Failed to add win role, please contact the admin or bot manager*")
				winChannelMsg.WriteString("\n\n> *Failed to add win role, please contact the admin or bot manager*")
			} else {
				winDM.WriteString(fmt.Sprintf(" and **%s** role!", winRole.Name))
				winChannelMsg.WriteString(fmt.Sprintf(" and **%s** role!", winRole.Mention()))
			}
		}
	}

	// send DM to winner if configured
	var dmErr error
	if guildConfig != nil && guildConfig.DM {
		dmErr = utils.SendDM(event.Client().Rest, event.Message.Author.ID, utils.MessageRequest{
			Emoji:    utils.EmojiSuccess,
			Content:  winDM.String(),
			WithVote: true,
		})
		if dmErr != nil {
			h.Logger.FromContext(ctx).Error("failed to send DM to winner", "error", dmErr.Error())
		}
	}

	// send game completion message and pin it
	_, messageErr := utils.SendMessage(event.Client().Rest, utils.MessageRequest{
		ChannelID:           event.ChannelID,
		Emoji:               utils.EmojiSuccess,
		Content:             winChannelMsg.String(),
		WithVote:            true,
		WithStartGameButton: true,
	})
	if messageErr != nil {
		h.Logger.FromContext(ctx).Error("failed to send game completion message", "error", messageErr.Error())
	}

	// NOTE: not pinning the win message to keep the channel clean
	// TODO: make it configurable
	var pinErr error
	// if messageErr == nil {
	// 	pinErr = event.Client().Rest.PinMessage(event.Message.ChannelID, winMsg.ID, rest.WithReason("Pinning game completion message"))
	// 	if pinErr != nil {
	// 		h.Logger.FromContext(ctx).Error("failed to pin game completion message", "error", pinErr.Error())
	// 	}
	// }

	// un-pin other messages pinned by bot
	var unpinErr error
	var unpinCount, unpinFailures int
	pinnedMessages, unpinErr := event.Client().Rest.GetChannelPins(event.Message.ChannelID, 0, 0)
	if unpinErr != nil {
		h.Logger.FromContext(ctx).Error("failed to get pinned messages", "error", unpinErr.Error())
	} else {
		for _, pinnedMessage := range pinnedMessages.Items {
			if pinnedMessage.Message.Author.ID == event.Client().ID() && !strings.HasPrefix(pinnedMessage.Message.Content, "<:greentick:768464483009691648> | **Congratulations") {
				err := event.Client().Rest.UnpinMessage(event.Message.ChannelID, pinnedMessage.Message.ID, rest.WithReason("Un-pinning other messages pinned by bot"))
				if err != nil {
					h.Logger.FromContext(ctx).Error("failed to un-pin old message after game completion", "error", err.Error())
					unpinFailures++
				} else {
					unpinCount++
				}
			}
		}
	}

	// Log the game completion to the log channel
	h.logGameCompletion(ctx, event, response, guildConfig, winRole, winRoleExists, winRoleErr, dmErr, messageErr, pinErr, unpinErr, unpinCount, unpinFailures)

	// Handle auto restart if enabled
	h.handleAutoRestart(ctx, event, response, guildConfig)
}

func (h *MessageCreateListener) logGameCompletion(ctx context.Context, event *disgoEvents.MessageCreate, response *types.HandleAttemptResponse, guildConfig *domain.GuildConfig, winRole discord.Role, winRoleExists bool, winRoleErr error, dmErr error, messageErr error, pinErr error, unpinErr error, unpinCount int, unpinFailures int) {
	if guildConfig == nil || guildConfig.LogChannel == "" {
		return
	}

	var logContent strings.Builder
	logContent.WriteString(fmt.Sprintf("**Channel:** <#%s>\n", event.ChannelID.String()))
	logContent.WriteString(fmt.Sprintf("**Winner:** <@%s>\n", event.Message.Author.ID.String()))
	logContent.WriteString(fmt.Sprintf("**Answer:** %d\n", response.Game.Answer))
	logContent.WriteString(fmt.Sprintf("**Points Awarded:** %d\n\n", response.Game.Points))

	// Add status information
	logContent.WriteString("**Status Report:**\n")

	// Win role status
	if guildConfig.WinRole != "" {
		if !winRoleExists {
			logContent.WriteString("❌ Win role not found in server cache\n")
		} else if winRoleErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to award %s role: %s\n", winRole.Mention(), winRoleErr.Error()))
		} else {
			logContent.WriteString(fmt.Sprintf("✅ Successfully awarded %s role\n", winRole.Mention()))
		}
	}

	// DM status
	if guildConfig.DM {
		if dmErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to send DM: %s\n", dmErr.Error()))
		} else {
			logContent.WriteString("✅ DM sent successfully\n")
		}
	} else {
		logContent.WriteString("⚪ DM disabled\n")
	}

	// Message sending status
	if messageErr != nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to send completion message: %s\n", messageErr.Error()))
	} else {
		logContent.WriteString("✅ Completion message sent\n")
	}

	// Message pinning status
	if messageErr != nil {
		logContent.WriteString("⚪ Cannot pin - message send failed\n")
	} else if pinErr != nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to pin message: %s\n", pinErr.Error()))
	} else {
		logContent.WriteString("✅ Message pinned successfully\n")
	}

	// Unpinning status
	if unpinErr != nil {
		logContent.WriteString(fmt.Sprintf("❌ Failed to get pinned messages: %s", unpinErr.Error()))
	} else if unpinCount > 0 || unpinFailures > 0 {
		logContent.WriteString(fmt.Sprintf("📌 Unpinned %d old messages", unpinCount))
		if unpinFailures > 0 {
			logContent.WriteString(fmt.Sprintf(", %d failures", unpinFailures))
		}
	} else {
		logContent.WriteString("📌 No old messages to unpin")
	}

	// Add game id to log
	logContent.WriteString(fmt.Sprintf("\n\n**Game ID:** `%s`", response.Game.ID))

	utils.LogToChannel(event.Client().Rest, guildConfig.LogChannel, utils.LogTypeGameActivity, "🏆 Game Won", logContent.String())
}

func (h *MessageCreateListener) handleAutoRestart(ctx context.Context, event *disgoEvents.MessageCreate, response *types.HandleAttemptResponse, guildConfig *domain.GuildConfig) {
	// Check if auto restart is enabled globally and for this specific game
	if guildConfig == nil || !guildConfig.AutoRestart {
		return
	}

	h.Logger.FromContext(ctx).Debugw("auto restarting game")

	// Get the channel where the game was played
	channel, ok := event.Client().Caches.Channel(snowflake.MustParse(response.Game.ChannelID))
	if !ok {
		h.Logger.FromContext(ctx).Error("failed to get channel for auto restart", "channelID", response.Game.ChannelID)
		return
	}

	// fetch discord user from id
	createdBy, err := event.Client().Rest.GetUser(snowflake.MustParse(response.Game.CreatedBy))
	if err != nil {
		h.Logger.FromContext(ctx).Error("failed to get user for auto restart", "userID", response.Game.CreatedBy)
		return
	}

	// Create a new game with the same configuration
	gameStartRequest := &commons.GameStartRequest{
		Client:            event.Client(),
		TargetChannel:     channel,
		CreatedBy:         *createdBy,
		LowerBound:        response.Game.LowerBound,
		UpperBound:        response.Game.UpperBound,
		AutoReactionHints: response.Game.AutoReactionHints,
		AutoRestarting:    true,
		PreviousGameID:    response.Game.ID,
	}

	// Start the new game
	_, err = commons.StartGameCommon(ctx, event.Client().Rest, h.GameService, gameStartRequest, guildConfig)
	if err != nil {
		h.Logger.FromContext(ctx).Error("failed to auto restart game", "error", err.Error(), "channelID", response.Game.ChannelID)

		// Send an error message to the channel
		_, sendErr := utils.SendMessage(event.Client().Rest, utils.MessageRequest{
			ChannelID:   event.ChannelID,
			Emoji:       utils.EmojiError,
			Content:     "**Auto Restart Failed**\nUnable to automatically start a new game. Please start one manually using `/start`.",
			IsEphemeral: false,
		})
		if sendErr != nil {
			h.Logger.FromContext(ctx).Error("failed to send auto restart error message", "error", sendErr.Error())
		}
		return
	}

	h.Logger.FromContext(ctx).Infow("game auto restarted successfully", "old_game_id", response.Game.ID)
}
