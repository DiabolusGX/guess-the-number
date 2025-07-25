package mod

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type StartCommand struct {
	name        string
	definition  discord.ApplicationCommandCreate
	gameService service.GameService
}

func NewStartCommand(params commands.CommandParams) *StartCommand {
	return &StartCommand{
		name: "start",
		definition: discord.SlashCommandCreate{
			Name:        "start",
			Description: "Starts a new game of guess the number",
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
			},
		},
		gameService: params.GameService,
	}
}

func (c *StartCommand) Name() string {
	return c.name
}

func (c *StartCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *StartCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	min := event.SlashCommandInteractionData().Int("min")
	max := event.SlashCommandInteractionData().Int("max")
	channel := event.SlashCommandInteractionData().Channel("channel")

	_, err := c.gameService.CreateGame(context.Background(), event.GuildID().String(), channel.ID.String(), event.User().ID.String(), int64(min), int64(max))
	if err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), channel.ID, "Game Started", fmt.Sprintf("Game started in <#%s> with a number between %d and %d. Good luck!", channel.ID, min, max), 0x00FF00)
}
