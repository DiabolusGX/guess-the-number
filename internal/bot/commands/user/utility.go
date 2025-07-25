package user

import (
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type PingCommand struct {
	name       string
	definition discord.ApplicationCommandCreate
}

func NewPingCommand(params commands.CommandParams) *PingCommand {
	return &PingCommand{
		name: "ping",
		definition: discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Pings the bot",
		},
	}
}

func (c *PingCommand) Name() string {
	return c.name
}

func (c *PingCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *PingCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	latency := time.Since(event.ID().Time())
	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Pong!", fmt.Sprintf("Latency: %s", latency), 0x00FF00)
}

type UserinfoCommand struct {
	name       string
	definition discord.ApplicationCommandCreate
}

func NewUserinfoCommand(params commands.CommandParams) *UserinfoCommand {
	return &UserinfoCommand{
		name: "userinfo",
		definition: discord.SlashCommandCreate{
			Name:        "userinfo",
			Description: "Shows information about a user",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user to get info for",
					Required:    false,
				},
			},
		},
	}
}

func (c *UserinfoCommand) Name() string {
	return c.name
}

func (c *UserinfoCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *UserinfoCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	var user discord.User
	if optUser, ok := event.SlashCommandInteractionData().OptUser("user"); ok {
		user = optUser
	} else {
		user = event.User()
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "User Info", fmt.Sprintf("User: %s\nID: %s", user.Username, user.ID), 0x00FF00)
}

type InviteCommand struct {
	name       string
	definition discord.ApplicationCommandCreate
}

func NewInviteCommand(params commands.CommandParams) *InviteCommand {
	return &InviteCommand{
		name: "invite",
		definition: discord.SlashCommandCreate{
			Name:        "invite",
			Description: "Get the bot's invite link",
		},
	}
}

func (c *InviteCommand) Name() string {
	return c.name
}

func (c *InviteCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *InviteCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", event.Client().ApplicationID())
	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Invite", inviteURL, 0x00FF00)
}
