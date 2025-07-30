package user

import (
	"context"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
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
	var content strings.Builder

	content.WriteString("🤖 **Guess The Number Bot - Help**\n\n")
	content.WriteString("Welcome to the ultimate number guessing game bot! Here are all the available commands:\n\n")

	// Game Commands
	content.WriteString("🎮 **Game Commands:**\n")
	content.WriteString("• `/start` - Start a new guessing game in the current channel\n")
	content.WriteString("• `/end` - End the current game (moderators only)\n")
	content.WriteString("• `/hint` - Get a hint for the current game (moderators only)\n")
	content.WriteString("• `/gameinfo <channel>` - Show information about a game in a specific channel\n\n")

	// User Commands
	content.WriteString("👥 **User Commands:**\n")
	content.WriteString("• `/leaderboard wins` - Show top players by total wins\n")
	content.WriteString("• `/leaderboard points` - Show top players by total points\n")
	content.WriteString("• `/userinfo [user]` - Show detailed information about a user\n")
	content.WriteString("• `/ping` - Check bot latency and status\n")
	content.WriteString("• `/invite` - Get the bot's invite link\n")
	content.WriteString("• `/help` - Show this help message\n\n")

	// Moderator Commands
	content.WriteString("🛠️ **Moderator Commands:**\n")
	content.WriteString("• `/setup` - Configure bot settings for your server\n")
	content.WriteString("• `/start` - Start games (if required role is set)\n")
	content.WriteString("• `/end` - End active games\n")
	content.WriteString("• `/hint` - Provide hints to players\n\n")

	// How to Play
	content.WriteString("📖 **How to Play:**\n")
	content.WriteString("1️⃣ Use `/start` to begin a new game\n")
	content.WriteString("2️⃣ The bot will think of a number between 1-100\n")
	content.WriteString("3️⃣ Players guess by typing numbers in chat\n")
	content.WriteString("4️⃣ Get feedback: 📈 (higher), 📉 (lower), or 🎉 (correct!)\n")
	content.WriteString("5️⃣ First to guess wins points and possibly a role!\n\n")

	// Features
	content.WriteString("✨ **Features:**\n")
	content.WriteString("• Customizable point rewards\n")
	content.WriteString("• Role rewards for winners\n")
	content.WriteString("• Required roles to play\n")
	content.WriteString("• Server-wide leaderboards\n")
	content.WriteString("• Game statistics tracking\n")
	content.WriteString("• Multiple simultaneous games\n\n")

	// Support
	content.WriteString("🆘 **Need Help?**\n")
	content.WriteString("Join our support server for assistance, updates, and community:\n")
	content.WriteString("🔗 **[DEV Studios](https://discord.gg/8kdx63YsDf)**\n\n")
	content.WriteString("💡 *Tip: Use `/setup` to configure the bot for your server's needs!*")

	return utils.EventReply(event, utils.MessageRequest{
		Content: content.String(),
		Emoji:   utils.EmojiSuccess,
	})
}
