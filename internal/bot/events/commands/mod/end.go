package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/internal/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type FinishCommand struct {
	name        string
	gameService service.GameService
}

func NewFinishCommand(params commands.CommandParams) *FinishCommand {
	return &FinishCommand{
		name:        "finish-game",
		gameService: params.GameService,
	}
}

func (c *FinishCommand) Name() string {
	return c.name
}

func (c *FinishCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Finishs the running game in a channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to finish the game in",
				Required:    true,
			},
		},
	}
}

func (c *FinishCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")

	game, err := c.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{
		ChannelID: channel.ID.String(),
	})
	if err != nil || game == nil {
		err = utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("No game found in <#%s>", channel.ID.String()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
		if err != nil {
			utils.HandleError(ctx, event, err)
			return nil
		}
		return nil
	}

	_, err = c.gameService.FinishGame(ctx, &types.FinishGameRequest{
		ChannelID: game.Game.ChannelID,
	})
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("Game finished in <#%s>\n\n", channel.ID.String()))
	response.WriteString(fmt.Sprintf("Use %s command with Game ID: `%s` to see the stats.\n", utils.MentionApplicationCommand(event.Client().ID(), utils.CommandGameStats), game.Game.ID))

	// TODO: add some game stats
	_, err = utils.SendMessage(event.Client().Rest(), utils.MessageRequest{
		ChannelID: channel.ID,
		Content:   fmt.Sprintf("Game has beed ended by <@%s>", event.User().ID.String()),
		Emoji:     utils.EmojiSuccess,
	})
	var gameFinishedMessageStatus string
	if err != nil {
		response.WriteString(fmt.Sprintf("> *Failed to send 'Game Finished' message in <#%s> please verify permissions.*", channel.ID.String()))
		gameFinishedMessageStatus = fmt.Sprintf("❌ Failed to send end message: %s", err.Error())
	} else {
		gameFinishedMessageStatus = "✅ End message sent successfully"
	}

	// Log the game finish to the log channel
	if data.GuildConfig != nil && data.GuildConfig.LogChannel != "" {
		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("**Channel:** <#%s>\n", channel.ID.String()))
		logContent.WriteString(fmt.Sprintf("**Answer:** %d\n", game.Game.Answer))
		logContent.WriteString(fmt.Sprintf("**Total Guesses:** %d\n", game.Game.Guesses))
		logContent.WriteString(fmt.Sprintf("**Force Ended by:** <@%s>\n\n", event.User().ID.String()))

		// Add status information
		logContent.WriteString("**Status Report:**\n")
		logContent.WriteString(gameFinishedMessageStatus)

		// Add game id to log
		logContent.WriteString(fmt.Sprintf("\n\n**Game ID:** `%s`", game.Game.ID))

		utils.LogToChannel(event.Client().Rest(), data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "🔴 Game Force Ended", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		Content:     response.String(),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}
