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
)

type HintCommand struct {
	name        string
	gameService service.GameService
}

func NewHintCommand(params commands.CommandParams) *HintCommand {
	return &HintCommand{
		name:        "hint",
		gameService: params.GameService,
	}
}

func (c *HintCommand) Name() string {
	return c.name
}

func (c *HintCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Get a hint for the current game",
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
	}
}

func (c *HintCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
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
		} else {
			logContent.WriteString("✅ Hint message sent\n")
		}

		// Message pinning status
		if messageErr != nil {
			logContent.WriteString("⚪ Cannot pin - message send failed")
		} else if pinErr != nil {
			logContent.WriteString(fmt.Sprintf("❌ Failed to pin hint message: %s", pinErr.Error()))
		} else {
			logContent.WriteString("✅ Hint message pinned")
		}

		utils.LogToChannel(event.Client().Rest(), data.GuildConfig.LogChannel, utils.LogTypeGameActivity, "💡 Hint Given", logContent.String())
	}

	return utils.EventReply(event, utils.MessageRequest{
		Content:     userResponse,
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}
