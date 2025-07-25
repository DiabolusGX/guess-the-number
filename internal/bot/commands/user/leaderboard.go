package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type LeaderboardCommand struct {
	name        string
	definition  discord.ApplicationCommandCreate
	gameService service.GameService
}

func NewLeaderboardCommand(params commands.CommandParams) *LeaderboardCommand {
	return &LeaderboardCommand{
		name: "leaderboard",
		definition: discord.SlashCommandCreate{
			Name:        "leaderboard",
			Description: "Shows the leaderboard",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "daily",
					Description: "Shows the daily leaderboard",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "weekly",
					Description: "Shows the weekly leaderboard",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "all",
					Description: "Shows the all-time leaderboard",
				},
			},
		},
		gameService: params.GameService,
	}
}

func (c *LeaderboardCommand) Name() string {
	return c.name
}

func (c *LeaderboardCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *LeaderboardCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName
	switch subcommand {
	case "daily":
		return handleLeaderboardDaily(event, c.gameService)
	case "weekly":
		return handleLeaderboardWeekly(event, c.gameService)
	case "all":
		return handleLeaderboardAll(event, c.gameService)
	}
	return nil
}

func handleLeaderboardDaily(event *events.ApplicationCommandInteractionCreate, service service.GameService) error {
	games, err := service.GetDailyLeaderboard(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}
	if len(games) == 0 {
		return event.CreateMessage(discord.MessageCreate{
			Content: "There are no winners yet today",
		})
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Daily Leaderboard", formatLeaderboard(games), 0x00FF00)
}

func handleLeaderboardWeekly(event *events.ApplicationCommandInteractionCreate, service service.GameService) error {
	games, err := service.GetWeeklyLeaderboard(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}
	if len(games) == 0 {
		return event.CreateMessage(discord.MessageCreate{
			Content: "There are no winners yet this week",
		})
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Weekly Leaderboard", formatLeaderboard(games), 0x00FF00)
}

func handleLeaderboardAll(event *events.ApplicationCommandInteractionCreate, service service.GameService) error {
	games, err := service.GetAllTimeLeaderboard(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}
	if len(games) == 0 {
		return event.CreateMessage(discord.MessageCreate{
			Content: "There are no winners yet",
		})
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "All-Time Leaderboard", formatLeaderboard(games), 0x00FF00)
}

func formatLeaderboard(games []domain.Game) string {
	var builder strings.Builder
	builder.WriteString("Leaderboard:\n")
	for i, game := range games {
		builder.WriteString(fmt.Sprintf("%d. <@%s> - %d points\n", i+1, game.WonBy, game.Points))
	}
	return builder.String()
}
