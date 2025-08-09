package user

import (
	"context"

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
	description := "Welcome to the ultimate number guessing game bot! Here are all the available commands and features:"

	// Game Commands Field
	gameCommands := "• `/start` - Start a new guessing game in the current channel\n"
	gameCommands += "• `/end` - End the current game (moderators only)\n"
	gameCommands += "• `/hint` - Get a hint for the current game (moderators only)\n"
	gameCommands += "• `/gameinfo <channel>` - Show information about a game in a specific channel"

	// User Commands Field
	userCommands := "• `/leaderboard wins` - Show top players by total wins\n"
	userCommands += "• `/leaderboard points` - Show top players by total points\n"
	userCommands += "• `/userinfo [user]` - Show detailed information about a user\n"
	userCommands += "• `/ping` - Check bot latency and status\n"
	userCommands += "• `/invite` - Get the bot's invite link\n"
	userCommands += "• `/help` - Show this help message"

	// Moderator Commands Field
	moderatorCommands := "• `/setup` - Configure bot settings for your server\n"
	moderatorCommands += "• `/start` - Start games (if required role is set)\n"
	moderatorCommands += "• `/end` - End active games\n"
	moderatorCommands += "• `/hint` - Provide hints to players"

	// How to Play Field
	howToPlay := "1️⃣ Use `/start` to begin a new game\n"
	howToPlay += "2️⃣ The bot will think of a number between 1-100\n"
	howToPlay += "3️⃣ Players guess by typing numbers in chat\n"
	howToPlay += "4️⃣ Get feedback: 📈 (higher), 📉 (lower), or 🎉 (correct!)\n"
	howToPlay += "5️⃣ First to guess wins points and possibly a role!"

	// Features Field
	features := "• Customizable point rewards\n"
	features += "• Role rewards for winners\n"
	features += "• Required roles to play\n"
	features += "• Server-wide leaderboards\n"
	features += "• Game statistics tracking\n"
	features += "• Multiple simultaneous games"

	// Support Field
	support := "Join our support server for assistance, updates, and community:\n"
	support += "🔗 **[DEV Studios](https://discord.gg/8kdx63YsDf)**\n\n"
	support += "💡 *Tip: Use `/setup` to configure the bot for your server's needs!*"

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
			{
				Name:   "✨ Features",
				Value:  features,
				Inline: omit.NewPtr(false).Value,
			},
			{
				Name:   "🆘 Need Help?",
				Value:  support,
				Inline: omit.NewPtr(false).Value,
			},
		},
		WithVote:          true,
		WithSupportServer: true,
	})
}
