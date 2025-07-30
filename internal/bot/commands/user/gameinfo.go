package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/internal/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type GameInfoCommand struct {
	name                   string
	gameService            service.GameService
	guildManagementService service.GuildManagementService
}

func NewGameInfoCommand(params commands.CommandParams) *GameInfoCommand {
	return &GameInfoCommand{
		name:                   "gameinfo",
		gameService:            params.GameService,
		guildManagementService: params.GuildManagementService,
	}
}

func (c *GameInfoCommand) Name() string {
	return c.name
}

func (c *GameInfoCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Shows information about the current game in a channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to get game info for",
				Required:    true,
			},
		},
	}
}

func (c *GameInfoCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")
	guildID := event.GuildID().String()

	game, err := c.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{ChannelID: channel.ID.String()})
	if err != nil {
		return err
	}

	// Get guild data for statistics
	guildData, err := c.guildManagementService.GetGuildData(ctx, guildID)
	if err != nil {
		return err
	}

	var content strings.Builder

	if game == nil || game.Game == nil {
		// No game running - show prompt to start game and statistics
		content.WriteString(fmt.Sprintf("No game found in <#%s>\n\n", channel.ID.String()))
		content.WriteString("🎮 **Start a new game with `/start` command!**\n\n")
	} else {
		// Game is running - show current game info
		gameInfo := game.Game
		content.WriteString(fmt.Sprintf("🎯 **Game in <#%s>**\n", channel.ID.String()))
		content.WriteString(fmt.Sprintf("• Guesses so far: %d\n", gameInfo.Guesses))
		content.WriteString(fmt.Sprintf("• Points for winner: %d\n\n", gameInfo.Points))
	}

	// Add game configuration details
	guildConfig := data.GuildConfig
	content.WriteString("⚙️ **Game Configuration:**\n")

	// Required role to play
	if guildConfig.ReqRole != "" {
		content.WriteString(fmt.Sprintf("• Required role to play: <@&%s>\n", guildConfig.ReqRole))
	} else {
		content.WriteString("• Required role to play: None (everyone can play)\n")
	}

	// Winner role reward
	if guildConfig.WinRole != "" {
		content.WriteString(fmt.Sprintf("• Winner receives role: <@&%s>\n", guildConfig.WinRole))
	} else {
		content.WriteString("• Winner receives role: None\n")
	}

	// Total games statistics
	content.WriteString(fmt.Sprintf("\n📊 **Statistics:**\n"))
	content.WriteString(fmt.Sprintf("• Total games played: %d", guildData.TotalGames))

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}
