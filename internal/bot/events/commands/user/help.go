package user

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
)

type HelpCommand struct {
	name       string
	definition discord.ApplicationCommandCreate
}

func NewHelpCommand(params commands.CommandParams) *HelpCommand {
	return &HelpCommand{
		name: "help",
		definition: discord.SlashCommandCreate{
			Name:        "help",
			Description: "Shows all available commands and how to use the bot",
		},
	}
}

func (c *HelpCommand) Name() string {
	return c.name
}

func (c *HelpCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *HelpCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	description := "Welcome to the ultimate number guessing game bot! Here are all the available commands:"

	// Game Commands Field
	gameCommands := "• `/game info <channel>` - Show information about a game in a specific channel\n"
	gameCommands += "• `/game hint <type> <channel>` - Get a hint for the current game (moderators only)\n"
	gameCommands += "• `/game answer <channel>` - Reveal the answer for the current game (moderators only)\n"
	gameCommands += "• `/game stats <channel>` - Show game statistics for a channel"

	// User Commands Field
	userCommands := "• `/userinfo [user]` - Show detailed information about a user including game statistics\n"
	userCommands += "• `/ping` - Check bot latency and status\n"
	userCommands += "• `/invite` - Get the bot's invite link to add it to other servers\n"
	userCommands += "• `/help` - Show this help message"

	// Moderator Commands Field
	moderatorCommands := "• `/setup` - Configure bot settings for your server (prefix, roles, channels, etc.)\n"
	moderatorCommands += "• `/start <min> <max> <channel>` - Start a new game with custom range and channel\n"
	moderatorCommands += "• `/end <channel>` - End the running game in a channel\n"
	moderatorCommands += "• `/finish-game <channel>` - Finish the running game in a channel"

	// How to Play Field
	howToPlay := fmt.Sprintf("1️⃣ Moderators use %s to begin a new game\n", utils.MentionApplicationCommand(event.Client().ID(), utils.CommandStart))
	howToPlay += "2️⃣ The bot will think of a number within the specified range\n"
	howToPlay += "3️⃣ Players guess by typing numbers in the game channel\n"
	howToPlay += "4️⃣ Get feedback (if enabled): 📈 (higher), 📉 (lower), or 🎉 (correct!)\n"
	howToPlay += "5️⃣ First to guess correctly wins points and possibly a role!"
	howToPlay += "\n\n💡 *Tip: Use `/setup` to configure the bot for your server's needs!*"

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		EmbedTitle:       "🤖 Guess The Number Bot - Help",
		EmbedDescription: description,
		EmbedColor:       utils.InfoEmbedColor,
		Fields: []discord.EmbedField{
			{
				Name:   "🎮 Game Commands",
				Value:  gameCommands,
				Inline: omit.NewPtr(false).Value,
			},
			{
				Name:   "👥 User Commands",
				Value:  userCommands,
				Inline: omit.NewPtr(false).Value,
			},
			{
				Name:   "🛠️ Moderator Commands",
				Value:  moderatorCommands,
				Inline: omit.NewPtr(false).Value,
			},
			{
				Name:   "📖 How to Play",
				Value:  howToPlay,
				Inline: omit.NewPtr(false).Value,
			},
		},
		WithVote:          true,
		WithSupportServer: true,
	})
}
