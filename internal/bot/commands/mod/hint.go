package mod

import (
	"context"
	"strconv"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type HintCommand struct {
	name        string
	definition  discord.ApplicationCommandCreate
	gameService service.GameService
}

func NewHintCommand(params commands.CommandParams) *HintCommand {
	return &HintCommand{
		name: "hint",
		definition: discord.SlashCommandCreate{
			Name:        "hint",
			Description: "Get a hint for the current game",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "The type of hint to get",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "first",
							Value: "first",
						},
						{
							Name:  "last",
							Value: "last",
						},
					},
				},
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel to get the hint for",
					Required:    true,
				},
			},
		},
		gameService: params.GameService,
	}
}

func (c *HintCommand) Name() string {
	return c.name
}

func (c *HintCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *HintCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	hintType := event.SlashCommandInteractionData().String("type")
	channel := event.SlashCommandInteractionData().Channel("channel")

	game, err := c.gameService.GetGameByChannelID(context.Background(), channel.ID.String())
	if err != nil {
		return err
	}
	if game == nil {
		return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "No Game", "There is no game running in this channel", 0xFF0000)
	}

	switch hintType {
	case "first":
		return c.handleFirstDigitHint(event, game)
	case "last":
		return c.handleLastDigitHint(event, game)
	}
	return nil
}

func (c *HintCommand) handleFirstDigitHint(event *events.ApplicationCommandInteractionCreate, game *domain.Game) error {
	answerStr := strconv.FormatInt(game.Answer, 10)
	hint := string(answerStr[0])
	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Hint", "The first digit of the answer is "+hint, 0x00FF00)
}

func (c *HintCommand) handleLastDigitHint(event *events.ApplicationCommandInteractionCreate, game *domain.Game) error {
	answerStr := strconv.FormatInt(game.Answer, 10)
	hint := string(answerStr[len(answerStr)-1])
	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Hint", "The last digit of the answer is "+hint, 0x00FF00)
}
