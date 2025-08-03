package mod

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/commands"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type EndCommand struct {
	name          string
	finishCommand *FinishCommand
}

func NewEndCommand(params commands.CommandParams) *EndCommand {
	return &EndCommand{
		name:          "end",
		finishCommand: NewFinishCommand(params),
	}
}

func (c *EndCommand) Name() string {
	return c.name
}

func (c *EndCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Ends the running game in a channel (alias for finish-game)",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to end the game in",
				Required:    true,
			},
		},
	}
}

func (c *EndCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	// Delegate to the finish command handler
	return c.finishCommand.Handler(ctx, event, data)
}
