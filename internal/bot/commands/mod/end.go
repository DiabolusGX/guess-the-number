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

type EndCommand struct {
	name        string
	definition  discord.ApplicationCommandCreate
	gameService service.GameService
}

func NewEndCommand(params commands.CommandParams) *EndCommand {
	return &EndCommand{
		name: "end",
		definition: discord.SlashCommandCreate{
			Name:        "end",
			Description: "Ends the current game in a channel",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel to end the game in",
					Required:    true,
				},
			},
		},
		gameService: params.GameService,
	}
}

func (c *EndCommand) Name() string {
	return c.name
}

func (c *EndCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *EndCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	channel := event.SlashCommandInteractionData().Channel("channel")
	game, err := c.gameService.GetGameByChannelID(context.Background(), channel.ID.String())
	if err != nil {
		return err
	}
	if game == nil {
		return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "No Game", "There is no game running in this channel", 0xFF0000)
	}

	if err := c.gameService.FinishGame(context.Background(), game); err != nil {
		return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Error", fmt.Sprintf("failed to end game: %s", err.Error()), 0xFF0000)
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Game Ended", "The game has been ended.", 0x00FF00)
}
