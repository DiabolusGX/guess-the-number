package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
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
			Description: "Pings the bot and shows latency information",
		},
	}
}

func (c *PingCommand) Name() string {
	return c.name
}

func (c *PingCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *PingCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	latency := time.Since(event.ID().Time())

	var description strings.Builder
	description.WriteString(fmt.Sprintf("• Response time: %s\n", latency))
	description.WriteString("• Status: Online and ready!\n")
	description.WriteString("• Ready to start games! 🎮")

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		EmbedTitle:       "🏓 Pong!",
		EmbedDescription: description.String(),
		EmbedColor:       utils.SuccessEmbedColor,
	})
}

type UserinfoCommand struct {
	name                   string
	definition             discord.ApplicationCommandCreate
	guildManagementService service.GuildManagementService
}

func NewUserinfoCommand(params commands.CommandParams) *UserinfoCommand {
	return &UserinfoCommand{
		name: "userinfo",
		definition: discord.SlashCommandCreate{
			Name:        "userinfo",
			Description: "Shows detailed information about a user including game statistics",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user to get info for (defaults to yourself)",
					Required:    false,
				},
			},
		},
		guildManagementService: params.GuildManagementService,
	}
}

func (c *UserinfoCommand) Name() string {
	return c.name
}

func (c *UserinfoCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *UserinfoCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	var user discord.User
	if optUser, ok := event.SlashCommandInteractionData().OptUser("user"); ok {
		user = optUser
	} else {
		user = event.User()
	}

	guildID := event.GuildID().String()

	// Get guild data for user statistics
	guildData, err := c.guildManagementService.GetGuildData(ctx, guildID)
	if err != nil {
		return err
	}

	// Get user stats (default to 0 if user not found)
	userStats := &domain.UserStats{Wins: 0, Points: 0}
	if stats, exists := guildData.Users[user.ID.String()]; exists {
		userStats = stats
	}

	var description strings.Builder
	description.WriteString(fmt.Sprintf("**User ID:** `%s`\n", user.ID))
	description.WriteString(fmt.Sprintf("**Mention:** <@%s>", user.ID))

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		EmbedTitle:       fmt.Sprintf("👤 %s", user.Username),
		EmbedDescription: description.String(),
		EmbedColor:       utils.InfoEmbedColor,
		Fields: []discord.EmbedField{
			{
				Name:   "🎮 Game Statistics",
				Value:  fmt.Sprintf("• Total Wins: **%d** 🎉\n• Total Points: **%d** ⚖️", userStats.Wins, userStats.Points),
				Inline: omit.NewPtr(false).Value,
			},
		},
		WithVote: true,
	})
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
			Description: "Get the bot's invite link to add it to other servers",
		},
	}
}

func (c *InviteCommand) Name() string {
	return c.name
}

func (c *InviteCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *InviteCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	inviteURL := utils.GetBotInviteLink(event.Client().ApplicationID())

	var description strings.Builder
	description.WriteString(fmt.Sprintf("🔗 **[Click here to invite the bot](%s)**\n\n", inviteURL))
	description.WriteString("✨ **What you'll get:**\n")
	description.WriteString("• Fun number guessing games\n")
	description.WriteString("• Leaderboards and statistics\n")
	description.WriteString("• Customizable game settings\n")
	description.WriteString("• Role rewards for winners\n")
	description.WriteString("• And much more! 🎮")

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:            true,
		EmbedTitle:          "🤖 Add Guess The Number Bot to Your Server!",
		EmbedDescription:    description.String(),
		EmbedColor:          utils.SuccessEmbedColor,
		WithSupportServer:   true,
		WithBotInviteButton: true,
	})
}
