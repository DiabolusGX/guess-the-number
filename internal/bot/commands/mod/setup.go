package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type SetupCommand struct {
	name                   string
	guildManagementService service.GuildManagementService
}

func NewSetupCommand(params commands.CommandParams) *SetupCommand {
	return &SetupCommand{
		name:                   "setup",
		guildManagementService: params.GuildManagementService,
	}
}

func (c *SetupCommand) Name() string {
	return c.name
}

func (c *SetupCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Setup the bot for your server",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "prefix",
				Description: "Set the prefix for the bot",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "prefix",
						Description: "The new prefix",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "manager",
				Description: "Set the manager role for the bot",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The manager role",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "dm",
				Description: "Enable or disable DMs to the winner of the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionBool{
						Name:        "enabled",
						Description: "Whether to enable DMs to the winner of the game",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "win-role",
				Description: "Set the role to be awarded to the winner of the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role to award",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "req-role",
				Description: "Set the role required to play the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role required to play",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "lock-role",
				Description: "Set the role to lock the channel for after the game is finished",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role to lock the channel for after the game is finished",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "log-channel",
				Description: "Set the channel where all bot activities will be logged",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel where logs will be sent",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "auto-reaction-hints",
				Description: "Enable or disable automatic reaction hints for all games",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionBool{
						Name:        "enabled",
						Description: "Whether to enable automatic reaction hints (⬆️/⬇️) for wrong guesses",
						Required:    false,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "show",
				Description: "Show all current server configuration settings",
			},
		},
	}
}

func (c *SetupCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName
	switch subcommand {
	case "prefix":
		return c.handlePrefix(ctx, event, data)
	case "manager":
		return c.handleManager(ctx, event, data)
	case "dm":
		return c.handleDM(ctx, event, data)
	case "win-role":
		return c.handleWinRole(ctx, event, data)
	case "req-role":
		return c.handleReqRole(ctx, event, data)
	case "lock-role":
		return c.handleLockRole(ctx, event, data)
	case "log-channel":
		return c.handleLogChannel(ctx, event, data)
	case "auto-reaction-hints":
		return c.handleAutoReactionHints(ctx, event, data)
	case "show":
		return c.handleShow(ctx, event, data)
	}
	return nil
}

func (c *SetupCommand) handlePrefix(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	prefix := event.SlashCommandInteractionData().String("prefix")

	// Validate prefix
	if len(prefix) == 0 {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Prefix cannot be empty.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	if len(prefix) > 10 {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Prefix cannot be longer than 10 characters.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	oldPrefix := data.GuildConfig.Prefix
	if oldPrefix == "" {
		oldPrefix = "gg" // Default prefix
	}

	cfg := data.GuildConfig
	cfg.Prefix = prefix
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update prefix. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Prefix Updated",
		fmt.Sprintf("**Previous:** `%s`\n**New:** `%s`\n**Updated by:** <@%s>", oldPrefix, prefix, event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Prefix updated to `%s`", prefix),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleManager(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	role, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("Given role (%s) is not there in server.", role.Mention()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	var oldManagerRole string
	if data.GuildConfig.BotManager != "" {
		if oldRole, exists := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.BotManager)); exists {
			oldManagerRole = oldRole.Mention()
		} else {
			oldManagerRole = fmt.Sprintf("<@&%s> (deleted)", data.GuildConfig.BotManager)
		}
	} else {
		oldManagerRole = "Not configured"
	}

	cfg := data.GuildConfig
	cfg.BotManager = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return err
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Manager Role Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldManagerRole, role.Mention(), event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Manager role updated to %s.\n> *Make sure to assign the role to the users who are allowed to manage the bot.*", role.Mention()),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}

func (c *SetupCommand) handleDM(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	enabled := event.SlashCommandInteractionData().Bool("enabled")

	// Store old value for logging
	oldStatus := "disabled"
	if data.GuildConfig.DM {
		oldStatus = "enabled"
	}

	newStatus := "disabled"
	if enabled {
		newStatus = "enabled"
	}

	cfg := data.GuildConfig
	cfg.DM = enabled
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update DM settings. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 DM Settings Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldStatus, newStatus, event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("DMs have been **%s**", newStatus),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleWinRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("Given role (%s) is not there in server.", role.Mention()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	var oldWinRole string
	if data.GuildConfig.WinRole != "" {
		if oldRole, exists := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.WinRole)); exists {
			oldWinRole = oldRole.Mention()
		} else {
			oldWinRole = fmt.Sprintf("<@&%s> (deleted)", data.GuildConfig.WinRole)
		}
	} else {
		oldWinRole = "Not configured"
	}

	cfg := data.GuildConfig
	cfg.WinRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update win role. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Win Role Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldWinRole, role.Mention(), event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Win role updated to %s", role.Mention()),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleReqRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("Given role (%s) is not there in server.", role.Mention()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	var oldReqRole string
	if data.GuildConfig.ReqRole != "" {
		if oldRole, exists := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.ReqRole)); exists {
			oldReqRole = oldRole.Mention()
		} else {
			oldReqRole = fmt.Sprintf("<@&%s> (deleted)", data.GuildConfig.ReqRole)
		}
	} else {
		oldReqRole = "Not configured"
	}

	cfg := data.GuildConfig
	cfg.ReqRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update required role. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Required Role Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldReqRole, role.Mention(), event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Required role updated to %s.\n> *Make sure to assign the role to the users who are allowed to play the game.*", role.Mention()),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleLockRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("Given role (%s) is not there in server.", role.Mention()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	var oldLockRole string
	if data.GuildConfig.LockRole != "" {
		if oldRole, exists := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(data.GuildConfig.LockRole)); exists {
			oldLockRole = oldRole.Mention()
		} else {
			oldLockRole = fmt.Sprintf("<@&%s> (deleted)", data.GuildConfig.LockRole)
		}
	} else {
		oldLockRole = "Not configured"
	}

	cfg := data.GuildConfig
	cfg.LockRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update lock role. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Lock Role Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldLockRole, role.Mention(), event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Lock role updated to %s", role.Mention()),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleLogChannel(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")

	// Validate channel exists in server and bot can access it
	guildChannel, ok := event.Client().Caches().Channel(channel.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     fmt.Sprintf("Given channel (<#%s>) is not accessible.", channel.ID.String()),
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Check if channel is a text channel
	if guildChannel.Type() != discord.ChannelTypeGuildText {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Log channel must be a text channel.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Store old value for logging
	var oldLogChannel string
	if data.GuildConfig.LogChannel != "" {
		oldLogChannel = fmt.Sprintf("<#%s>", data.GuildConfig.LogChannel)
	} else {
		oldLogChannel = "Not configured"
	}

	cfg := data.GuildConfig
	cfg.LogChannel = channel.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		// Note: Can't log this failure to the log channel since the update failed
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update log channel. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change to the new log channel with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Log Channel Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** <#%s>\n**Updated by:** <@%s>\n\nAll bot activities will now be logged to this channel.", oldLogChannel, channel.ID.String(), event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Log channel updated to <#%s>", channel.ID.String()),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleAutoReactionHints(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	enabled := event.SlashCommandInteractionData().Bool("enabled")

	// Store old value for logging
	oldStatus := "disabled"
	if data.GuildConfig.AutoReactionHints {
		oldStatus = "enabled"
	}

	newStatus := "disabled"
	if enabled {
		newStatus = "enabled"
	}

	cfg := data.GuildConfig
	cfg.AutoReactionHints = enabled
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			Content:     "Failed to update auto reaction hints settings. Please try again later.",
			Emoji:       utils.EmojiError,
			IsEphemeral: true,
		})
	}

	// Log the configuration change with before/after values
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel,
		"🔧 Auto Reaction Hints Updated",
		fmt.Sprintf("**Previous:** %s\n**New:** %s\n**Updated by:** <@%s>", oldStatus, newStatus, event.User().ID.String()),
		"Configuration Change")

	return utils.EventReply(event, utils.MessageRequest{
		Content:     fmt.Sprintf("Auto reaction hints have been **%s**\n> *This applies to all new games unless overridden with the /start command*", newStatus),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: false,
	})
}

