package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
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

	var content strings.Builder
	content.WriteString("🏓 **Pong!**\n\n")
	content.WriteString("📊 **Connection Info:**\n")
	content.WriteString(fmt.Sprintf("• Response time: %s\n", latency))
	content.WriteString("• Status: Online and ready!\n")
	content.WriteString("• Ready to start games! 🎮")

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
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

	var content strings.Builder
	content.WriteString(fmt.Sprintf("👤 **%s** (`%s`)\n\n", user.Username, user.ID))

	// Game Statistics
	content.WriteString("🎮 **Game Statistics:**\n")
	content.WriteString(fmt.Sprintf("• Total Wins: **%d** 🎉\n", userStats.Wins))
	content.WriteString(fmt.Sprintf("• Total Points: **%d** ⚖️", userStats.Points))

	return utils.EventReply(event, utils.MessageRequest{
		Content:  content.String(),
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
	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", event.Client().ApplicationID())

	var content strings.Builder
	content.WriteString("🤖 **Add Guess The Number Bot to Your Server!**\n\n")
	content.WriteString("🔗 **Invite Link:**\n")
	content.WriteString(fmt.Sprintf("[Click here to invite the bot](%s)\n\n", inviteURL))
	content.WriteString("✨ **What you'll get:**\n")
	content.WriteString("• Fun number guessing games\n")
	content.WriteString("• Leaderboards and statistics\n")
	content.WriteString("• Customizable game settings\n")
	content.WriteString("• Role rewards for winners\n")
	content.WriteString("• And much more! 🎮")

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}
