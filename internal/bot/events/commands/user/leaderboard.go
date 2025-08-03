package user

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type LeaderboardCommand struct {
	name                   string
	gameService            service.GameService
	guildManagementService service.GuildManagementService
}

func NewLeaderboardCommand(params commands.CommandParams) *LeaderboardCommand {
	return &LeaderboardCommand{
		name:                   "leaderboard",
		gameService:            params.GameService,
		guildManagementService: params.GuildManagementService,
	}
}

func (c *LeaderboardCommand) Name() string {
	return c.name
}

func (c *LeaderboardCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Shows the leaderboard for this server",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "wins",
				Description: "Shows leaderboard sorted by total wins",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "points",
				Description: "Shows leaderboard sorted by total points",
			},
		},
	}
}

func (c *LeaderboardCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName
	guildID := event.GuildID().String()

	switch subcommand {
	case "wins":
		return c.handleLeaderboardWins(ctx, event, guildID)
	case "points":
		return c.handleLeaderboardPoints(ctx, event, guildID)
	}
	return nil
}

func (c *LeaderboardCommand) handleLeaderboardWins(ctx context.Context, event *events.ApplicationCommandInteractionCreate, guildID string) error {
	guildData, err := c.guildManagementService.GetGuildData(ctx, guildID)
	if err != nil {
		return err
	}

	// Convert users map to slice for sorting
	type userEntry struct {
		userID string
		stats  *domain.UserStats
	}

	var users []userEntry
	for userID, stats := range guildData.Users {
		if stats.Wins > 0 { // Only show users with at least one win
			users = append(users, userEntry{userID: userID, stats: stats})
		}
	}

	// Sort by wins (descending)
	sort.Slice(users, func(i, j int) bool {
		return users[i].stats.Wins > users[j].stats.Wins
	})

	var content strings.Builder
	content.WriteString("🏆 **Wins Leaderboard**\n\n")

	if len(users) == 0 {
		content.WriteString("📋 No players with wins yet!\n")
		content.WriteString("🎮 Start playing to appear on the leaderboard!")
	} else {
		content.WriteString("📊 **Top Players by Wins:**\n")
		maxShow := 10
		if len(users) < maxShow {
			maxShow = len(users)
		}

		for i := 0; i < maxShow; i++ {
			user := users[i]
			medal := getMedalEmoji(i + 1)
			content.WriteString(fmt.Sprintf("%s **%d.** <@%s> - %d wins (%d points)\n",
				medal, i+1, user.userID, user.stats.Wins, user.stats.Points))
		}

		if len(users) > maxShow {
			content.WriteString(fmt.Sprintf("\n... and %d more players", len(users)-maxShow))
		}
	}

	content.WriteString(fmt.Sprintf("\n\n📈 **Server Stats:**\n"))
	content.WriteString(fmt.Sprintf("• Total games played: %d\n", guildData.TotalGames))
	content.WriteString(fmt.Sprintf("• Active players: %d", len(guildData.Users)))

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}

func (c *LeaderboardCommand) handleLeaderboardPoints(ctx context.Context, event *events.ApplicationCommandInteractionCreate, guildID string) error {
	guildData, err := c.guildManagementService.GetGuildData(ctx, guildID)
	if err != nil {
		return err
	}

	// Convert users map to slice for sorting
	type userEntry struct {
		userID string
		stats  *domain.UserStats
	}

	var users []userEntry
	for userID, stats := range guildData.Users {
		if stats.Points > 0 { // Only show users with at least one point
			users = append(users, userEntry{userID: userID, stats: stats})
		}
	}

	// Sort by points (descending)
	sort.Slice(users, func(i, j int) bool {
		return users[i].stats.Points > users[j].stats.Points
	})

	var content strings.Builder
	content.WriteString("💎 **Points Leaderboard**\n\n")

	if len(users) == 0 {
		content.WriteString("📋 No players with points yet!\n")
		content.WriteString("🎮 Start playing to appear on the leaderboard!")
	} else {
		content.WriteString("📊 **Top Players by Points:**\n")
		maxShow := 10
		if len(users) < maxShow {
			maxShow = len(users)
		}

		for i := 0; i < maxShow; i++ {
			user := users[i]
			medal := getMedalEmoji(i + 1)
			content.WriteString(fmt.Sprintf("%s **%d.** <@%s> - %d points (%d wins)\n",
				medal, i+1, user.userID, user.stats.Points, user.stats.Wins))
		}

		if len(users) > maxShow {
			content.WriteString(fmt.Sprintf("\n... and %d more players", len(users)-maxShow))
		}
	}

	content.WriteString(fmt.Sprintf("\n\n📈 **Server Stats:**\n"))
	content.WriteString(fmt.Sprintf("• Total games played: %d\n", guildData.TotalGames))
	content.WriteString(fmt.Sprintf("• Active players: %d", len(guildData.Users)))

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}

func getMedalEmoji(position int) string {
	switch position {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return "🏅"
	}
}