func (c *SetupCommand) handleShow(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	cfg := data.GuildConfig
	var content strings.Builder

	content.WriteString("🔧 **Server Configuration**\n\n")

	// Basic settings
	content.WriteString("**📝 Basic Settings:**\n")
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "gg" // Default prefix
	}
	content.WriteString(fmt.Sprintf("• Prefix: `%s`\n", prefix))
	content.WriteString(fmt.Sprintf("• Premium: %s\n", formatBooleanSetting(cfg.Premium)))
	content.WriteString(fmt.Sprintf("• DMs to Winners: %s\n", formatBooleanSetting(cfg.DM)))
	content.WriteString(fmt.Sprintf("• Auto Reaction Hints: %s\n\n", formatBooleanSetting(cfg.AutoReactionHints)))

	// Role settings
	content.WriteString("**👥 Role Settings:**\n")
	content.WriteString(fmt.Sprintf("• Manager Role: %s\n", formatRoleSetting(event, cfg.BotManager)))
	content.WriteString(fmt.Sprintf("• Win Role: %s\n", formatRoleSetting(event, cfg.WinRole)))
	content.WriteString(fmt.Sprintf("• Required Role: %s\n", formatRoleSetting(event, cfg.ReqRole)))
	content.WriteString(fmt.Sprintf("• Lock Role: %s\n\n", formatRoleSetting(event, cfg.LockRole)))

	// Channel settings
	content.WriteString("**📺 Channel Settings:**\n")
	content.WriteString(fmt.Sprintf("• Log Channel: %s", formatChannelSetting(cfg.LogChannel)))

	return utils.EventReply(event, utils.MessageRequest{
		Content:     content.String(),
		Emoji:       utils.EmojiSuccess,
		IsEphemeral: true,
	})
}

func formatBooleanSetting(value bool) string {
	if value {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

func formatRoleSetting(event *events.ApplicationCommandInteractionCreate, roleID string) string {
	if roleID == "" {
		return "❌ Not configured"
	}
	if role, ok := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(roleID)); ok {
		return fmt.Sprintf("✅ %s", role.Mention())
	}
	return fmt.Sprintf("⚠️ <@&%s> (deleted)", roleID)
}

func formatChannelSetting(channelID string) string {
	if channelID == "" {
		return "❌ Not configured"
	}
	return fmt.Sprintf("✅ <#%s>", channelID)
}
