package user

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type GameInfoCommand struct {
	name        string
	definition  discord.ApplicationCommandCreate
	gameService service.GameService
}

func NewGameInfoCommand(params commands.CommandParams) *GameInfoCommand {
	return &GameInfoCommand{
		name: "gameinfo",
		definition: discord.SlashCommandCreate{
			Name:        "gameinfo",
			Description: "Shows information about the current game in a channel",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel to get game info for",
					Required:    true,
				},
			},
		},
		gameService: params.GameService,
	}
}

func (c *GameInfoCommand) Name() string {
	return c.name
}

func (c *GameInfoCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *GameInfoCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	channel := event.SlashCommandInteractionData().Channel("channel")
	game, err := c.gameService.GetGameByChannelID(context.Background(), channel.ID.String())
	if err != nil {
		return err
	}
	if game == nil {
		return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "No Game", "There is no game running in this channel", 0xFF0000)
	}

	return utils.SendEmbed(event.Client().Rest(), channel.ID, "Game Info", fmt.Sprintf("Game in <#%s> has %d guesses so far. The winner will get %d points.", channel.ID, game.Guesses, game.Points), 0x00FF00)
}
